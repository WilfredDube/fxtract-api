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

// CADFileRepository -
type CADFileRepository interface {
	// Create a new project
	Create(project *entity.CADFile) (*entity.CADFile, error)

	// Find a project by its id
	Find(id string) (*entity.CADFile, error)

	// Find all projects
	FindAll() ([]entity.CADFile, error)

	// Delete a project
	Delete(id string) (int64, error)
}

const (
	cadFileCollectionName string = "projects"
)

// userRepoConnection -
type cadFileRepoConnection struct {
	connection configuration.MongoRepository
}

// NewCadFileRepository -
func NewCadFileRepository(db configuration.MongoRepository) CADFileRepository {
	return &cadFileRepoConnection{
		connection: db,
	}
}

func (r *cadFileRepoConnection) Create(cadfile *entity.CADFile) (*entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"_id":           cadfile.ID,
			"filename":      cadfile.FileName,
			"step_url":      cadfile.StepURL,
			"obj_url":       cadfile.ObjpURL,
			"material_id":   cadfile.Material,
			"filesize":      cadfile.Filesize,
			"feature_props": cadfile.FeatureProps,
			"created_at":    cadfile.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.CADFile.Create")
	}

	return cadfile, nil
}

func (r *cadFileRepoConnection) Find(id string) (*entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	cadfile := &entity.CADFile{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	cid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("CADFile {id} incorrect"), "repository.CADFile.Find")
		}
		return nil, errors.Wrap(err, "repository.CADFile.Find")
	}

	filter := bson.M{"_id": cid}
	err = collection.FindOne(ctx, filter).Decode(&cadfile)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("CADFile not found"), "repository.CADFile.Find")
		}
		return nil, errors.Wrap(err, "repository.CADFile.Find")
	}

	return cadfile, nil
}

func (r *cadFileRepoConnection) FindAll() ([]entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	cadfiles := &[]entity.CADFile{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Projects not found"), "repository.CADFile.FindAll")
		}
		return nil, errors.Wrap(err, "repository.CADFile.FindAll")
	}

	cursor.All(ctx, cadfiles)
	defer cursor.Close(ctx)

	return *cadfiles, nil
}

func (r *cadFileRepoConnection) Delete(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	cid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return 0, errors.Wrap(errors.New("CADFile {id} incorrect"), "repository.CADFile.Find")
		}
		return 0, errors.Wrap(err, "repository.CADFile.Find")
	}

	filter := bson.M{"_id": cid}
	cursor, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.CADFile.Delete")
		}
		return 0, errors.Wrap(err, "repository.CADFile.Delete")
	}

	return cursor.DeletedCount, nil
}
