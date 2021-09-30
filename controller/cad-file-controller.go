package controller

import (
	"encoding/json"
	"net/http"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MaxUploadSize -
const MaxUploadSize = 1024 * 1024 // 1MB

type cadFileController struct {
	cadFileService service.CadFileService
	projectService service.ProjectService
	jwtService     service.JWTService
}

// CadFileController -
type CadFileController interface {
	FindByID(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	FindAllFiles(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewCADFileController -
func NewCADFileController(service service.CadFileService, pService service.ProjectService, jwtService service.JWTService) CadFileController {
	return &cadFileController{
		cadFileService: service,
		projectService: pService,
		jwtService:     jwtService,
	}
}

// TODO: fix update for FRE
func (c *cadFileController) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	cadFile := &entity.CADFile{}
	err = json.NewDecoder(r.Body).Decode(cadFile)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	claims := token.Claims.(jwt.MapClaims)
	uid, err := primitive.ObjectIDFromHex(claims["id"].(string))
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	cadFile.ID = uid

	u, err := c.cadFileService.Update(*cadFile)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "OK!", u)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// FindByID -
func (c *cadFileController) FindByID(w http.ResponseWriter, r *http.Request) {
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

		cadFile, err := c.cadFileService.Find(id)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "CAD file not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "OK!", cadFile)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}

// FindAll -
func (c *cadFileController) FindAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ownerID, _ := primitive.ObjectIDFromHex(claims["user_id"].(string))

		params := mux.Vars(r)
		projectID := params["id"]

		project, err := c.projectService.Find(projectID)
		if err != nil {
			res := helper.BuildErrorResponse("Project error", "The project you want to upload to does not exist", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		if project.OwnerID != ownerID {
			res := helper.BuildErrorResponse("Project owner does not exist", "Token error", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		cadFiles, err := c.cadFileService.FindAll(projectID)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "OK!", cadFiles)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}

func (c *cadFileController) FindAllFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		_, _ = primitive.ObjectIDFromHex(claims["user_id"].(string))

		projects, err := c.cadFileService.FindAllFiles()
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "OK!", projects)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}

// Delete -
func (c *cadFileController) Delete(w http.ResponseWriter, r *http.Request) {
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

		cadFile, err := c.cadFileService.Find(id)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", "CAD file not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		deleteCount, err := c.cadFileService.Delete(id)
		if err != nil {
			res := helper.BuildErrorResponse("Process failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		if deleteCount == 0 {
			res := helper.BuildErrorResponse("File not found: ", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(res)
			return
		}

		helper.DeleteFile(cadFile.StepURL)
		helper.DeleteFile(cadFile.ObjpURL)

		res := helper.BuildResponse(true, "OK!", deleteCount)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}
