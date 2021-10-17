package repository

import (
	"context"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskRepository -
type TaskRepository interface {
	// Create a new task
	Create(task *entity.Task) (*entity.Task, error)

	Update(task *entity.Task) (*entity.Task, error)

	// Find a task by its id
	Find(id string) (*entity.Task, error)

	FindByUserID(id string) (*entity.Task, error)

	// Find all tasks
	FindAll() ([]entity.Task, error)
}

const (
	taskCollectionName string = "tasks"
)

// userRepoConnection -
type taskRepoConnection struct {
	connection configuration.MongoRepository
}

// NewTaskRepository -
func NewTaskRepository(db configuration.MongoRepository) TaskRepository {
	return &taskRepoConnection{
		connection: db,
	}
}

func (r *taskRepoConnection) Create(task *entity.Task) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(taskCollectionName)
	_, err := collection.InsertOne(
		ctx,
		bson.M{
			"_id":                          task.ID,
			"task_id":                      task.TaskID,
			"user_id":                      task.UserID,
			"cadfiles":                     task.CADFiles,
			"processed_cadfiles":           task.ProcessedCADFiles,
			"status":                       task.Status,
			"quantity":                     task.Quantity,
			"processing_time":              task.ProcessingTime,
			"estimated_manufacturing_time": task.EstimatedManufacturingTime,
			"total_cost":                   task.TotalCost,
			"created_at":                   task.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.Task.Create")
	}

	return task, nil
}

// Update -
func (r *taskRepoConnection) Update(task *entity.Task) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(taskCollectionName)

	filter := bson.D{{"_id", task.ID}}
	update := bson.D{
		{"$set", bson.M{
			"status":                       task.Status,
			"processing_time":              task.ProcessingTime,
			"processed_cadfiles":           task.ProcessedCADFiles,
			"estimated_manufacturing_time": task.EstimatedManufacturingTime,
		}}}
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, errors.Wrap(err, "repository.Task.Update")
	}

	return task, nil
}

func (r *taskRepoConnection) Find(id string) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	task := &entity.Task{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(taskCollectionName)

	taskID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("User {id} incorrect"), "repository.User.Find")
		}
		return nil, errors.Wrap(err, "repository.User.Find")
	}

	filter := bson.M{"_id": taskID}
	err = collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Task not found"), "repository.Task.Find")
		}
		return nil, errors.Wrap(err, "repository.Task.Find")
	}

	return task, nil
}

func (r *taskRepoConnection) FindByUserID(id string) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	task := &entity.Task{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(taskCollectionName)

	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("User {id} incorrect"), "repository.User.Find")
		}
		return nil, errors.Wrap(err, "repository.User.Find")
	}

	filter := bson.M{"user_id": userID}
	err = collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Task not found"), "repository.Task.FindByUserID")
		}
		return nil, errors.Wrap(err, "repository.Task.FindByUserID")
	}

	return task, nil
}

func (r *taskRepoConnection) FindAll() ([]entity.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	tasks := &[]entity.Task{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(taskCollectionName)

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Tasks not found"), "repository.Task.FindAll")
		}
		return nil, errors.Wrap(err, "repository.Task.FindAll")
	}

	cursor.All(ctx, tasks)
	defer cursor.Close(ctx)

	return *tasks, nil
}
