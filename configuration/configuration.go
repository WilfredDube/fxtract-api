package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

// DBTYPE -
type DBTYPE string

// Const -
const (
	MONGODB  DBTYPE = "mongodb"
	DYNAMODB DBTYPE = "dynamodb"
	MYSQL    DBTYPE = "mysql"
	POSTGRES DBTYPE = "postgres"
	REDIS    DBTYPE = "redis"
)

// var -
var (
	DBTypeDefault       = MONGODB
	DBConnectionDefault = "mongodb://localhost:27017"
	DBNameDefault       = "fxtract_db"
	DBTimeoutDefault    = 30
	RestfulEPDefault    = ":8000"
	RestfulTLSEPDefault = ":8080"
)

// ServiceConfig -
type ServiceConfig struct {
	DatabaseType       DBTYPE `json:"database_type"`       //dbType
	DatabaseConnection string `json:"database_connection"` // dbURL
	DatabaseName       string `json:"database_name"`       // dbname e.g fxtract_db
	DatabaseTimeout    int    `json:"database_timeout"`    // dbtimeout
	RestfulEndPoint    string `json:"restful_endpoint"`    // service connection end point
	RestfulTLSEndPoint string `json:"restful_tlsendpoint"`
}

// ExtractConfiguration - extracts all database configurations from a file
func ExtractConfiguration(filename string) (ServiceConfig, error) {
	config := ServiceConfig{
		DBTypeDefault,
		DBConnectionDefault,
		DBNameDefault,
		DBTimeoutDefault,
		RestfulEPDefault,
		RestfulTLSEPDefault,
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Config file not found. Continuing with default values.")
		fmt.Printf("Using: %v %v %v %v %v\n", DBTypeDefault, DBConnectionDefault, DBNameDefault, DBTimeoutDefault, RestfulEPDefault)

		return config, err
	}

	err = json.NewDecoder(file).Decode(&config)

	fmt.Printf("Using: %v : %v %v %v %v %v\n", filename, config.DatabaseType, config.DatabaseConnection, config.DatabaseName, config.DatabaseTimeout, config.RestfulEndPoint)

	return config, err
}
