package helper

import (
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

const (
	// UploadFolder -
	UploadFolder = "uploads/"
)

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

// DeleteFile -
func DeleteFile(file string) {
	_, err := os.Stat(file)

	if os.IsNotExist(err) {
		log.Fatal(err)
	}

	os.Remove(file)
}

// DeleteFolder -
func DeleteFolder(dir string) {
	dir = UploadFolder + dir
	_, err := os.Stat(dir)

	if os.IsNotExist(err) {
		log.Fatal(err)
	}

	os.RemoveAll(dir)
}

// FileNameWithoutExtSlice -
func FileNameWithoutExtSlice(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}

// UploadBalanced -
func UploadBalanced(files []*multipart.FileHeader) bool {
	count := 0
	for _, file := range files {
		if filepath.Ext(file.Filename) == ".obj" {
			count++
		}
	}

	length := len(files) / 2

	return (count == length)
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
