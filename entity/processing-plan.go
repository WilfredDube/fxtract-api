package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// ProcessingPlan -
type ProcessingPlan struct {
	ID                         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	CADFileID                  primitive.ObjectID `json:"cadfile_id" bson:"cadfile_id,omitempty" validate:"empty=false"`
	Moderator                  string             `json:"moderator" bson:"moderator" validate:"empty=false"`
	PartNo                     string             `json:"part_no" bson:"part_no" validate:"empty=false"`
	Rotations                  int64              `json:"rotations" bson:"rotations" validate:"empty=false"`
	Flips                      int64              `json:"flips" bson:"flips" validate:"empty=false"`
	Tools                      int64              `json:"tools" bson:"tools" validate:"empty=false"`
	Modules                    int64              `json:"modules" bson:"modules" validate:"empty=false"`
	Quantity                   int64              `json:"quantity" bson:"quantity" validate:"empty=false"`
	ProcessingTime             float64            `json:"processing_time" bson:"processing_time" validate:"empty=false"`
	EstimatedManufacturingTime float64            `json:"estimated_manufacturing_time" bson:"estimated_manufacturing_time" validate:"empty=false"`
	TotalToolDistance          float64            `json:"total_tool_distance" bson:"total_tool_distance" validate:"empty=false"`
	BendingSequences           []BendingSequence  `json:"bend_sequences" bson:"bend_sequences" validate:"empty=false"`
	CreatedAt                  int64              `json:"created_at" bson:"created_at" validate:"empty=false"`
}

// BendingSequence -
type BendingSequence struct {
	// ProcessNo int64 `json:"process_no" bson:"process_no" validate:"empty=false"`
	BendID int64 `json:"bend_id" bson:"bend_id" validate:"empty=false"`
}
