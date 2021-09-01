package repository

import (
	"context"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ToolRepository -
type ToolRepository interface {
	// Create a new tool
	Create(tool *entity.Tool) (*entity.Tool, error)

	// Find a tool by its id
	Find(id string) (*entity.Tool, error)

	FindByAngle(angle int64) (*entity.Tool, error)

	// Find all tools
	FindAll() ([]entity.Tool, error)

	// Delete a tool
	Delete(id string) (int64, error)
}

const (
	toolCollectionName string = "tools"
)

// userRepoConnection -
type toolRepoConnection struct {
	connection configuration.MongoRepository
}

// NewToolRepository -
func NewToolRepository(db configuration.MongoRepository) ToolRepository {
	return &toolRepoConnection{
		connection: db,
	}
}

func (r *toolRepoConnection) Create(tool *entity.Tool) (*entity.Tool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(toolCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"tool_id":    tool.ToolID,
			"tool_name":  tool.ToolName,
			"angle":      tool.Angle,
			"length":     tool.Length,
			"min_radius": tool.MinRadius,
			"max_radius": tool.MaxRadius,
			"created_at": tool.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.Tool.Create")
	}

	return tool, nil
}

func (r *toolRepoConnection) Find(id string) (*entity.Tool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	tool := &entity.Tool{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(toolCollectionName)

	filter := bson.M{"tool_id": id}
	err := collection.FindOne(ctx, filter).Decode(&tool)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Tool not found"), "repository.Tool.Find")
		}
		return nil, errors.Wrap(err, "repository.Tool.Find")
	}

	return tool, nil
}

func (r *toolRepoConnection) FindByAngle(angle int64) (*entity.Tool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	tool := &entity.Tool{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(toolCollectionName)

	filter := bson.M{"angle": angle}
	err := collection.FindOne(ctx, filter).Decode(&tool)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Tool not found"), "repository.Tool.Find")
		}
		return nil, errors.Wrap(err, "repository.Tool.Find")
	}

	return tool, nil
}

func (r *toolRepoConnection) FindAll() ([]entity.Tool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	tools := &[]entity.Tool{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(toolCollectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Tools not found"), "repository.Tool.FindAll")
		}
		return nil, errors.Wrap(err, "repository.Tool.FindAll")
	}

	cursor.All(ctx, tools)
	defer cursor.Close(ctx)

	return *tools, nil
}

func (r *toolRepoConnection) Delete(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(toolCollectionName)

	filter := bson.M{"tool_id": id}
	cursor, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.Tool.Delete")
		}
		return 0, errors.Wrap(err, "repository.Tool.Delete")
	}

	return cursor.DeletedCount, nil
}
