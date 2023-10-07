package domain

type SearchCondition struct {
	SpotName          *string       `json:"spot_name"`
	Category          *string       `json:"category"`
	Geo               *GeoCondition `json:"geo"`
	HasInstagramImage *bool         `json:"has_instagram_image"`
	Limit             int           `json:"limit"`
	Offset            int           `json:"offset"`
}

type GeoCondition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
