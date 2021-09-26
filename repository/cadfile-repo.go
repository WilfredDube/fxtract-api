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

	Update(cadFile entity.CADFile) (*entity.CADFile, error)

	// Find a project by its id
	Find(id string) (*entity.CADFile, error)

	// Find all projects
	FindAll(projectID string) ([]entity.CADFile, error)

	FindAllFiles() ([]entity.CADFile, error)

	FindSelected(selectedFiles []string) ([]entity.CADFile, error)

	// Delete a project
	Delete(projectID string) (int64, error)

	CascadeDelete(id string) (int64, error)
}

const (
	cadFileCollectionName string = "cadfiles"
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

func (r *cadFileRepoConnection) Create(cadFile *entity.CADFile) (*entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"_id":           cadFile.ID,
			"project_id":    cadFile.ProjectID,
			"filename":      cadFile.FileName,
			"step_url":      cadFile.StepURL,
			"obj_url":       cadFile.ObjpURL,
			"material_id":   cadFile.Material,
			"filesize":      cadFile.Filesize,
			"feature_props": cadFile.FeatureProps,
			"bend_features": cadFile.BendFeatures,
			"created_at":    cadFile.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.CADFile.Create")
	}

	return cadFile, nil
}

func (r *cadFileRepoConnection) Update(cadFile entity.CADFile) (*entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": cadFile.ID},
		bson.D{
			{"$set", bson.M{
				"_id":           cadFile.ID,
				"project_id":    cadFile.ProjectID,
				"filename":      cadFile.FileName,
				"step_url":      cadFile.StepURL,
				"obj_url":       cadFile.ObjpURL,
				"material_id":   cadFile.Material,
				"filesize":      cadFile.Filesize,
				"feature_props": cadFile.FeatureProps,
				"bend_features": cadFile.BendFeatures,
				"created_at":    cadFile.CreatedAt,
			}}},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.User.Update")
	}

	return &cadFile, nil
}

func (r *cadFileRepoConnection) Find(id string) (*entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	cadFile := &entity.CADFile{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	cid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("CADFile {id} incorrect"), "repository.CADFile.Find")
		}
		return nil, errors.Wrap(err, "repository.CADFile.Find")
	}

	filter := bson.M{"_id": cid}
	err = collection.FindOne(ctx, filter).Decode(&cadFile)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("CADFile not found"), "repository.CADFile.Find")
		}
		return nil, errors.Wrap(err, "repository.CADFile.Find")
	}

	return cadFile, nil
}

func (r *cadFileRepoConnection) FindAll(projectID string) ([]entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(projectID)

	cadfiles := &[]entity.CADFile{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	cursor, err := collection.Find(ctx, bson.M{"project_id": id})
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

func (r *cadFileRepoConnection) FindSelected(selectedFiles []string) ([]entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	var ids []primitive.ObjectID
	for _, selectedID := range selectedFiles {
		id, _ := primitive.ObjectIDFromHex(selectedID)
		ids = append(ids, id)
	}

	cadfiles := &[]entity.CADFile{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	cursor, err := collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Cadfiles not found"), "repository.CADFile.FindSelected")
		}
		return nil, errors.Wrap(err, "repository.CADFile.FindSelected")
	}

	cursor.All(ctx, cadfiles)
	defer cursor.Close(ctx)

	return *cadfiles, nil
}

func (r *cadFileRepoConnection) FindAllFiles() ([]entity.CADFile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	cadfiles := &[]entity.CADFile{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("CAD files not found"), "repository.CADFile.FindAll")
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

func (r *cadFileRepoConnection) CascadeDelete(projectID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(cadFileCollectionName)

	pid, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return 0, errors.Wrap(errors.New("Project {id} incorrect"), "repository.CADFile.Find")
		}
		return 0, errors.Wrap(err, "repository.CADFile.Find")
	}

	filter := bson.M{"project_id": pid}
	cursor, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.CADFile.Delete")
		}
		return 0, errors.Wrap(err, "repository.CADFile.Delete")
	}

	return cursor.DeletedCount, nil
}
