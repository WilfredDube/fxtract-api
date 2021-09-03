package contracts

import "github.com/WilfredDube/fxtract-backend/entity"

type FeatureRecognitionComplete struct {
	UserID       string                 `json:"user_id"`
	CADFileID    string                 `json:"cadfile_id"`
	BendFeatures []entity.BendFeature   `json:"features"`
	FeatureProps entity.FeatureProperty `json:"feature_props"`
}

// EventName returns the event's name
func (c *FeatureRecognitionComplete) EventName() string {
	return "featureRecognitionComplete"
}
