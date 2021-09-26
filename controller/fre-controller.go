package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/contracts"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/WilfredDube/fxtract-backend/lib/msgqueue"
	persistence "github.com/WilfredDube/fxtract-backend/repository/reposelect"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type freController struct {
	userService           service.UserService
	cadFileService        service.CadFileService
	processingPlanService service.ProcessingPlanService
	jwtService            service.JWTService
	config                configuration.ServiceConfig
	taskService           service.TaskService
	cache                 *redis.Client
	eventEmitter          msgqueue.EventEmitter
}

// FREController -
type FREController interface {
	ProcessCADFile(w http.ResponseWriter, r *http.Request)
	ExtractBendFeatures(UserID string, TaskID string, cadFile *entity.CADFile)
	GenerateProcessingPlan(UserID string, TaskID string, cadFile *entity.CADFile)
	BatchProcessCADFiles(w http.ResponseWriter, r *http.Request)
}

// NewFREController -
func NewFREController(configuration configuration.ServiceConfig, cadService service.CadFileService, pPlanService service.ProcessingPlanService,
	uService service.UserService, jwtService service.JWTService, taskService service.TaskService, cache *redis.Client, eventEmitter msgqueue.EventEmitter) FREController {
	return &freController{
		userService:           uService,
		cadFileService:        cadService,
		processingPlanService: pPlanService,
		jwtService:            jwtService,
		config:                configuration,
		taskService:           taskService,
		cache:                 cache,
		eventEmitter:          eventEmitter,
	}
}

func (c *freController) ExtractBendFeatures(UserID string, TaskID string, cadFile *entity.CADFile) {
	request := &contracts.FeatureRecognitionStarted{
		UserID:    UserID,
		CADFileID: cadFile.ID.Hex(),
		TaskID:    TaskID,
		URL:       cadFile.StepURL,
		EventType: "featureRecognitionStarted",
	}

	c.eventEmitter.Emit(request)

	fmt.Println("Publish successful!")
	log.Printf("*********************************************************************")
}

func (c *freController) GenerateProcessingPlan(UserID string, TaskID string, cadFile *entity.CADFile) {
	request := &contracts.ProcessPlanningStarted{
		CADFileID:      cadFile.ID.Hex(),
		UserID:         UserID,
		TaskID:         TaskID,
		BendCount:      int64(cadFile.FeatureProps.BendCount),
		SerializedData: cadFile.FeatureProps.SerialData,
		EventType:      "processPlanningStarted",
		FRETime:        cadFile.FeatureProps.FRETime,
	}

	c.eventEmitter.Emit(request)
	fmt.Println("Publishing successful!")
	log.Printf("*********************************************************************")
}

// ProcessCADFile -
func (c *freController) ProcessCADFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
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
		cadFileID := params["id"]

		cadFile, err := c.cadFileService.Find(cadFileID)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "CAD file not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		if cadFile.FeatureProps.ProcessLevel == 0 || cadFile.FeatureProps.ProcessLevel == 1 {
			var task entity.Task

			task.TaskID = primitive.NewObjectID()
			task.UserID, err = primitive.ObjectIDFromHex(id)
			if err != nil {
				res := helper.BuildErrorResponse("Process failed", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(res)
				return
			}

			task.Status = entity.Processing
			task.Quantity = 1
			task.CADFiles = append(task.CADFiles, cadFile.FileName)
			task.CreatedAt = time.Now().Unix()

			var res helper.Response
			if cadFile.FeatureProps.ProcessLevel == 0 {
				res = helper.BuildResponse(true, "Feature recognition started", &helper.EmptyObj{})

				task.ProcessType = append(task.ProcessType, entity.FeatureRecognition)
				task, err := c.taskService.Create(&task)
				if err != nil {
					response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}

				c.ExtractBendFeatures(id, task.TaskID.Hex(), cadFile)
			} else if cadFile.FeatureProps.ProcessLevel == 1 {
				res = helper.BuildResponse(true, "Process planning started", &helper.EmptyObj{})

				task.ProcessType = append(task.ProcessType, entity.ProcessPlanning)
				task, err := c.taskService.Create(&task)
				if err != nil {
					response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}

				c.GenerateProcessingPlan(id, task.TaskID.Hex(), cadFile)
			}

			go persistence.ClearCache(TASKCACHE)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(res)
			return
		}

		processingPlan, err := c.processingPlanService.Find(cadFileID)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "Processing plan not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "CAD file has been fully processed", processingPlan)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

// BatchProcessCADFiles -
func (c *freController) BatchProcessCADFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
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

		var cadFilesToProcess []string
		err := json.NewDecoder(r.Body).Decode(&cadFilesToProcess)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		cadFiles, err := c.cadFileService.FindSelected(cadFilesToProcess)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "CAD file not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		var freNum = 0
		var ppNum = 0

		var task entity.Task
		task.TaskID = primitive.NewObjectID()
		task.UserID, err = primitive.ObjectIDFromHex(id)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}

		task.ProcessType = []entity.ProcessType{}
		task.Status = entity.Processing
		task.CreatedAt = time.Now().Unix()

		for _, cadFile := range cadFiles {
			if cadFile.FeatureProps.ProcessLevel == 0 {
				res := helper.BuildResponse(true, fmt.Sprintf("Feature recognition started: %v", cadFile.FileName), &helper.EmptyObj{})
				c.ExtractBendFeatures(id, task.TaskID.Hex(), &cadFile)
				task.ProcessType = append(task.ProcessType, entity.FeatureRecognition)
				json.NewEncoder(w).Encode(res)
				freNum++
			} else if cadFile.FeatureProps.ProcessLevel == 1 {
				res := helper.BuildResponse(true, fmt.Sprintf("Process planning started: %v", cadFile.FileName), &helper.EmptyObj{})
				c.GenerateProcessingPlan(id, task.TaskID.Hex(), &cadFile)
				task.ProcessType = append(task.ProcessType, entity.ProcessPlanning)
				json.NewEncoder(w).Encode(res)
				ppNum++
			}

			task.CADFiles = append(task.CADFiles, cadFile.FileName)
		}

		if ppNum > 0 || freNum > 0 {
			task.Quantity = int64(ppNum) + int64(freNum)
			return
		}

		_, err = c.taskService.Create(&task)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		go persistence.ClearCache(TASKCACHE)

		res := helper.BuildResponse(true, "Process planning strted.", &helper.EmptyObj{})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}
