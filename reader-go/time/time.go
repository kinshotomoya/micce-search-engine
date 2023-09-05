package time

import "time"

type Time struct {
	Zone *time.Location
}

func NewTime() *Time {
	zone, _ := time.LoadLocation("Asia/Tokyo")
	return &Time{
		Zone: zone,
	}
}

func (t *Time) Now() time.Time {
	return time.Now().In(t.Zone)
}

func (t *Time) BeforeOneHour() time.Time {
	return time.Now().In(t.Zone).Add(-60 * time.Minute)
}

func (t *Time) BeforeOneYear() time.Time {
	return time.Now().In(t.Zone).AddDate(-1, 0, 0)
}
