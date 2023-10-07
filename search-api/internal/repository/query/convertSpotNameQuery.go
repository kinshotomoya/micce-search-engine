package query

import "search-api/internal/domain"

func ConvertSpotNameQuery(condition *domain.SearchCondition) *OrQuery {
	if condition.SpotName == nil {
		return nil
	}

	return &OrQuery{
		whereQuery: []WhereQuery{
			NewMatchQuery("name", *condition.SpotName),
			NewMatchQuery("korea_name", *condition.SpotName),
		},
	}
}
