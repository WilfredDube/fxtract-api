package contracts

import (
	"github.com/WilfredDube/fxtract-backend/entity"
)

type FeatureRecognitionComplete struct {
	UserID       string                 `json:"user_id"`
	CADFileID    string                 `json:"cadfile_id"`
	TaskID       string                 `json:"task_id" `
	BendFeatures []entity.BendFeature   `json:"features"`
	FeatureProps entity.FeatureProperty `json:"feature_props"`
	EventType    string                 `json:"event_type"`
}

// EventName returns the event's name
func (c *FeatureRecognitionComplete) EventName() string {
	return "featureRecognitionComplete"
}
