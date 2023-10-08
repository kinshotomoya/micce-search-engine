package query

import (
	"fmt"
	"reflect"
	"strings"
)

type WhereQuery interface {
	BuildQuery() string
}

type AndQuery struct {
	whereQuery []WhereQuery
}

func (a AndQuery) BuildQuery() string {
	queries := make([]string, 0, len(a.whereQuery))
	for index := range a.whereQuery {
		if a.whereQuery[index] == nil || reflect.ValueOf(a.whereQuery[index]).IsNil() {
			continue
		}
		queries = append(queries, a.whereQuery[index].BuildQuery())
	}
	return fmt.Sprintf("(%s)", strings.Join(queries, " and "))
}

type OrQuery struct {
	whereQuery []WhereQuery
}

func (a OrQuery) BuildQuery() string {
	queries := make([]string, 0, len(a.whereQuery))
	for index := range a.whereQuery {
		if a.whereQuery[index] == nil || reflect.ValueOf(a.whereQuery[index]).IsNil() {
			continue
		}
		queries = append(queries, a.whereQuery[index].BuildQuery())
	}
	return fmt.Sprintf("(%s)", strings.Join(queries, " or "))
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
	return fmt.Sprintf("%s contains \"%s\"", m.field, m.keyword)
}

type GeoQuery struct {
	field    string
	lat      float64
	lon      float64
	distance string
}

func NewGeoQuery(field string, lat float64, lon float64, distance string) *GeoQuery {
	return &GeoQuery{
		field:    field,
		lat:      lat,
		lon:      lon,
		distance: distance,
	}
}

func (g GeoQuery) BuildQuery() string {
	return fmt.Sprintf("geoLocation(%s, %f, %f, \"%s\")", g.field, g.lat, g.lon, g.distance)
}

type ComparisonOperator int

const (
	Eq ComparisonOperator = iota
	Gt
	GtEq
	Lt
	LtEq
)

func (o ComparisonOperator) ToString() string {
	switch o {
	case Eq:
		return "="
	case Gt:
		return ">"
	case GtEq:
		return ">="
	case Lt:
		return "<"
	case LtEq:
		return "<="
	default:
		return "="
	}
}

type FilterQuery struct {
	field    string
	operator ComparisonOperator
	value    string
}

func (f FilterQuery) BuildQuery() string {
	return fmt.Sprintf("%s %s %s", f.field, f.operator.ToString(), f.value)
}

func NewFilterQuery(field string, op ComparisonOperator, value string) *FilterQuery {
	return &FilterQuery{
		field:    field,
		operator: op,
		value:    value,
	}
}
