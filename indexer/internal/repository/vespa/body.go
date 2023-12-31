package vespa

type Body struct {
	Fields Fields `json:"fields"`
}

type Fields struct {
	Id                 Assign  `json:"spot_id"`
	Name               *Assign `json:"name,omitempty"`
	KoreaName          *Assign `json:"korea_name,omitempty"`
	SpotGeoLocation    *Assign `json:"spot_geo_location"`
	Category           *Assign `json:"category,omitempty"`
	HasInstagramImages Assign  `json:"has_instagram_images"`
}

type Assign struct {
	Assign any `json:"assign"`
}
