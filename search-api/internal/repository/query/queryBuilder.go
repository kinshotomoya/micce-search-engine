package query

import "search-api/internal/domain"

type Builder struct {
	schema  string // spot
	ranking string // rank profile
}

func (b Builder) Build(condition domain.SearchCondition) *WhereQuery {
	// TODO: 続き
}
