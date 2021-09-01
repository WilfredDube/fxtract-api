package service

import (
	"errors"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
)

var (
	materialRepo repository.MaterialRepository
)

// MaterialService -
type MaterialService interface {
	Validate(material *entity.Material) error
	Create(material *entity.Material) (*entity.Material, error)
	Find(id string) (*entity.Material, error)
	FindAll() ([]entity.Material, error)
	Delete(id string) (int64, error)
}

type materialService struct{}

// NewMaterialService -
func NewMaterialService(dbRepository repository.MaterialRepository) MaterialService {
	materialRepo = dbRepository
	return &materialService{}
}

func (*materialService) Validate(material *entity.Material) error {
	if material == nil {
		return errors.New("material is empty")
	}

	if material.Name == "" || material.TensileStrength == 0 || material.KFactor == 0 {
		return errors.New("fill in all the fields")
	}

	return nil
}

func (*materialService) Create(material *entity.Material) (*entity.Material, error) {
	return materialRepo.Create(material)
}

func (*materialService) Find(id string) (*entity.Material, error) {
	return materialRepo.Find(id)
}

func (*materialService) FindAll() ([]entity.Material, error) {
	return materialRepo.FindAll()
}

func (*materialService) Delete(id string) (int64, error) {
	return materialRepo.Delete(id)
}
