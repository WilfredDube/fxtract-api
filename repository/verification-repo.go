package repository

import (
	"context"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserRepository -
type VerificationRepository interface {
	// Create a new project
	Create(verification *entity.Verification) (*entity.Verification, error)

	Update(verification entity.Verification) (*entity.Verification, error)

	Find(email string, verificationType entity.VerificationDataType) (*entity.Verification, error)

	// Find all projects
	FindAll() ([]entity.Verification, error)

	// Delete a project
	Delete(id string) (int64, error)
}

const (
	verificationCollectionName string = "verifications"
)

// verificationRepoConnection -
type verificationRepoConnection struct {
	connection configuration.MongoRepository
}

// NewUserRepository -
func NewVerificationRepository(db configuration.MongoRepository) VerificationRepository {
	return &verificationRepoConnection{
		connection: db,
	}
}

// Create -
func (r *verificationRepoConnection) Create(verification *entity.Verification) (*entity.Verification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(verificationCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"email":      verification.Email,
			"code":       verification.Code,
			"type":       verification.Type,
			"expires_at": verification.ExpiresAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.Verification.Create")
	}

	return verification, nil
}

// Update -
func (r *verificationRepoConnection) Update(verification entity.Verification) (*entity.Verification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(verificationCollectionName)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": verification.ID},
		bson.D{
			{"$set", bson.M{
				"email":      verification.Email,
				"code":       verification.Code,
				"type":       verification.Type,
				"expires_at": verification.ExpiresAt,
			}}},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.Verification.Update")
	}

	return &verification, nil
}

func (r *verificationRepoConnection) Find(email string, verificationType entity.VerificationDataType) (*entity.Verification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	verification := &entity.Verification{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(verificationCollectionName)

	filter := bson.M{"email": email, "type": verificationType}
	err := collection.FindOne(ctx, filter).Decode(&verification)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Verification not found"), "repository.Verification.Find")
		}
		return nil, errors.Wrap(err, "repository.Verification.Find")
	}

	return verification, nil
}

// FindAll -
func (r *verificationRepoConnection) FindAll() ([]entity.Verification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	users := &[]entity.Verification{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(verificationCollectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Verification not found"), "repository.Verification.FindAll")
		}
		return nil, errors.Wrap(err, "repository.Verification.FindAll")
	}

	cursor.All(ctx, users)
	defer cursor.Close(ctx)

	return *users, nil
}

// Delete -
func (r *verificationRepoConnection) Delete(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(verificationCollectionName)

	pid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return 0, errors.Wrap(errors.New("Verification {id} incorrect"), "repository.Verification.Delete")
		}
		return 0, errors.Wrap(err, "repository.Verification.Delete")
	}

	filter := bson.M{"_id": pid}
	cursor, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.Verification.Delete")
		}
		return 0, errors.Wrap(err, "repository.Verification.Delete")
	}

	return cursor.DeletedCount, nil
}
