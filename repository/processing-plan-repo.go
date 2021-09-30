package repository

import (
	"context"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/pkg/errors"
	"github.com/teris-io/shortid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProcessingPlanRepository -
type ProcessingPlanRepository interface {
	// Create a new processingPlan
	Create(processingPlan *entity.ProcessingPlan) (*entity.ProcessingPlan, error)

	Update(processingPlan entity.ProcessingPlan) (*entity.ProcessingPlan, error)

	// Find a processingPlan by its id
	Find(id string) (*entity.ProcessingPlan, error)

	// Find all projects
	FindAll(processingPlanID string) ([]entity.ProcessingPlan, error)

	// Delete a processingPlan
	Delete(processingPlanID string) (int64, error)

	CascadeDelete(id string) (int64, error)
}

const (
	processingPlanCollectionName string = "processing_plans"
)

// userRepoConnection -
type processingPlanRepoConnection struct {
	connection configuration.MongoRepository
}

// NewProcessingPlanRepository -
func NewProcessingPlanRepository(db configuration.MongoRepository) ProcessingPlanRepository {
	return &processingPlanRepoConnection{
		connection: db,
	}
}

func (r *processingPlanRepoConnection) Create(processingPlan *entity.ProcessingPlan) (*entity.ProcessingPlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	if err != nil {
		return nil, errors.Wrap(err, "repository.ProcessingPlan.Create")
	}

	collection := r.connection.Client.Database(r.connection.Database).Collection(processingPlanCollectionName)
	_, err = collection.InsertOne(
		ctx,
		bson.M{
			"_id":                          processingPlan.ID,
			"cadfile_id":                   processingPlan.CADFileID,
			"filename":                     processingPlan.FileName,
			"project_title":                processingPlan.ProjectTitle,
			"engineer":                     processingPlan.Engineer,
			"moderator":                    processingPlan.Moderator,
			"material":                     processingPlan.Material,
			"bending_force":                processingPlan.BendingForce,
			"rotations":                    processingPlan.Rotations,
			"part_no":                      sid.MustGenerate(),
			"flips":                        processingPlan.Flips,
			"tools":                        processingPlan.Tools,
			"modules":                      processingPlan.Modules,
			"quantity":                     processingPlan.Quantity,
			"processing_time":              processingPlan.ProcessingTime,
			"total_tool_distance":          processingPlan.TotalToolDistance,
			"estimated_manufacturing_time": processingPlan.EstimatedManufacturingTime,
			"bend_sequences":               processingPlan.BendingSequences,
			"bend_features":                processingPlan.BendFeatures,
			"created_at":                   processingPlan.CreatedAt,
		},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.ProcessingPlan.Create")
	}

	return processingPlan, nil
}

func (r *processingPlanRepoConnection) Update(processingPlan entity.ProcessingPlan) (*entity.ProcessingPlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(processingPlanCollectionName)
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": processingPlan.ID},
		bson.D{
			{"$set", bson.M{
				"_id":                          processingPlan.ID,
				"cadfile_id":                   processingPlan.CADFileID,
				"filename":                     processingPlan.FileName,
				"project_title":                processingPlan.ProjectTitle,
				"engineer":                     processingPlan.Engineer,
				"moderator":                    processingPlan.Moderator,
				"material":                     processingPlan.Material,
				"bending_force":                processingPlan.BendingForce,
				"part_no":                      processingPlan.PartNo,
				"rotations":                    processingPlan.Rotations,
				"flips":                        processingPlan.Flips,
				"tools":                        processingPlan.Tools,
				"modules":                      processingPlan.Modules,
				"quantity":                     processingPlan.Quantity,
				"processing_time":              processingPlan.ProcessingTime,
				"estimated_manufacturing_time": processingPlan.EstimatedManufacturingTime,
				"total_tool_distance":          processingPlan.TotalToolDistance,
				"bend_sequences":               processingPlan.BendingSequences,
				"bend_features":                processingPlan.BendFeatures,
				"created_at":                   processingPlan.CreatedAt,
			}}},
	)

	if err != nil {
		return nil, errors.Wrap(err, "repository.User.Update")
	}

	return &processingPlan, nil
}

func (r *processingPlanRepoConnection) Find(id string) (*entity.ProcessingPlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	processingPlan := &entity.ProcessingPlan{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(processingPlanCollectionName)

	cid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("ProcessingPlan {id} incorrect"), "repository.ProcessingPlan.Find")
		}
		return nil, errors.Wrap(err, "repository.ProcessingPlan.Find")
	}

	filter := bson.M{"cadfile_id": cid}
	err = collection.FindOne(ctx, filter).Decode(&processingPlan)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("ProcessingPlan not found"), "repository.ProcessingPlan.Find")
		}
		return nil, errors.Wrap(err, "repository.ProcessingPlan.Find")
	}

	return processingPlan, nil
}

func (r *processingPlanRepoConnection) FindAll(processingPlanID string) ([]entity.ProcessingPlan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(processingPlanID)

	cadfiles := &[]entity.ProcessingPlan{}
	collection := r.connection.Client.Database(r.connection.Database).Collection(processingPlanCollectionName)

	cursor, err := collection.Find(ctx, bson.M{"project_id": id})
	if err != nil {
		if err == mongo.ErrNilDocument {
			return nil, errors.Wrap(errors.New("Projects not found"), "repository.ProcessingPlan.FindAll")
		}
		return nil, errors.Wrap(err, "repository.ProcessingPlan.FindAll")
	}

	cursor.All(ctx, cadfiles)
	defer cursor.Close(ctx)

	return *cadfiles, nil
}

func (r *processingPlanRepoConnection) Delete(id string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(processingPlanCollectionName)

	cid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return 0, errors.Wrap(errors.New("ProcessingPlan {id} incorrect"), "repository.ProcessingPlan.Find")
		}
		return 0, errors.Wrap(err, "repository.ProcessingPlan.Find")
	}

	filter := bson.M{"cadfile_id": cid}
	cursor, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.ProcessingPlan.Delete")
		}
		return 0, errors.Wrap(err, "repository.ProcessingPlan.Delete")
	}

	return cursor.DeletedCount, nil
}

func (r *processingPlanRepoConnection) CascadeDelete(processingPlanID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.connection.Timeout)
	defer cancel()

	collection := r.connection.Client.Database(r.connection.Database).Collection(processingPlanCollectionName)

	pid, err := primitive.ObjectIDFromHex(processingPlanID)
	if err != nil {
		if err == mongo.ErrNilDocument {
			return 0, errors.Wrap(errors.New("Project {id} incorrect"), "repository.ProcessingPlan.Find")
		}
		return 0, errors.Wrap(err, "repository.ProcessingPlan.Find")
	}

	filter := bson.M{"project_id": pid}
	cursor, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		if err == mongo.ErrNilCursor {
			return 0, errors.Wrap(errors.New("Delete failed"), "repository.ProcessingPlan.Delete")
		}
		return 0, errors.Wrap(err, "repository.ProcessingPlan.Delete")
	}

	return cursor.DeletedCount, nil
}
