package contracts

type FeatureRecognitionStarted struct {
	URL       string `json:"url"`
	CADFileID string `json:"cadfile_id"`
	UserID    string `json:"user_id"`
}

// EventName returns the event's name
func (c *FeatureRecognitionStarted) EventName() string {
	return "featureRecognitionStarted"
}
