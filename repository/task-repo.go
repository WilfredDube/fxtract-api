package repository

import (
	"context"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TaskRepository -
type TaskRepository interface {
	// Create a new task
	Create(task *entity.Task) (*entity.Task, error)

	// Find a task by its id
	Find(id string) (*entity.Task, error)

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
			"task_id":                      task.TaskID,
			"user_id":                      task.UserID,
			"cadfiles":                     task.CADFiles,
			"process_type":                 task.ProcessType,
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

func (r *taskRepoConnection) Find(id string) (*entity.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	task := &entity.Task{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(taskCollectionName)

	filter := bson.M{"task_id": id}
	err := collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Task not found"), "repository.Task.Find")
		}
		return nil, errors.Wrap(err, "repository.Task.Find")
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
