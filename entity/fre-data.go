package entity

// FRERequest - request sent to the feature recognition service
type FRERequest struct {
	URL       string `json:"url"`
	CADFileID string `json:"cadfile_id"`
	UserID    string `json:"user_id"`
}

// FREResponse - response after feature recognition is complete
type FREResponse struct {
	UserID       string          `json:"user_id"`
	CADFileID    string          `json:"cadfile_id"`
	BendFeatures []BendFeature   `json:"features"`
	FeatureProps FeatureProperty `json:"feature_props"`
}
