package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
	DBTypeDefault            = MONGODB
	DBConnectionDefault      = "mongodb://mongo:27017"
	DBNameDefault            = "fxtract_db"
	DBTimeoutDefault         = 30
	RestfulEPDefault         = ":8000"
	RestfulTLSEPDefault      = ":8080"
	AMQPMessageBrokerDefault = "amqp://guest:guest@rabbitmq"
	RabbitHostDefault        = "rabbitmq"
	RabbitPortDefault        = "5672"
	RabbitUserDefault        = "guest"
	RabbitPasswordDefault    = "guest"
	RedisHostDefault         = "redis"
	RedisPortDefault         = "6379"
)

// ServiceConfig -
type ServiceConfig struct {
	DatabaseType            DBTYPE `json:"database_type"`       //dbType
	DatabaseConnection      string `json:"database_connection"` // dbURL
	DatabaseName            string `json:"database_name"`       // dbname e.g fxtract_db
	DatabaseTimeout         int    `json:"database_timeout"`    // dbtimeout
	RestfulEndPoint         string `json:"restful_endpoint"`    // service connection end point
	RestfulTLSEndPoint      string `json:"restful_tlsendpoint"`
	AMQPMessageBroker       string `json:"amqp_message_broker"`
	RabbitHost              string `json:"rabbit_host"`
	RabbitPort              string `json:"rabbit_port"`
	RabbitUser              string `json:"rabbit_user"`
	RabbitPassword          string `json:"rabbit_password"`
	RedisHost               string `json:"redis_host"`
	RedisPort               string `json:"redis_port"`
	SendGridApiKey          string
	MailVerifCodeExpiration int64 // in hours
	PassResetCodeExpiration int64 // in minutes
	MailVerifTemplateID     string
	PassResetTemplateID     string
}

// ExtractConfiguration - extracts all database configurations from a file
func ExtractConfiguration(filename string) (ServiceConfig, error) {
	MailVerifCodeExpiration, _ := strconv.ParseInt(os.Getenv("MAIL_VERIFICATION_CODE_EXPIRATION"), 10, 64)
	PassResetCodeExpiration, _ := strconv.ParseInt(os.Getenv("PASSWORD_RESET_CODE_EXPIRATION"), 10, 64)
	MailVerifTemplateID := os.Getenv("MAIL_VERIFICATION_TEMPLATE_ID")
	PassResetTemplateID := os.Getenv("PASSWORD_RESET_TEMPLATE_ID")
	SendGridApiKey := os.Getenv("SENDGRID_API_KEY")

	config := ServiceConfig{
		DBTypeDefault,
		DBConnectionDefault,
		DBNameDefault,
		DBTimeoutDefault,
		RestfulEPDefault,
		RestfulTLSEPDefault,
		AMQPMessageBrokerDefault,
		RabbitHostDefault,
		RabbitPortDefault,
		RabbitUserDefault,
		RabbitPasswordDefault,
		RedisHostDefault,
		RedisPortDefault,
		SendGridApiKey,
		MailVerifCodeExpiration,
		PassResetCodeExpiration,
		MailVerifTemplateID,
		PassResetTemplateID,
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Config file not found. Continuing with default values.")
		fmt.Printf("Using: %v %v %v %v %v\n", DBTypeDefault, DBConnectionDefault, DBNameDefault, DBTimeoutDefault, RestfulEPDefault)

		return config, err
	}

	err = json.NewDecoder(file).Decode(&config)

	fmt.Printf("Using: %v : %v %v %v %v redis: %v %v\n", filename, config.DatabaseType, config.DatabaseConnection, config.DatabaseName, config.DatabaseTimeout, config.RestfulEndPoint,
		config.RedisHost+":"+config.RedisPort)

	return config, err
}
