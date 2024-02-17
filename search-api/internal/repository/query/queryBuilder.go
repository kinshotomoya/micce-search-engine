package query

import (
	"fmt"
	"search-api/internal/domain"
)

type QueryBuilder struct {
	schema string // spot
	fields string // *
}

func NewQueryBuilder(schema string, fields string) *QueryBuilder {
	return &QueryBuilder{
		schema: schema,
		fields: fields,
	}
}

func (b *QueryBuilder) BuildQuery(condition *domain.SearchCondition, synonymKeyword *string) string {
	andQuery := AndQuery{
		whereQuery: []WhereQuery{
			ConvertSpotNameQuery(condition, synonymKeyword),
			ConvertGeoQuery(condition),
			ConvertCategoryQuery(condition),
			ConvertHasInstagramImageQuery(condition),
		},
	}
	whereQuery := andQuery.BuildQuery()

	// NOTE:
	// page:1, limit: 10の場合
	// limit 10, offset 0

	// page: 2, limit: 10の場合
	// limit 20, offset 10

	// page: 3, limit: 10の場合
	// limit 30, offset 20
	limit := *condition.Limit * *condition.Page
	offset := *condition.Limit**condition.Page - *condition.Limit
	return fmt.Sprintf("select %s from %s where %s limit %d offset %d", b.fields, b.schema, whereQuery, limit, offset)
}
