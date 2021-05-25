package entity

// Material -
type Material struct {
	Name            string  `json:"name" bson:"name" validate:"empty=false"`
	TensileStrength float64 `json:"tensile_strength" bson:"tensile_strength" validate:"empty=false"`
	KFactor         float64 `json:"k_factor" bson:"k_factor" validate:"empty=false"`
}
