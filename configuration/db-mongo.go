package configuration

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoRepository -
type MongoRepository struct {
	Client   *mongo.Client
	Database string
	Timeout  time.Duration
}

func newMongoClient(mongoURL string, mongoTimeout int) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mongoTimeout)*time.Second)
	defer cancel()

	Client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		return Client, err
	}

	err = Client.Ping(ctx, readpref.Primary())
	if err != nil {
		return Client, err
	}

	return Client, nil
}

// NewMongoRepository -
func NewMongoRepository(mongoURL, mongoDB string, mongoTimeout int) (*MongoRepository, error) {
	repo := &MongoRepository{
		Timeout:  time.Duration(mongoTimeout) * time.Second,
		Database: mongoDB,
	}

	Client, err := newMongoClient(mongoURL, mongoTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "repository.NewMongoRepository")
	}

	repo.Client = Client

	return repo, nil
}
