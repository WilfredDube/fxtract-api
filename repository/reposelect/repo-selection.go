package repository

import (
	"log"

	"github.com/WilfredDube/fxtract-backend/configuration"
)

// NewPersistenceLayer -
func NewPersistenceLayer(config configuration.ServiceConfig) *configuration.MongoRepository {
	switch config.DatabaseType {
	case "mongodb":
		repo, err := configuration.NewMongoRepository(config.DatabaseConnection, config.DatabaseName, config.DatabaseTimeout)
		if err != nil {
			log.Fatal(err)
		}

		return repo
	}

	return nil
}
