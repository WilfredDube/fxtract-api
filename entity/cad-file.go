package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// CADFile -
type CADFile struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ProjectID    primitive.ObjectID `json:"project_id" bson:"project_id,omitempty" validate:"empty=false"`
	FileName     string             `json:"filename" bson:"filename" validate:"empty=false"`
	StepURL      string             `json:"step_url" bson:"step_url" validate:"empty=false"`
	ObjpURL      string             `json:"obj_url" bson:"obj_url" validate:"empty=false"`
	Material     string             `json:"material_id" bson:"material_id" validate:"empty=false"`
	Filesize     int64              `json:"filesize" bson:"filesize" validate:"empty=false"`
	FeatureProps FeatureProperty    `json:"feature_props" bson:"feature_props" validate:"empty=false"`
	BendFeatures []BendFeature      `json:"bend_features" bson:"bend_features" validate:"empty=false"`
	CreatedAt    int64              `json:"created_at" bson:"created_at" validate:"empty=false"`
}

// FeatureProperty -
type FeatureProperty struct {
	SerialData   string  `json:"serial_data" bson:"serial_data" validate:"empty=false"`
	Thickness    float64 `json:"thickness" bson:"thickness" validate:"empty=false"`
	BendingForce float64 `json:"bending_force" bson:"bending_force" validate:"empty=false"`
	ProcessLevel int     `json:"process_level" bson:"process_level" validate:"empty=false"`
	FRETime      int     `json:"fre_time" bson:"fre_time" validate:"empty=false"`
	BendCount    int     `json:"bend_count" bson:"bend_count" validate:"empty=false"`
}

// BendFeature -
type BendFeature struct {
	BendID       int64   `json:"bend_id" bson:"bend_id" validate:"empty=false"`
	FirstFaceID  int64   `json:"first_face_id" bson:"first_face_id" validate:"empty=false"`
	SecondFaceID int64   `json:"second_face_id" bson:"second_face_id" validate:"empty=false"`
	Angle        float64 `json:"angle" bson:"angle" validate:"empty=false"`
	Length       float64 `json:"length" bson:"length" validate:"empty=false"`
	Radius       float64 `json:"radius" bson:"radius" validate:"empty=false"`
	Direction    float64 `json:"direction" bson:"direction" validate:"empty=false"`
	ToolID       string  `json:"tool_id" bson:"tool_id" validate:"empty=false"`
	CreatedAt    int64   `json:"created_at" bson:"created_at" validate:"empty=false"`
}
