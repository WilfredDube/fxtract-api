package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

const (
	TASKCACHE = "tasks"
)

type taskController struct {
	taskService service.TaskService
	jwtService  service.JWTService
	cache       *redis.Client
}

// TaskController -
type TaskController interface {
	Find(w http.ResponseWriter, r *http.Request)
	FindByUserID(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
}

// NewTaskController -
func NewTaskController(service service.TaskService, jwtService service.JWTService, cache *redis.Client) TaskController {
	return &taskController{
		taskService: service,
		jwtService:  jwtService,
		cache:       cache,
	}
}

// Find -
func (c *taskController) Find(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	// TODO: fix cache for all methods

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		params := mux.Vars(r)
		id := params["id"]

		// result, err := c.cache.Get(id).Result()

		var task *entity.Task
		// if err != nil {

		task, err = c.taskService.Find(id)
		if err != nil {
			res := helper.BuildErrorResponse("Task not found", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}
		// bytes, err := json.Marshal(task)
		// if err != nil {
		// 	response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	json.NewEncoder(w).Encode(response)
		// 	return
		// }

		// if err := c.cache.Set(id, bytes, 30*time.Minute).Err(); err != nil {
		// 	response := helper.BuildErrorResponse("Failed to cache request", err.Error(), helper.EmptyObj{})
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	json.NewEncoder(w).Encode(response)
		// 	return
		// }
		// } else {
		// json.Unmarshal([]byte(result), &task)
		// }

		if task.Status == entity.Processing && task.Quantity == int64(len(task.ProcessedCADFiles)) {
			res := helper.BuildResponse(true, string(entity.Processing), &ProcessResult{TaskID: id})
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, string(entity.Complete), task)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Task not found", "Unknown task ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// FindByUserID -
func (c *taskController) FindByUserID(w http.ResponseWriter, r *http.Request) {
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

		// result, err := c.cache.Get(TOOLCACHE).Result()

		var tasks []entity.Task
		// if err != nil {

		tasks, err = c.taskService.FindByUserID(id)
		if err != nil {
			res := helper.BuildErrorResponse("Task not found", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		// 	bytes, err := json.Marshal(tasks)
		// 	if err != nil {
		// 		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		json.NewEncoder(w).Encode(response)
		// 		return
		// 	}

		// 	if err := c.cache.Set(TOOLCACHE, bytes, 30*time.Minute).Err(); err != nil {
		// 		response := helper.BuildErrorResponse("Failed to cache request", err.Error(), helper.EmptyObj{})
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		json.NewEncoder(w).Encode(response)
		// 		return
		// 	}
		// } else {
		// 	json.Unmarshal([]byte(result), &tasks)
		// }

		res := helper.BuildResponse(true, "OK", tasks)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Task not found", "Unknown task ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// FindAll -
func (c *taskController) FindAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result, err := c.cache.Get(TOOLCACHE).Result()

		var tasks []entity.Task
		if err != nil {

			tasks, err = c.taskService.FindAll()
			if err != nil {
				res := helper.BuildErrorResponse("Task not found", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(res)
				return
			}

			bytes, err := json.Marshal(tasks)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			if err := c.cache.Set(TOOLCACHE, bytes, 30*time.Minute).Err(); err != nil {
				response := helper.BuildErrorResponse("Failed to cache request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		} else {
			json.Unmarshal([]byte(result), &tasks)
		}

		res := helper.BuildResponse(true, "OK", tasks)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Task not found", "Unknown task ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}
