package repository

import (
	"context"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MaterialRepository -
type MaterialRepository interface {
	// Create a new material
	Create(material *entity.Material) (*entity.Material, error)

	// Find a material by its id
	Find(id string) (*entity.Material, error)

	// Find all materials
	FindAll(ownerID string) ([]entity.Material, error)

	// Delete a material
	Delete(id string) (int64, error)
}

const (
	materialCollectionName string = "materials"
)

// userRepoConnection -
type materialRepoConnection struct {
	connection configuration.MongoRepository
}

// NewMaterialRepository -
func NewMaterialRepository(db configuration.MongoRepository) MaterialRepository {
	return &materialRepoConnection{
		connection: db,
	}
}

func (r *materialRepoConnection) Create(material *entity.Material) (*entity.Material, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(materialCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"name":             material.Name,
			"tensile_strength": material.TensileStrength,
			"k_factor":         material.KFactor,
			"created_at":       material.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.Material.Create")
	}

	return material, nil
}

func (r *materialRepoConnection) Find(id string) (*entity.Material, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	material := &entity.Material{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(materialCollectionName)

	filter := bson.M{"name": id}
	err := collection.FindOne(ctx, filter).Decode(&material)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Material not found"), "repository.Material.Find")
		}
		return nil, errors.Wrap(err, "repository.Material.Find")
	}

	return material, nil
}

func (r *materialRepoConnection) FindAll(ownerID string) ([]entity.Material, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	materials := &[]entity.Material{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(materialCollectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Materials not found"), "repository.Material.FindAll")
		}
		return nil, errors.Wrap(err, "repository.Material.FindAll")
	}

	cursor.All(ctx, materials)
	defer cursor.Close(ctx)

	return *materials, nil
}

func (r *materialRepoConnection) Delete(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(materialCollectionName)

	filter := bson.M{"name": id}
	cursor, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.Material.Delete")
		}
		return 0, errors.Wrap(err, "repository.Material.Delete")
	}

	return cursor.DeletedCount, nil
}
