package contracts

type FeatureRecognitionStarted struct {
	TaskID    string `json:"task_id" `
	URL       string `json:"url"`
	CADFileID string `json:"cadfile_id"`
	UserID    string `json:"user_id"`
	EventType string `json:"event_type"`
}

// EventName returns the event's name
func (c *FeatureRecognitionStarted) EventName() string {
	return "featureRecognitionStarted"
}
