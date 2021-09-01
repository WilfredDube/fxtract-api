package controller

import (
	"encoding/json"
	"net/http"
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
	MATERIALCACHE = "tools"
)

type materialController struct {
	userService     service.UserService
	materialService service.MaterialService
	jwtService      service.JWTService
	cache           *redis.Client
}

// MaterialController -
type MaterialController interface {
	AddMaterial(w http.ResponseWriter, r *http.Request)
	Find(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewMaterialController -
func NewMaterialController(service service.MaterialService, uService service.UserService, jwtService service.JWTService, cache *redis.Client) MaterialController {
	return &materialController{
		userService:     uService,
		materialService: service,
		jwtService:      jwtService,
		cache:           cache,
	}
}

// NewProject - add a new material
func (c *materialController) AddMaterial(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}
	var response *entity.Material

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		material := &entity.Material{}
		err := json.NewDecoder(r.Body).Decode(material)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		result, _ := c.materialService.Find(material.Name)
		if result != nil {
			response := helper.BuildErrorResponse("Material already exist", "Duplicate request", helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		err = c.materialService.Validate(material)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		material.CreatedAt = time.Now().Unix()

		response, err = c.materialService.Create(material)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		go persistence.ClearCache(MATERIALCACHE)

		res := helper.BuildResponse(true, "OK", response)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Failed to process request", "Material creation failed", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// Find -
func (c *materialController) Find(w http.ResponseWriter, r *http.Request) {
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

		var material *entity.Material
		if err != nil {

			material, err = c.materialService.Find(id)
			if err != nil {
				res := helper.BuildErrorResponse("Material not found", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(res)
				return
			}
			bytes, err := json.Marshal(material)
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
			json.Unmarshal([]byte(result), &material)
		}

		res := helper.BuildResponse(true, "OK!", material)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Material not found", "Unknown material ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// FindAll -
func (c *materialController) FindAll(w http.ResponseWriter, r *http.Request) {
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

		var materials []entity.Material
		if err != nil {

			materials, err = c.materialService.FindAll()
			if err != nil {
				res := helper.BuildErrorResponse("Material not found", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(res)
				return
			}
			bytes, err := json.Marshal(materials)
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
			json.Unmarshal([]byte(result), &materials)
		}

		res := helper.BuildResponse(true, "OK", materials)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Material not found", "Unknown material ID", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// Delete -
func (c *materialController) Delete(w http.ResponseWriter, r *http.Request) {
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

		_, err := c.materialService.Find(id)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		deleteCount, err := c.materialService.Delete(id)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		if 0 == deleteCount {
			response := helper.BuildErrorResponse("Failed to process request", "Material not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		go persistence.ClearCache(id)
		go persistence.ClearCache(MATERIALCACHE)

		res := helper.BuildResponse(true, "OK", helper.EmptyObj{})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	response := helper.BuildErrorResponse("Failed to process request", "Material deletion failed", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}
