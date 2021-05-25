package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type cadFileController struct {
	cadFileService service.CadFileService
	jwtService     service.JWTService
}

// CadFileController -
type CadFileController interface {
	AddProject(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewCADFileController -
func NewCADFileController(service service.CadFileService, jwtService service.JWTService) CadFileController {
	return &cadFileController{
		cadFileService: service,
		jwtService:     jwtService,
	}
}

// NewProject - add a new cadFile
func (c *cadFileController) AddProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cadFile := &entity.CADFile{}
	err := json.NewDecoder(r.Body).Decode(cadFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// json.NewEncoder(w).Encode(errors.ServiceError{Message: "Error unmarshalling data!"})
		return
	}

	err = c.cadFileService.Validate(cadFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// json.NewEncoder(w).Encode(errors.ServiceError{Message: err.Error()})
		return
	}

	cadFile.ID = primitive.NewObjectID()
	// cadFile.OwnerID = user.ID
	cadFile.CreatedAt = time.Now().Unix()
	// cadFile.CadFiles = []primitive.ObjectID{}

	response, err := c.cadFileService.Create(cadFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// json.NewEncoder(w).Encode(errors.ServiceError{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// FindByID -
func (c *cadFileController) FindByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id := params["id"]

	cadFile, err := c.cadFileService.Find(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// json.NewEncoder(w).Encode(errors.ServiceError{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cadFile)
}

// FindAll -
func (c *cadFileController) FindAll(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	projects, err := c.cadFileService.FindAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// json.NewEncoder(w).Encode(errors.ServiceError{Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(projects)
}

// Delete -
func (c *cadFileController) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id := params["id"]

	deleteCount, err := c.cadFileService.Delete(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// json.NewEncoder(w).Encode(errors.ServiceError{Message: err.Error()})
		return
	}

	if 0 == deleteCount {
		w.WriteHeader(http.StatusNotFound)
		// json.NewEncoder(w).Encode(errors.ServiceError{Message: "CADFile not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(errors.ServiceError{Message: "CADFile deletion successful"})
}
