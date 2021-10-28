package entity

// Tool -
type Tool struct {
	ToolID    string  `json:"tool_id" bson:"tool_id" validate:"empty=false"`
	ToolName  string  `json:"-" bson:"tool_name" validate:"empty=false"`
	Angle     float64 `json:"-" bson:"angle" validate:"empty=false"`
	Length    float64 `json:"-" bson:"length" validate:"empty=false"`
	MinRadius float64 `json:"-" bson:"min_radius" validate:"empty=false"`
	MaxRadius float64 `json:"-" bson:"max_radius" validate:"empty=false"`
	CreatedAt int64   `json:"-" bson:"created_at" validate:"empty=false"`
}
