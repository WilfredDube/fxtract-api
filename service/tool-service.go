package service

import (
	"errors"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
)

var (
	toolRepo repository.ToolRepository
)

// ToolService -
type ToolService interface {
	Validate(tool *entity.Tool) error
	Create(tool *entity.Tool) (*entity.Tool, error)
	Find(id string) (*entity.Tool, error)
	FindByAngle(angle int64) (*entity.Tool, error)
	FindAll() ([]entity.Tool, error)
	Delete(id string) (int64, error)
}

type toolService struct{}

// NewToolService -
func NewToolService(dbRepository repository.ToolRepository) ToolService {
	toolRepo = dbRepository
	return &toolService{}
}

func (*toolService) Validate(tool *entity.Tool) error {
	if tool == nil {
		return errors.New("Tool is empty")
	}

	if tool.ToolID == "" || tool.ToolName == "" || tool.Angle == 0 || tool.Length == 0 || tool.MinRadius == 0 || tool.MaxRadius == 0 {
		return errors.New("Fill in all the fields")
	}

	return nil
}

func (*toolService) Create(tool *entity.Tool) (*entity.Tool, error) {
	return toolRepo.Create(tool)
}

func (*toolService) Find(id string) (*entity.Tool, error) {
	return toolRepo.Find(id)
}

func (*toolService) FindByAngle(angle int64) (*entity.Tool, error) {
	return toolRepo.FindByAngle(angle)
}

func (*toolService) FindAll() ([]entity.Tool, error) {
	return toolRepo.FindAll()
}

func (*toolService) Delete(id string) (int64, error) {
	return toolRepo.Delete(id)
}
