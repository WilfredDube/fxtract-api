package entity

// Material -
type Material struct {
	Name            string  `json:"name" bson:"name" validate:"empty=false"`
	TensileStrength float64 `json:"-" bson:"tensile_strength" validate:"empty=false"`
	KFactor         float64 `json:"-" bson:"k_factor" validate:"empty=false"`
	CreatedAt       int64   `json:"-" bson:"created_at" validate:"empty=false"`
}
