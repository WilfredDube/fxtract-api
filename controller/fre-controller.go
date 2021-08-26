package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
)

type freController struct {
	userService           service.UserService
	cadFileService        service.CadFileService
	processingPlanService service.ProcessingPlanService
	jwtService            service.JWTService
	config                configuration.ServiceConfig
}

// FREController -
type FREController interface {
	ProcessCADFile(w http.ResponseWriter, r *http.Request)
	ExtractBendFeatures(UserID string, cadFile *entity.CADFile)
	GenerateProcessingPlan(UserID string, cadFile *entity.CADFile)
	BatchProcessCADFiles(w http.ResponseWriter, r *http.Request)
}

// NewFREController -
func NewFREController(configuration configuration.ServiceConfig, cadService service.CadFileService, pPlanService service.ProcessingPlanService,
	uService service.UserService, jwtService service.JWTService) FREController {
	return &freController{
		userService:           uService,
		cadFileService:        cadService,
		processingPlanService: pPlanService,
		jwtService:            jwtService,
		config:                configuration,
	}
}

func (c *freController) ExtractBendFeatures(UserID string, cadFile *entity.CADFile) {
	request := &entity.FRERequest{}

	request.CADFileID = cadFile.ID.Hex()
	request.UserID = UserID
	request.URL = cadFile.StepURL

	message, err := json.Marshal(request)

	log.Printf("*********************************************************************")
	fmt.Println("Publisher received message: " + string(message))

	conn, err := amqp.Dial("amqp://" + c.config.RabbitUser + ":" + c.config.RabbitPassword + "@" + c.config.RabbitHost + ":" + c.config.RabbitPort + "/")
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"FREC", // name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})

	if err != nil {
		log.Fatalf("%s: %s", "Failed to publish a message", err)
	}

	fmt.Println("Publish successful!")
	log.Printf("*********************************************************************")
}

func (c *freController) GenerateProcessingPlan(UserID string, cadFile *entity.CADFile) {
	request := &entity.PPRequest{}

	request.CADFileID = cadFile.ID.Hex()
	request.UserID = UserID
	request.BendCount = int64(cadFile.FeatureProps.BendCount)
	request.SerializedData = cadFile.FeatureProps.SerialData

	message, err := json.Marshal(request)

	log.Printf("*********************************************************************")
	fmt.Println("Publisher received message: " + string(message))

	conn, err := amqp.Dial("amqp://" + c.config.RabbitUser + ":" + c.config.RabbitPassword + "@" + c.config.RabbitHost + ":" + c.config.RabbitPort + "/")
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"PPLAN", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})

	if err != nil {
		log.Fatalf("%s: %s", "Failed to publish a message", err)
	}

	fmt.Println("Publishing successful!")
	log.Printf("*********************************************************************")
}

// ProcessCADFile -
func (c *freController) ProcessCADFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if token == nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["user_id"].(string)

		if _, err := c.userService.Profile(id); err != nil {
			response := helper.BuildErrorResponse("Invalid token", "User does not exist", helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		params := mux.Vars(r)
		cid := params["id"]

		cadFile, err := c.cadFileService.Find(cid)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "CAD file not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		if cadFile.FeatureProps.ProcessLevel == 0 {
			res := helper.BuildResponse(true, "Feature recognition started", &helper.EmptyObj{})
			c.ExtractBendFeatures(id, cadFile)
			json.NewEncoder(w).Encode(res)
			return
		} else if cadFile.FeatureProps.ProcessLevel == 1 {
			res := helper.BuildResponse(true, "Process planning started", &helper.EmptyObj{})
			c.GenerateProcessingPlan(id, cadFile)
			json.NewEncoder(w).Encode(res)
			return
		}

		processingPlan, err := c.processingPlanService.Find(cadFile.ID.Hex())
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "Processing plan not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "CAD file has been fully processed", processingPlan)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}

// BatchProcessCADFiles -
func (c *freController) BatchProcessCADFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if token == nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["user_id"].(string)

		if _, err := c.userService.Profile(id); err != nil {
			response := helper.BuildErrorResponse("Invalid token", "User does not exist", helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		params := mux.Vars(r)
		pid := params["pid"]

		cadFiles, err := c.cadFileService.FindAll(pid)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "CAD file not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		var freNum = 0
		var ppNum = 0
		for _, cadFile := range cadFiles {
			if cadFile.FeatureProps.ProcessLevel == 0 {
				res := helper.BuildResponse(true, fmt.Sprintf("Feature recognition started: %v", cadFile.FileName), &helper.EmptyObj{})
				c.ExtractBendFeatures(id, &cadFile)
				json.NewEncoder(w).Encode(res)
				freNum++
				// return
			} else if cadFile.FeatureProps.ProcessLevel == 1 {
				res := helper.BuildResponse(true, fmt.Sprintf("Process planning started: %v", cadFile.FileName), &helper.EmptyObj{})
				c.GenerateProcessingPlan(id, &cadFile)
				json.NewEncoder(w).Encode(res)
				ppNum++
				// return
			}
		}

		if ppNum > 0 || freNum > 0 {
			return
		}

		res := helper.BuildResponse(true, "Process planning complete: all files processed.", &helper.EmptyObj{})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}
