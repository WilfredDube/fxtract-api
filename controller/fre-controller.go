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
	"github.com/gorilla/websocket"
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
	processor             *service.Processor
}

type ProcessResult struct {
	Message string `json:"message"`
	TaskID  string `json:"task_id"`
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
	uService service.UserService, jwtService service.JWTService, taskService service.TaskService, cache *redis.Client, eventEmitter msgqueue.EventEmitter, processor *service.Processor) FREController {
	return &freController{
		userService:           uService,
		cadFileService:        cadService,
		processingPlanService: pPlanService,
		jwtService:            jwtService,
		config:                configuration,
		taskService:           taskService,
		cache:                 cache,
		eventEmitter:          eventEmitter,
		processor:             processor,
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

				// task.ProcessType = append(task.ProcessType, entity.FeatureRecognition)
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

				// task.ProcessType = append(task.ProcessType, entity.ProcessPlanning)
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

		var conn *websocket.Conn
		if conn, ok = c.processor.Users[id]; !ok {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			resp, err := json.Marshal(response)
			if err != nil {
				log.Println(err.Error())
			}
			conn.WriteMessage(websocket.TextMessage, resp)
			return
		}

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				resp, err := json.Marshal(response)
				if err != nil {
					log.Println(err.Error())
				}
				conn.WriteMessage(messageType, resp)
				return
			}

			var cadFilesToProcess []string

			err = json.Unmarshal(p, &cadFilesToProcess)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				resp, err := json.Marshal(response)
				if err != nil {
					log.Println(err.Error())
				}
				conn.WriteMessage(messageType, resp)
				return
			}

			cadFiles, err := c.cadFileService.FindSelected(cadFilesToProcess)
			if err != nil {
				response := helper.BuildErrorResponse("Process failed", "CAD file not found", helper.EmptyObj{})
				resp, err := json.Marshal(response)
				if err != nil {
					log.Println(err.Error())
				}
				conn.WriteMessage(messageType, resp)
				return
			}

			var freNum = 0
			var ppNum = 0

			var task entity.Task
			task.ID = primitive.NewObjectID()
			task.TaskID = primitive.NewObjectID()
			task.UserID, err = primitive.ObjectIDFromHex(id)
			if err != nil {
				response := helper.BuildErrorResponse("Process failed", err.Error(), helper.EmptyObj{})
				resp, err := json.Marshal(response)
				if err != nil {
					log.Println(err.Error())
				}
				conn.WriteMessage(messageType, resp)
				return
			}

			task.ProcessedCADFiles = []entity.Processed{}
			task.Status = entity.Processing
			task.CreatedAt = time.Now().Unix()

			for _, cadFile := range cadFiles {
				if cadFile.FeatureProps.ProcessLevel == 0 {
					c.ExtractBendFeatures(id, task.ID.Hex(), &cadFile)
					freNum++
				} else if cadFile.FeatureProps.ProcessLevel == 1 {
					c.GenerateProcessingPlan(id, task.ID.Hex(), &cadFile)
					ppNum++
				}

				task.CADFiles = append(task.CADFiles, cadFile.FileName)
			}

			resultString := ""
			if ppNum > 0 && freNum > 0 {
				task.Quantity = int64(ppNum) + int64(freNum)
				resultString = fmt.Sprintf("%d feature recognition process(es) and %d process planning process(es) started", freNum, ppNum)
			} else if ppNum > 0 {
				task.Quantity = int64(ppNum)
				resultString = fmt.Sprintf("%d process planning process(es) started", ppNum)
			} else if freNum > 0 {
				task.Quantity = int64(freNum)
				resultString = fmt.Sprintf("%d feature recognition process(es) started", freNum)
			}

			task.Description = resultString

			_, err = c.taskService.Create(&task)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				resp, err := json.Marshal(response)
				if err != nil {
					log.Println(err.Error())
				}
				conn.WriteMessage(messageType, resp)
				return
			}

			go persistence.ClearCache(TASKCACHE)

			response := helper.Response{
				Status:  true,
				Message: resultString,
				Type:    "init",
				Errors:  nil,
				Data:    &helper.EmptyObj{},
			}
			resp, err := json.Marshal(response)
			if err != nil {
				log.Println(err.Error())
			}
			conn.WriteMessage(messageType, resp)
		}
	}
}
