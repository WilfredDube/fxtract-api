package service

import (
	"errors"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
)

var (
	userRepo repository.UserRepository
)

// UserService -
type UserService interface {
	Validate(user *entity.User) error
	Update(user *entity.User) (*entity.User, error)
	Profile(id string) (*entity.User, error)
	GetAll() ([]entity.User, error)
	Delete(id string) (int64, error)
}

type userService struct{}

// NewUserService -
func NewUserService(dbRepository repository.UserRepository) UserService {
	userRepo = dbRepository
	return &userService{}
}

func (*userService) Validate(user *entity.User) error {
	if user == nil {
		return errors.New("User is empty")
	}

	if user.Firstname == "" || user.Lastname == "" || user.Email == "" || user.Password == "" {
		return errors.New("Title or description can not be empty")
	}

	return nil
}

func (*userService) Update(user *entity.User) (*entity.User, error) {
	return userRepo.Update(*user)
}

func (*userService) Profile(id string) (*entity.User, error) {
	return userRepo.Profile(id)
}

func (*userService) GetAll() ([]entity.User, error) {
	return userRepo.FindAll()
}

func (*userService) Delete(id string) (int64, error) {
	return userRepo.Delete(id)
}
