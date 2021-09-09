package repository

import (
	"context"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository -
type UserRepository interface {
	// Create a new project
	Create(user *entity.User) (*entity.User, error)

	Update(user entity.User) (*entity.User, error)

	// Find a project by its id
	Profile(id string) (*entity.User, error)

	// Find all projects
	FindAll() ([]entity.User, error)

	// Delete a project
	Delete(id string) (int64, error)

	VerifyCredential(email string, password string) (entity.User, error)

	IsDuplicateEmail(email string) error

	FindByEmail(email string) *entity.User
}

const (
	userCollectionName string = "users"
)

// userRepoConnection -
type userRepoConnection struct {
	connection configuration.MongoRepository
}

// NewUserRepository -
func NewUserRepository(db configuration.MongoRepository) UserRepository {
	return &userRepoConnection{
		connection: db,
	}
}

// Create -
func (r *userRepoConnection) Create(user *entity.User) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"_id":        user.ID,
			"firstname":  user.Firstname,
			"lastname":   user.Lastname,
			"email":      user.Email,
			"password":   user.Password,
			"role":       user.UserRole,
			"created_at": user.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.User.Create")
	}

	return user, nil
}

// Update -
func (r *userRepoConnection) Update(user entity.User) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.D{
			{"$set", bson.M{
				"_id":        user.ID,
				"firstname":  user.Firstname,
				"lastname":   user.Lastname,
				"email":      user.Email,
				"password":   user.Password,
				"role":       user.UserRole,
				"created_at": user.CreatedAt,
			}}},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.User.Update")
	}

	return &user, nil
}

// Find -
func (r *userRepoConnection) Profile(id string) (*entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	user := &entity.User{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)

	pid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("User {id} incorrect"), "repository.User.Find")
		}
		return nil, errors.Wrap(err, "repository.User.Find")
	}

	filter := bson.M{"_id": pid}
	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("User not found"), "repository.User.Find")
		}
		return nil, errors.Wrap(err, "repository.User.Find")
	}

	return user, nil
}

// FindAll -
func (r *userRepoConnection) FindAll() ([]entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	users := &[]entity.User{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Users not found"), "repository.User.FindAll")
		}
		return nil, errors.Wrap(err, "repository.User.FindAll")
	}

	cursor.All(ctx, users)
	defer cursor.Close(ctx)

	return *users, nil
}

// Delete -
func (r *userRepoConnection) Delete(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)

	pid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return 0, errors.Wrap(errors.New("User {id} incorrect"), "repository.User.Find")
		}
		return 0, errors.Wrap(err, "repository.User.Find")
	}

	filter := bson.M{"_id": pid}
	cursor, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.User.Delete")
		}
		return 0, errors.Wrap(err, "repository.User.Delete")
	}

	return cursor.DeletedCount, nil
}

func (r *userRepoConnection) VerifyCredential(email string, password string) (entity.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()
	var user entity.User

	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)

	err := collection.FindOne(ctx, bson.D{primitive.E{Key: "email", Value: email}}).Decode(&user)
	if err != nil {
		return entity.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (r *userRepoConnection) IsDuplicateEmail(email string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()
	var user entity.User

	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)

	err := collection.FindOne(ctx, bson.D{primitive.E{Key: "email", Value: email}}).Decode(&user)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepoConnection) FindByEmail(email string) *entity.User {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	user := &entity.User{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(userCollectionName)

	filter := bson.M{"email": email}
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil
		}
		return nil
	}

	return user
}
