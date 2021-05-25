package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type controller struct {
	projectService service.ProjectService
	jwtService     service.JWTService
}

// ProjectController -
type ProjectController interface {
	AddProject(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewProjectController -
func NewProjectController(service service.ProjectService, jwtService service.JWTService) ProjectController {
	return &controller{
		projectService: service,
		jwtService:     jwtService,
	}
}

// NewProject - add a new project
func (c *controller) AddProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	var response *entity.Project

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["user_id"].(string)

		project := &entity.Project{}
		err := json.NewDecoder(r.Body).Decode(project)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		err = c.projectService.Validate(project)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		project.ID = primitive.NewObjectID()
		project.OwnerID, _ = primitive.ObjectIDFromHex(id)
		project.CreatedAt = time.Now().Unix()

		response, err = c.projectService.Create(project)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		projectFolder := response.ID.Hex() + "/" + project.ID.Hex()
		helper.CreateFolder(projectFolder, false)

		res := helper.BuildResponse(true, "OK", response)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Failed to process request", "Project creation failed", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// FindByID -
func (c *controller) FindByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["user_id"].(string)

		params := mux.Vars(r)
		id = params["id"]

		project, err := c.projectService.Find(id)
		if err != nil {
			res := helper.BuildErrorResponse("Project not found", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildErrorResponse("Project not found", "OK!", project)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Project not found", errToken.Error(), helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// FindAll -
func (c *controller) FindAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ownerID := claims["user_id"].(string)

		projects, err := c.projectService.FindAll(ownerID)
		if err != nil {
			res := helper.BuildErrorResponse("Project not found", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "OK", projects)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Project not found", errToken.Error(), helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(res)
}

// Delete -
func (c *controller) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		params := mux.Vars(r)
		id := params["id"]

		deleteCount, err := c.projectService.Delete(id)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", errToken.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		if 0 == deleteCount {
			response := helper.BuildErrorResponse("Failed to process request", "Project not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		res := helper.BuildResponse(true, "OK", helper.EmptyObj{})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	response := helper.BuildErrorResponse("Failed to process request", "Project deletion failed", helper.EmptyObj{})
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}
