package vespa

import (
	"encoding/json"
	"log"
)

type Document struct {
	Id                 string   `json:"id"`
	Name               *string  `json:"name"`
	KoreaName          *string  `json:"korea_name"`
	Latitude           *float64 `json:"latitude"`
	Longitude          *float64 `json:"longitude"`
	Category           *string  `json:"category"`
	HasInstagramImages bool     `json:"has_instagram_images"`
}

func DecodeDocument(data []byte) Document {

	var document Document
	err := json.Unmarshal(data, &document)
	if err != nil {
		log.Printf("decodeに失敗: %s", data)
	}

	return document
}
