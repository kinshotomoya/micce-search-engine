package vespa

type Body struct {
	Fields Fields `json:"fields"`
}

type Fields struct {
	Id                 Assign  `json:"id"`
	Name               *Assign `json:"name,omitempty"`
	KoreaName          *Assign `json:"korea_name,omitempty"`
	Latitude           *Assign `json:"latitude,omitempty"`
	Longitude          *Assign `json:"longitude,omitempty"`
	Category           *Assign `json:"category,omitempty"`
	HasInstagramImages Assign  `json:"has_instagram_images"`
}

type Assign struct {
	Assign any `json:"assign"`
}
