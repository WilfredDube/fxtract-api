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

// ProjectRepository -
type ProjectRepository interface {
	// Create a new project
	Create(project *entity.Project) (*entity.Project, error)

	// Find a project by its id
	Find(id string) (*entity.Project, error)

	FindByName(name string) (*entity.Project, error)

	IsDuplicate(name string, OwnerID primitive.ObjectID) bool

	// Find all projects
	FindAll(ownerID string) ([]entity.Project, error)

	// Delete a project
	Delete(id string) (int64, error)
}

const (
	projectCollectionName string = "projects"
)

// userRepoConnection -
type projectRepoConnection struct {
	connection configuration.MongoRepository
}

// NewProjectRepository -
func NewProjectRepository(db configuration.MongoRepository) ProjectRepository {
	return &projectRepoConnection{
		connection: db,
	}
}

func (r *projectRepoConnection) Create(project *entity.Project) (*entity.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(projectCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"_id":         project.ID,
			"title":       project.Title,
			"description": project.Description,
			"owner_id":    project.OwnerID,
			"created_at":  project.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.Project.Create")
	}

	return project, nil
}

func (r *projectRepoConnection) Find(id string) (*entity.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	project := &entity.Project{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(projectCollectionName)

	pid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Project {id} incorrect"), "repository.Project.Find")
		}
		return nil, errors.Wrap(err, "repository.Project.Find")
	}

	filter := bson.M{"_id": pid}
	err = collection.FindOne(ctx, filter).Decode(&project)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Project not found"), "repository.Project.Find")
		}
		return nil, errors.Wrap(err, "repository.Project.Find")
	}

	return project, nil
}

func (r *projectRepoConnection) FindByName(name string) (*entity.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	project := &entity.Project{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(projectCollectionName)

	filter := bson.M{"title": name}
	err := collection.FindOne(ctx, filter).Decode(&project)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Project not found"), "repository.Project.Find")
		}
		return nil, errors.Wrap(err, "repository.Project.Find")
	}

	return project, nil
}

func (r *projectRepoConnection) IsDuplicate(name string, OwnerID primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	project := &entity.Project{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(projectCollectionName)

	filter := bson.M{"title": name, "owner_id": OwnerID}
	err := collection.FindOne(ctx, filter).Decode(&project)
	if err != nil {
		return false
	}

	return true
}

func (r *projectRepoConnection) FindAll(ownerID string) ([]entity.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	projects := &[]entity.Project{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(projectCollectionName)

	id, err := primitive.ObjectIDFromHex(ownerID)
	if err != nil {
		return nil, errors.Wrap(errors.New("Incorrect user id"), "repository.Project.FindAll")
	}

	cursor, err := collection.Find(ctx, bson.M{"owner_id": id})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Projects not found"), "repository.Project.FindAll")
		}
		return nil, errors.Wrap(err, "repository.Project.FindAll")
	}

	cursor.All(ctx, projects)
	defer cursor.Close(ctx)

	return *projects, nil
}

func (r *projectRepoConnection) Delete(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(projectCollectionName)

	pid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return 0, errors.Wrap(errors.New("Project {id} incorrect"), "repository.Project.Find")
		}
		return 0, errors.Wrap(err, "repository.Project.Find")
	}

	filter := bson.M{"_id": pid}
	cursor, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.Project.Delete")
		}
		return 0, errors.Wrap(err, "repository.Project.Delete")
	}

	return cursor.DeletedCount, nil
}
