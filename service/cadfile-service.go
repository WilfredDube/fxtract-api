package service

import (
	"errors"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
)

var (
	cadFileRepo repository.CADFileRepository
)

// CadFileService -
type CadFileService interface {
	Validate(cadFile *entity.CADFile) error
	Create(cadFile *entity.CADFile) (*entity.CADFile, error)
	Find(id string) (*entity.CADFile, error)
	FindAll() ([]entity.CADFile, error)
	Delete(id string) (int64, error)
}

type cadFileService struct{}

// NewCadFileService -
func NewCadFileService(dbRepository repository.CADFileRepository) CadFileService {
	cadFileRepo = dbRepository
	return &cadFileService{}
}

func (*cadFileService) Validate(cadFile *entity.CADFile) error {
	if cadFile == nil {
		return errors.New("CADFile is empty")
	}

	// if cadFile.Title == "" || cadFile.Description == "" {
	// 	return errors.New("Title or description can not be empty")
	// }

	return nil
}

func (*cadFileService) Create(cadFile *entity.CADFile) (*entity.CADFile, error) {
	return cadFileRepo.Create(cadFile)
}

func (*cadFileService) Find(id string) (*entity.CADFile, error) {
	return cadFileRepo.Find(id)
}

func (*cadFileService) FindAll() ([]entity.CADFile, error) {
	return cadFileRepo.FindAll()
}

func (*cadFileService) Delete(id string) (int64, error) {
	return cadFileRepo.Delete(id)
}
