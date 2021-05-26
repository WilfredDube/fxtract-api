package entity

// Tool -
type Tool struct {
	ToolID    string  `json:"tool_id" bson:"tool_id" validate:"empty=false"`
	ToolName  string  `json:"tool_name" bson:"tool_name" validate:"empty=false"`
	Angle     float64 `json:"angle" bson:"angle" validate:"empty=false"`
	Length    float64 `json:"length" bson:"length" validate:"empty=false"`
	MinRadius float64 `json:"min_radius" bson:"min_radius" validate:"empty=false"`
	MaxRadius float64 `json:"max_radius" bson:"max_radius" validate:"empty=false"`
	CreatedAt int64   `json:"created_at" bson:"created_at" validate:"empty=false"`
}
