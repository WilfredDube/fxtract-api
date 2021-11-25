package controller

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MaxUploadSize -
const MaxUploadSize = 1024 * 1024 // 1MB

type cadFileController struct {
	cadFileService        service.CadFileService
	projectService        service.ProjectService
	jwtService            service.JWTService
	processingPlanService service.ProcessingPlanService
	cache                 *redis.Client
}

type OBJCached struct {
	File string `json:"file"`
	Data string `json:"data"`
}

// CadFileController -
type CadFileController interface {
	FindByID(w http.ResponseWriter, r *http.Request)
	DownloadOBJ(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	FindAllFiles(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewCADFileController -
func NewCADFileController(service service.CadFileService, pService service.ProjectService, jwtService service.JWTService, processingPlanService service.ProcessingPlanService, cache *redis.Client) CadFileController {
	return &cadFileController{
		cadFileService:        service,
		projectService:        pService,
		jwtService:            jwtService,
		processingPlanService: processingPlanService,
		cache:                 cache,
	}
}

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
		w.WriteHeader(http.StatusInternalServerError)
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

func (c *cadFileController) DownloadOBJ(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		uri := r.FormValue("url")

		result, err := c.cache.Get(uri).Result()

		var data, filename string
		var objCached *OBJCached
		if err != nil {
			objBlob := service.NewAzureBlobService()

			data, filename, err = objBlob.GetOBj(uri)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			objCached = &OBJCached{File: filename, Data: data}

			bytes, err := json.Marshal(objCached)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err := c.cache.Set(uri, bytes, 10*time.Minute).Err(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			f, err := os.Create("objs/" + objCached.File)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			defer f.Close()

			_, err = f.WriteString(objCached.Data)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		} else {
			json.Unmarshal([]byte(result), &objCached)
		}

		w.WriteHeader(http.StatusOK)
		// json.NewEncoder(w).Encode("http://localhost:8000/objs/" + objCached.File)
		json.NewEncoder(w).Encode(objCached.Data)
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

		cloudService := service.NewAzureBlobService()
		_, err = cloudService.Delete(cadFile.ObjpURL, service.CADFILE)
		if err != nil {
			res := helper.BuildErrorResponse("Deletion failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}

		_, err = cloudService.Delete(cadFile.StepURL, service.CADFILE)
		if err != nil {
			res := helper.BuildErrorResponse("Deletion failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}

		if cadFile.FeatureProps.ProcessLevel == 2 {
			processingPlan, err := c.processingPlanService.Find(cadFile.ID.Hex())
			if err != nil {
				res := helper.BuildErrorResponse("Processing plan not found", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(res)
				return
			}

			_, err = cloudService.Delete(processingPlan.PdfURL, service.PDFFILE)
			if err != nil {
				res := helper.BuildErrorResponse("Deletion failed", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(res)
				return
			}

			_, err = c.processingPlanService.Delete(id)
			if err != nil {
				res := helper.BuildErrorResponse("Deletion failed", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(res)
				return
			}
		}

		deleteCount, err := c.cadFileService.Delete(id)
		if err != nil {
			res := helper.BuildErrorResponse("Deletion failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}

		if deleteCount == 0 {
			res := helper.BuildErrorResponse("Deletion failed", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "OK!", deleteCount)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}
}
