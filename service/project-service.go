package service

import (
	"errors"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	repo repository.ProjectRepository
)

// ProjectService -
type ProjectService interface {
	Validate(project *entity.Project) error
	Create(project *entity.Project) (*entity.Project, error)
	Find(id string) (*entity.Project, error)
	FindByName(name string) (*entity.Project, error)
	IsDuplicate(name string, OwnerID primitive.ObjectID) bool
	FindAll(ownerID string) ([]entity.Project, error)
	Delete(id string) (int64, error)
}

type service struct{}

// NewProjectService -
func NewProjectService(dbRepository repository.ProjectRepository) ProjectService {
	repo = dbRepository
	return &service{}
}

func (*service) Validate(project *entity.Project) error {
	if project == nil {
		return errors.New("Project is empty")
	}

	if project.Title == "" || project.Description == "" {
		return errors.New("Title or description can not be empty")
	}

	return nil
}

func (*service) Create(project *entity.Project) (*entity.Project, error) {
	return repo.Create(project)
}

func (*service) Find(id string) (*entity.Project, error) {
	return repo.Find(id)
}

func (*service) FindByName(name string) (*entity.Project, error) {
	return repo.FindByName(name)
}

func (*service) IsDuplicate(name string, OwnerID primitive.ObjectID) bool {
	return repo.IsDuplicate(name, OwnerID)
}
func (*service) FindAll(ownerID string) ([]entity.Project, error) {
	return repo.FindAll(ownerID)
}

func (*service) Delete(id string) (int64, error) {
	return repo.Delete(id)
}
