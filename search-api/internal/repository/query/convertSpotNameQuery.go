package query

import "search-api/internal/domain"

func ConvertSpotNameQuery(condition *domain.SearchCondition, synonymKeyword *string) *OrQuery {
	if condition.SpotName == nil {
		return nil
	}

	whereQuery := make([]WhereQuery, 0)
	if synonymKeyword != nil {
		whereQuery = append(whereQuery, NewMatchQuery("name", *synonymKeyword))
		whereQuery = append(whereQuery, NewMatchQuery("korea_name", *synonymKeyword))

	}
	whereQuery = append(whereQuery, NewMatchQuery("name", *condition.SpotName))
	whereQuery = append(whereQuery, NewMatchQuery("korea_name", *condition.SpotName))

	return &OrQuery{
		whereQuery: whereQuery,
	}
}
