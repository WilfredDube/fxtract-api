package service

import (
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
)

var (
	verificationRepo repository.VerificationRepository
)

// VerificationService -
type VerificationService interface {
	Create(verification *entity.Verification) (*entity.Verification, error)
	Find(email string, verificationType entity.VerificationDataType) (*entity.Verification, error)
	Update(verification *entity.Verification) (*entity.Verification, error)
	FindAll() ([]entity.Verification, error)
	Delete(id string) (int64, error)
}

type verificationService struct{}

// NewVerificationService -
func NewVerificationService(dbRepository repository.VerificationRepository) VerificationService {
	verificationRepo = dbRepository
	return &verificationService{}
}

func (*verificationService) Create(verification *entity.Verification) (*entity.Verification, error) {
	return verificationRepo.Create(verification)
}

func (*verificationService) Find(email string, verificationType entity.VerificationDataType) (*entity.Verification, error) {
	return verificationRepo.Find(email, verificationType)
}

func (*verificationService) Update(verification *entity.Verification) (*entity.Verification, error) {
	return verificationRepo.Update(*verification)
}

func (*verificationService) FindAll() ([]entity.Verification, error) {
	return verificationRepo.FindAll()
}

func (*verificationService) Delete(id string) (int64, error) {
	return verificationRepo.Delete(id)
}
