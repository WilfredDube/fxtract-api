package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/helper"
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
	Upload(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	uploadHandler(r *http.Request, projectFolder string, id primitive.ObjectID) (*[]entity.CADFile, error)
}

// NewCADFileController -
func NewCADFileController(service service.CadFileService, pService service.ProjectService, jwtService service.JWTService) CadFileController {
	return &cadFileController{
		cadFileService: service,
		projectService: pService,
		jwtService:     jwtService,
	}
}

// Upload - add a new cadFile
func (c *cadFileController) Upload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request: ", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		ownerID, _ := primitive.ObjectIDFromHex(claims["user_id"].(string))

		params := mux.Vars(r)
		id, err := primitive.ObjectIDFromHex(params["id"])

		project, err := c.projectService.Find(params["id"])
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

		projectFolder := "./uploads/" + project.OwnerID.Hex() + "/" + project.ID.Hex()

		uploadedFiles, err := c.uploadHandler(r, projectFolder, id)
		if err != nil {
			res := helper.BuildErrorResponse("Upload error", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(res)
			return
		}

		res := helper.BuildResponse(true, "Upload complete : OK!", uploadedFiles)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	res := helper.BuildErrorResponse("Upload error", "File upload failed", helper.EmptyObj{})
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(res)
	return
}

func (c *cadFileController) uploadHandler(r *http.Request, projectFolder string, id primitive.ObjectID) (*[]entity.CADFile, error) {
	var uploadedFiles []entity.CADFile
	tempCache := make(map[string]string)
	fileCache := make(map[string]entity.CADFile)

	// 32 MB is the default used by FormFile()
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}

	// Get a reference to the fileHeaders.
	// They are accessible only after ParseMultipartForm is called
	files := r.MultipartForm.File["file"]
	material := r.MultipartForm.Value["material"][0]

	nFiles := len(files)

	if (nFiles % 2) != 0 {
		return nil, fmt.Errorf("Each STEP/IGES file must be uploaded with its corresponding obj file")
	}

	if !helper.UploadBalanced(files) {
		return nil, fmt.Errorf("Unbalanced: Each STEP/IGES file must be uploaded with its corresponding obj file")
	}

	for _, fileHeader := range files {
		// Restrict the size of each uploaded file to 1MB.
		// To prevent the aggregate size from exceeding
		// a specified value, use the http.MaxBytesReader() method
		// before calling ParseMultipartForm()
		if fileHeader.Size > MaxUploadSize {
			return nil, fmt.Errorf("The uploaded image is too big: %s. Please use an image less than 1MB in size", fileHeader.Filename)
		}

		ext := filepath.Ext(fileHeader.Filename)
		if ext != ".stp" && ext != ".step" && ext != ".obj" {
			return nil, fmt.Errorf("The provided file format is not allowed. %s", ext)
		}

		// Open the file
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}

		defer file.Close()

		buff := make([]byte, 512)
		_, err = file.Read(buff)
		if err != nil {
			return nil, err
		}

		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}

		processed := false
		newName := time.Now().UnixNano()
		filename := helper.FileNameWithoutExtSlice(fileHeader.Filename)
		if _, ok := tempCache[filename]; ok {
			newName, _ = strconv.ParseInt(tempCache[filename], 10, 64)
			processed = true
		}

		f, err := os.Create(fmt.Sprintf(projectFolder+"/%d%s", newName, filepath.Ext(fileHeader.Filename)))
		if err != nil {
			return nil, err
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			return nil, err
		}

		// insert cad file file metadata into database
		var cadFile entity.CADFile

		if processed == false {
			tempCache[filename] = helper.FileNameWithoutExtSlice(filepath.Base(f.Name()))

			cadFile.ID = primitive.NewObjectID()
			cadFile.FileName = filename + ".stp"
			if ext == ".stp" || ext == ".step" {
				cadFile.StepURL = f.Name()
			} else {
				cadFile.ObjpURL = f.Name()
			}

			cadFile.Material = material
			cadFile.Filesize = fileHeader.Size
			cadFile.CreatedAt = time.Now().Unix()
			cadFile.ProjectID = id

			fileCache[tempCache[filename]] = cadFile

			_, err := c.cadFileService.Create(&cadFile)
			if err != nil {
				return nil, err // error updloading file failed
			}
		} else {
			fl := tempCache[filename]
			cadFile = fileCache[fl]

			if ext == ".stp" || ext == ".step" {
				cadFile.StepURL = f.Name()
			} else {
				cadFile.ObjpURL = f.Name()
			}

			_, err := c.cadFileService.Update(cadFile)
			if err != nil {
				return nil, err
			}

			uploadedFiles = append(uploadedFiles, cadFile)
			delete(tempCache, filename)
		}
	}

	return &uploadedFiles, nil
}

// TODO: fix update for FRE
func (c *cadFileController) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cadFile := &entity.CADFile{}
	err := json.NewDecoder(r.Body).Decode(cadFile)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		panic(errToken.Error())
	}

	claims := token.Claims.(jwt.MapClaims)
	uid, err := primitive.ObjectIDFromHex(claims["id"].(string))
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

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request: ", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
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

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request: ", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
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

// Delete -
func (c *cadFileController) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request: ", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
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

		if 0 == deleteCount {
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
