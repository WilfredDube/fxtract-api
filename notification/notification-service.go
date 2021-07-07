package notification

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mqResponseController -
type mqResponseController struct {
	userService           service.UserService
	cadFileService        service.CadFileService
	processingPlanService service.ProcessingPlanService
	jwtService            service.JWTService
	config                configuration.ServiceConfig
}

// MQResponseController  -
type MQResponseController interface {
	FeatureRecognitionNotifications(UserID string)
	ProcessingPlanNotifications(UserID string)
}

// NewMQResponseController -
func NewMQResponseController(configuration configuration.ServiceConfig, cadService service.CadFileService, pPlanService service.ProcessingPlanService,
	uService service.UserService, jwtService service.JWTService) MQResponseController {
	return &mqResponseController{
		userService:           uService,
		cadFileService:        cadService,
		processingPlanService: pPlanService,
		jwtService:            jwtService,
		config:                configuration,
	}
}

func (m *mqResponseController) FeatureRecognitionNotifications(name string) {
	conn, err := amqp.Dial("amqp://" + m.config.RabbitUser + ":" + m.config.RabbitPassword + "@" + m.config.RabbitHost + ":" + m.config.RabbitPort + "/")
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}

	q, err := ch.QueueDeclare(
		"FRECRESPONSE", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	fmt.Println("Channel and Queue established: ", name)

	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to register consumer", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			freResponse := &entity.FREResponse{}
			err := json.Unmarshal(d.Body, freResponse)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to unmarshal data: ", err)
			}

			log.Printf("==========================================================")
			fmt.Printf("Received a CAD file with ID: %v\n", freResponse.CADFileID)

			cadFile, err := m.cadFileService.Find(freResponse.CADFileID)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to unmarshal data: ", err)
			}

			cadFile.FeatureProps = freResponse.FeatureProps
			cadFile.BendFeatures = []entity.BendFeature{}
			cadFile.BendFeatures = freResponse.BendFeatures

			_, err = m.cadFileService.Update(*cadFile)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to unmarshal data: ", err)
			}

			d.Ack(false)
			fmt.Println("CAD file features saved successfully!")
			log.Printf("==========================================================")
		}
	}()

	fmt.Println("FRE Notification service started...")
	<-forever
}

func (m *mqResponseController) ProcessingPlanNotifications(name string) {
	conn, err := amqp.Dial("amqp://" + m.config.RabbitUser + ":" + m.config.RabbitPassword + "@" + m.config.RabbitHost + ":" + m.config.RabbitPort + "/")
	if err != nil {
		log.Fatalf("%s: %s", "Failed to connect to RabbitMQ", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s", "Failed to open a channel", err)
	}

	q, err := ch.QueueDeclare(
		"PPRESPONSE", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to declare a queue", err)
	}

	fmt.Println("Channel and Queue established: ", name)

	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		log.Fatalf("%s: %s", "Failed to register consumer", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			ppResponse := &entity.PPResponse{}
			err := json.Unmarshal(d.Body, ppResponse)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to unmarshal data: ", err)
			}

			log.Printf("==========================================================")
			fmt.Printf("Received a Processing plan for CAD file ID: %v\n", ppResponse.ProcessingPlan.CADFileID)

			processingPlan := entity.ProcessingPlan{}
			processingPlan.ID = primitive.NewObjectID()
			processingPlan.CADFileID = ppResponse.ProcessingPlan.CADFileID

			cadFile, err := m.cadFileService.Find(processingPlan.CADFileID.Hex())
			if err != nil {
				log.Fatalf("%s: %s", "Cadfile does not exist ", err)
			}

			processingPlan.Rotations = ppResponse.ProcessingPlan.Rotations
			processingPlan.Flips = ppResponse.ProcessingPlan.Flips
			processingPlan.Tools = ppResponse.ProcessingPlan.Tools
			processingPlan.Modules = ppResponse.ProcessingPlan.Modules
			processingPlan.ProcessingTime = ppResponse.ProcessingPlan.ProcessingTime
			processingPlan.EstimatedManufacturingTime = ppResponse.ProcessingPlan.EstimatedManufacturingTime
			processingPlan.TotalToolDistance = ppResponse.ProcessingPlan.TotalToolDistance
			processingPlan.BendingSequences = ppResponse.ProcessingPlan.BendingSequences
			processingPlan.Quantity = ppResponse.ProcessingPlan.Quantity
			processingPlan.CreatedAt = time.Now().Unix()

			_, err = m.processingPlanService.Create(&processingPlan)
			if err != nil {
				log.Fatalf("%s: %s", "Failed to save processing plan: ", err)
			}

			cadFile.FeatureProps.ProcessLevel = ppResponse.ProcessLevel
			cadFile, err = m.cadFileService.Update(*cadFile)
			if err != nil {
				log.Fatalf("%s: %s", "Cadfile update failed ", err)
			}

			d.Ack(false)
			fmt.Println("CAD file processing plan saved successfully!")
			log.Printf("==========================================================")
		}
	}()

	fmt.Println("Processing Plan Notification service started...")
	<-forever
}
