package service

import (
	"errors"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/repository"
)

var (
	processingPlanRepo repository.ProcessingPlanRepository
)

// ProcessingPlanService -
type ProcessingPlanService interface {
	Validate(processingPlan *entity.ProcessingPlan) error
	Create(processingPlan *entity.ProcessingPlan) (*entity.ProcessingPlan, error)
	Update(processingPlan entity.ProcessingPlan) (*entity.ProcessingPlan, error)
	Find(id string) (*entity.ProcessingPlan, error)
	FindAll(processingPlanID string) ([]entity.ProcessingPlan, error)
	Delete(id string) (int64, error)
	CascadeDelete(id string) (int64, error)
}

type processingPlanService struct{}

// NewProcessingPlanService -
func NewProcessingPlanService(dbRepository repository.ProcessingPlanRepository) ProcessingPlanService {
	processingPlanRepo = dbRepository
	return &processingPlanService{}
}

func (*processingPlanService) Validate(processingPlan *entity.ProcessingPlan) error {
	if processingPlan == nil {
		return errors.New("ProcessingPlan is empty")
	}

	return nil
}

func (c *processingPlanService) Create(processingPlan *entity.ProcessingPlan) (*entity.ProcessingPlan, error) {
	if err := c.Validate(processingPlan); err != nil {
		return nil, err
	}

	return processingPlanRepo.Create(processingPlan)
}

func (*processingPlanService) Update(processingPlan entity.ProcessingPlan) (*entity.ProcessingPlan, error) {
	return processingPlanRepo.Update(processingPlan)
}

func (*processingPlanService) Find(id string) (*entity.ProcessingPlan, error) {
	return processingPlanRepo.Find(id)
}

func (*processingPlanService) FindAll(processingPlanID string) ([]entity.ProcessingPlan, error) {
	return processingPlanRepo.FindAll(processingPlanID)
}

func (*processingPlanService) Delete(id string) (int64, error) {
	return processingPlanRepo.Delete(id)
}

func (*processingPlanService) CascadeDelete(processingPlanID string) (int64, error) {
	return processingPlanRepo.CascadeDelete(processingPlanID)
}
