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
	Update(cadFile entity.CADFile) (*entity.CADFile, error)
	Find(id string) (*entity.CADFile, error)
	FindAll(projectID string) ([]entity.CADFile, error)
	Delete(id string) (int64, error)
	CascadeDelete(id string) (int64, error)
}

type cadFileService struct{}

// NewCadFileService -
func NewCadFileService(dbRepository repository.CADFileRepository) CadFileService {
	cadFileRepo = dbRepository
	return &cadFileService{}
}

func (*cadFileService) Validate(cadFile *entity.CADFile) error {
	if cadFile == nil || cadFile.FileName == "" {
		return errors.New("CADFile is empty")
	}

	return nil
}

func (c *cadFileService) Create(cadFile *entity.CADFile) (*entity.CADFile, error) {
	if err := c.Validate(cadFile); err != nil {
		return nil, err
	}

	return cadFileRepo.Create(cadFile)
}

func (*cadFileService) Update(cadFile entity.CADFile) (*entity.CADFile, error) {
	return cadFileRepo.Update(cadFile)
}

func (*cadFileService) Find(id string) (*entity.CADFile, error) {
	return cadFileRepo.Find(id)
}

func (*cadFileService) FindAll(projectID string) ([]entity.CADFile, error) {
	return cadFileRepo.FindAll(projectID)
}

func (*cadFileService) Delete(id string) (int64, error) {
	return cadFileRepo.Delete(id)
}

func (*cadFileService) CascadeDelete(projectID string) (int64, error) {
	return cadFileRepo.CascadeDelete(projectID)
}
