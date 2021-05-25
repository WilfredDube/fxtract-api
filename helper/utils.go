package helper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// UploadFolder -
	UploadFolder = "uploads/"
)

func verifyToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}

		return []byte("secret"), nil
	})
}

// CreateFolder -
func CreateFolder(dir string, isMainFolder bool) {
	dir = UploadFolder + dir

	if isMainFolder {
		dir = "uploads"
	}
	_, err := os.Stat(dir)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dir, 0755)
		if errDir != nil {
			log.Fatal(err)
		}

	}
}

func fileNameWithoutExtSlice(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

// func getCadFileByID(cadFileID primitive.ObjectID) (model.CADFile, error) {
// 	var cadFile model.CADFile
// 	cadFilesCollection, err := db.GetCadModelsCollection()
// 	if err != nil {
// 		return cadFile, err
// 	}

// 	err = cadFilesCollection.FindOne(context.TODO(), bson.D{primitive.E{Key: "_id", Value: cadFileID}}).Decode(&cadFile)
// 	if err != nil {
// 		return cadFile, err
// 	}

// 	return cadFile, nil
// }

// func getProjectByID(projectID primitive.ObjectID) (model.Project, error) {
// 	var project model.Project
// 	projectsCollection, err := db.GetProjectsCollection()
// 	if err != nil {
// 		return project, err
// 	}

// 	err = projectsCollection.FindOne(context.TODO(), bson.D{primitive.E{Key: "_id", Value: projectID}}).Decode(&project)
// 	if err != nil {
// 		return project, err
// 	}

// 	return project, nil
// }

// func getProjectCadFiles(projectCadFiles []primitive.ObjectID) ([]model.CADFile, error) {
// 	var cadFileList []model.CADFile

// 	if len(projectCadFiles) == 0 {
// 		return cadFileList, fmt.Errorf("Project has no CAD files uploaded yet! Please upload some :)")
// 	}

// 	for _, cadFileID := range projectCadFiles {
// 		cadFile, err := getCadFileByID(cadFileID)
// 		if err != nil {
// 			return cadFileList, err
// 		}

// 		cadFileList = append(cadFileList, cadFile)
// 	}

// 	return cadFileList, nil
// }
