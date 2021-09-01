package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/helper"
	persistence "github.com/WilfredDube/fxtract-backend/repository/reposelect"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

const (
	TOOLCACHE = "tools"
)

type toolController struct {
	userService service.UserService
	toolService service.ToolService
	jwtService  service.JWTService
	cache       *redis.Client
}

// ToolController -
type ToolController interface {
	AddTool(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	FindByAngle(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewToolController -
func NewToolController(service service.ToolService, uService service.UserService, jwtService service.JWTService, cache *redis.Client) ToolController {
	return &toolController{
		userService: uService,
		toolService: service,
		jwtService:  jwtService,
		cache:       cache,
	}
}

// NewProject - add a new tool
func (c *toolController) AddTool(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	var response *entity.Tool

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tool := &entity.Tool{}
		err := json.NewDecoder(r.Body).Decode(tool)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		result, _ := c.toolService.Find(tool.ToolID)
		if result != nil {
			response := helper.BuildErrorResponse("Tool already exist", "Duplicate request", helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		err = c.toolService.Validate(tool)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		tool.CreatedAt = time.Now().Unix()

		response, err = c.toolService.Create(tool)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		go persistence.ClearCache(TOOLCACHE)

		res := helper.BuildResponse(true, "OK", response)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Failed to process request", "Tool creation failed", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// FindByID -
func (c *toolController) FindByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		params := mux.Vars(r)
		id := params["id"]

		result, err := c.cache.Get(id).Result()

		var tool *entity.Tool
		if err != nil {
			tool, err = c.toolService.Find(id)
			if err != nil {
				res := helper.BuildErrorResponse("Tool not found", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(res)
				return
			}
			bytes, err := json.Marshal(tool)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			if err := c.cache.Set(id, bytes, 30*time.Minute).Err(); err != nil {
				response := helper.BuildErrorResponse("Failed to cache request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		} else {
			json.Unmarshal([]byte(result), &tool)
		}

		res := helper.BuildResponse(true, "OK!", tool)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Tool not found", "Unknown tool ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

func (c *toolController) FindByAngle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		params := mux.Vars(r)
		value := params["angle"]
		angle, _ := strconv.ParseInt(value, 10, 64)

		result, err := c.cache.Get(fmt.Sprint(angle)).Result()

		var tool *entity.Tool
		if err != nil {

			tool, err = c.toolService.FindByAngle(angle)
			if err != nil {
				res := helper.BuildErrorResponse("Tool not found", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(res)
				return
			}
			bytes, err := json.Marshal(tool)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			if err := c.cache.Set(fmt.Sprint(angle), bytes, 30*time.Minute).Err(); err != nil {
				response := helper.BuildErrorResponse("Failed to cache request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		} else {
			json.Unmarshal([]byte(result), &tool)
		}

		res := helper.BuildResponse(true, "OK!", tool)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Tool not found", "Unknown tool ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// FindAll -
func (c *toolController) FindAll(w http.ResponseWriter, r *http.Request) {
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

		var tools []entity.Tool
		if err != nil {
			tools, err = c.toolService.FindAll()
			if err != nil {
				res := helper.BuildErrorResponse("Tool not found", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(res)
				return
			}

			bytes, err := json.Marshal(tools)
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
			json.Unmarshal([]byte(result), &tools)
		}

		res := helper.BuildResponse(true, "OK", tools)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Tool not found", "Unknown tool ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// Delete -
func (c *toolController) Delete(w http.ResponseWriter, r *http.Request) {
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
		id = params["id"]

		tool, err := c.toolService.Find(id)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		deleteCount, err := c.toolService.Delete(id)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		if deleteCount == 0 {
			response := helper.BuildErrorResponse("Failed to process request", "Tool not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		go persistence.ClearCache(id)
		go persistence.ClearCache(fmt.Sprint(tool.Angle))
		go persistence.ClearCache(TOOLCACHE)

		res := helper.BuildResponse(true, "OK", helper.EmptyObj{})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	response := helper.BuildErrorResponse("Failed to process request", "Tool deletion failed", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}
