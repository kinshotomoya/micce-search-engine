package domain

import (
	"encoding/json"
	"errors"
	"io"
)

type SearchCondition struct {
	SpotName          *string       `json:"spot_name"`
	Category          *string       `json:"category"`
	Geo               *GeoCondition `json:"geo"`
	HasInstagramImage *bool         `json:"has_instagram_image"`
	Limit             *int          `json:"limit"`
	Page              *int          `json:"page"`
}

type GeoCondition struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewSearchCondition(requestBody io.ReadCloser) (*SearchCondition, error) {
	var searchCondition SearchCondition
	err := json.NewDecoder(requestBody).Decode(&searchCondition)
	if err != nil {
		return nil, err
	}

	var errs []error

	// バリデーション
	if searchCondition.Limit == nil {
		errs = append(errs, errors.New("limit is required"))
	}

	if searchCondition.Page == nil {
		errs = append(errs, errors.New("page is required"))
	}

	err = errors.Join(errs...)
	if err != nil {
		return nil, err
	}

	return &searchCondition, nil

}
