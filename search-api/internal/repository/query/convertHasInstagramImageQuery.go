package query

import (
	"fmt"
	"search-api/internal/domain"
)

func ConvertHasInstagramImageQuery(condition *domain.SearchCondition) *FilterQuery {
	if condition.HasInstagramImage == nil {
		return nil
	}

	return NewFilterQuery("has_instagram_images", Eq, fmt.Sprintf("%t", *condition.HasInstagramImage))
}
