package contracts

import (
	"github.com/WilfredDube/fxtract-backend/entity"
)

type ProcessPlanningComplete struct {
	UserID         string                `json:"user_id"`
	CADFileID      string                `json:"cadfile_id"`
	TaskID         string                `json:"task_id" `
	ProcessingPlan entity.ProcessingPlan `json:"pp_plan"`
	ProcessLevel   int                   `json:"process_level"`
	EventType      string                `json:"event_type"`
}

// EventName returns the event's name
func (c *ProcessPlanningComplete) EventName() string {
	return "processPlanningComplete"
}
