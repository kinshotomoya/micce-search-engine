package model

type PreEventData struct {
	Type   EventType `json:"type"`
	SpotId string    `json:"spot_id"`
}

type EventType string

const (
	INDEX    EventType = "index"
	COMPLETE EventType = "complete"
)
