package model

import "time"

type CustomTimeInterface interface {
	DatetimeNow() string
}

type CustomTime struct {
	zone *time.Location
}

func NewCustomTime() (*CustomTime, error) {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return nil, err
	}
	return &CustomTime{
		zone: location,
	}, nil

}

func (c *CustomTime) DatetimeNow() string {
	return time.Now().In(c.zone).Format(time.DateTime)
}
