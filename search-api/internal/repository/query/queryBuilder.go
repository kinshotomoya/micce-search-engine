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

func (b *QueryBuilder) BuildQuery(condition *domain.SearchCondition) string {
	andQuery := AndQuery{
		whereQuery: []WhereQuery{
			ConvertSpotNameQuery(condition),
			ConvertGeoQuery(condition),
			ConvertCategoryQuery(condition),
			ConvertHasInstagramImageQuery(condition),
		},
	}
	whereQuery := andQuery.BuildQuery()
	return fmt.Sprintf("select %s from %s where %s %d %d", b.fields, b.schema, whereQuery, condition.Limit, condition.Offset)
}
