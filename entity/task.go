package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// ProcessingPlan -
type Task struct {
	ID                         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	TaskID                     primitive.ObjectID `json:"task_id,omitempty" bson:"task_id,omitempty"`
	UserID                     primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty"`
	CADFiles                   []string           `json:"cadfiles" bson:"cadfiles,omitempty" validate:"empty=false"`
	ProcessedCADFiles          []Processed        `json:"processed_cadfiles" bson:"processed_cadfiles,omitempty" validate:"empty=false"`
	Status                     Status             `json:"status" bson:"status" validate:"empty=false"`
	Quantity                   int64              `json:"quantity" bson:"quantity" validate:"empty=false"`
	ProcessingTime             float64            `json:"processing_time" bson:"processing_time" validate:"empty=false"`
	EstimatedManufacturingTime float64            `json:"estimated_manufacturing_time" bson:"estimated_manufacturing_time" validate:"empty=false"`
	TotalCost                  float64            `json:"total_cost" bson:"total_cost" validate:"empty=false"`
	CreatedAt                  int64              `json:"created_at" bson:"created_at" validate:"empty=false"`
}

type ProcessType string
type Status string

const (
	FeatureRecognition ProcessType = "Feature recognition"
	ProcessPlanning    ProcessType = "Process planning"
	Complete           Status      = "Complete"
	Processing         Status      = "Processing"
)

type Processed struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FileName    string             `json:"filename" bson:"filename" validate:"empty=false"`
	ProcessType ProcessType        `json:"process_type" bson:"process_type" validate:"empty=false"`
	Status      Status             `json:"status" bson:"status" validate:"empty=false"`
}
