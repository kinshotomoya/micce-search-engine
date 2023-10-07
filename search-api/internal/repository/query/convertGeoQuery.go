package query

import "search-api/internal/domain"

func ConvertGeoQuery(condition *domain.SearchCondition) *GeoQuery {
	if condition.Geo == nil {
		return nil
	}

	return NewGeoQuery("spot_geo_location", condition.Geo.Latitude, condition.Geo.Longitude, "200km")
}
