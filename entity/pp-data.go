package entity

// PPRequest - request sent to the process planning service
type PPRequest struct {
	UserID         string `json:"user_id"`
	CADFileID      string `json:"cadfile_id"`
	BendCount      int64  `json:"bend_count"`
	SerializedData string `json:"serialized_data"`
}

// PPResponse - response send back after process planning is complete
type PPResponse struct {
	UserID         string         `json:"user_id"`
	ProcessingPlan ProcessingPlan `json:"pp_plan"`
	ProcessLevel   int            `json:"process_level"`
}
