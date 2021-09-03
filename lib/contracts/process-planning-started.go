package contracts

type ProcessPlanningStarted struct {
	UserID         string `json:"user_id"`
	CADFileID      string `json:"cadfile_id"`
	BendCount      int64  `json:"bend_count"`
	SerializedData string `json:"serialized_data"`
}

// EventName returns the event's name
func (c *ProcessPlanningStarted) EventName() string {
	return "processPlanningStarted"
}
