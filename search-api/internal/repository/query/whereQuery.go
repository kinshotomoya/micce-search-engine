package query

import (
	"fmt"
	"strings"
)

type WhereQuery interface {
	BuildQuery() string
}

type AndQuery struct {
	whereQuery []WhereQuery
}

func (a AndQuery) BuildQuery() string {
	queries := make([]string, len(a.whereQuery))
	for index := range a.whereQuery {
		queries[index] = a.whereQuery[index].BuildQuery()
	}
	return fmt.Sprintf("(%s)", strings.Join(queries, " and "))
}

type MatchQuery struct {
	field   string
	keyword string
}

func NewMatchQuery(field string, keyword string) *MatchQuery {
	return &MatchQuery{
		field:   field,
		keyword: keyword,
	}
}

func (m MatchQuery) BuildQuery() string {
	return fmt.Sprintf("%s contains %s", m.field, m.keyword)
}

type GeoQuery struct {
	field string
	lat   float64
	lon   float64
}

func NewGeoQuery(field string, lat float64, lon float64) *GeoQuery {
	return &GeoQuery{
		field: field,
		lat:   lat,
		lon:   lon,
	}
}

func (g GeoQuery) BuildQuery() string {
	return fmt.Sprintf("geoLocation(%s, %f, %f, \"200km\")", g.field, g.lat, g.lon)
}
