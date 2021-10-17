package service

import (
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
)

var (
	taskRepo repository.TaskRepository
)

// TaskService -
type TaskService interface {
	Create(task *entity.Task) (*entity.Task, error)
	Update(task *entity.Task) (*entity.Task, error)
	Find(id string) (*entity.Task, error)
	FindByUserID(id string) (*entity.Task, error)
	FindAll() ([]entity.Task, error)
}

type taskService struct{}

// NewTaskService -
func NewTaskService(dbRepository repository.TaskRepository) TaskService {
	taskRepo = dbRepository
	return &taskService{}
}

func (*taskService) Create(task *entity.Task) (*entity.Task, error) {
	return taskRepo.Create(task)
}

func (*taskService) Update(task *entity.Task) (*entity.Task, error) {
	return taskRepo.Update(task)
}

func (*taskService) Find(id string) (*entity.Task, error) {
	return taskRepo.Find(id)
}

func (*taskService) FindByUserID(id string) (*entity.Task, error) {
	return taskRepo.FindByUserID(id)
}

func (*taskService) FindAll() ([]entity.Task, error) {
	return taskRepo.FindAll()
}
