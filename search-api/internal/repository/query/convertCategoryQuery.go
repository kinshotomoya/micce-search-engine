package query

import "search-api/internal/domain"

func ConvertCategoryQuery(condition *domain.SearchCondition) *MatchQuery {
	if condition.Category == nil {
		return nil
	}

	return NewMatchQuery("category", *condition.Category)
}
