package service

import (
	"log"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
	"golang.org/x/crypto/bcrypt"
)

//AuthService is a contract about something that this service can do
type AuthService interface {
	VerifyCredential(email string, password string) (entity.User, error)
	CreateUser(user entity.User) (entity.User, error)
	FindByEmail(email string) entity.User
	IsDuplicateEmail(email string) bool
	UpdateUserVerificationStatus(email string, true bool) error
}

type authService struct {
	userRepository repository.UserRepository
}

//NewAuthService creates a new instance of AuthService
func NewAuthService(userRep repository.UserRepository) AuthService {
	return &authService{
		userRepository: userRep,
	}
}

func (service *authService) VerifyCredential(email string, password string) (entity.User, error) {
	res, err := service.userRepository.VerifyCredential(email, password)

	return res, err
}

func (service *authService) CreateUser(user entity.User) (entity.User, error) {
	res, err := service.userRepository.Create(&user)
	if err != nil {
		return entity.User{}, err
	}

	return *res, nil
}

func (service *authService) FindByEmail(email string) entity.User {
	return *service.userRepository.FindByEmail(email)
}

func (service *authService) IsDuplicateEmail(email string) bool {
	res := service.userRepository.IsDuplicateEmail(email)
	return (res == nil)
}

func (service *authService) UpdateUserVerificationStatus(email string, status bool) error {
	return service.userRepository.UpdateUserVerificationStatus(email, status)
}

func comparePassword(hashedPwd string, plainPassword []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPassword)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}
