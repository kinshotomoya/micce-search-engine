package model

import (
	"search-api/internal/domain"
	"search-api/internal/repository/model"
)

type Response struct {
	TotalHits int      `json:"total_hits"`
	LastPage  bool     `json:"last_page"`
	SpotIds   []string `json:"spot_ids"`
}

func NewResponse(vespaResponse *model.VespaResponse, condition *domain.SearchCondition) *Response {
	spotIds := make([]string, len(vespaResponse.Root.Children))
	for index := range vespaResponse.Root.Children {
		spotIds[index] = vespaResponse.Root.Children[index].Fields.SpotId
	}

	response := &Response{
		SpotIds:   spotIds,
		TotalHits: vespaResponse.Root.Fields.TotalCount,
	}
	if isLastPage(vespaResponse.Root.Fields.TotalCount, *condition.Limit, *condition.Page) {
		response.LastPage = true
	} else {
		response.LastPage = false
	}

	return response

}

// 最後のページかどうか
// totalhits = 34
// limit 10 page 1 ならまだある
// limit 10 page 2 ならまだある
// limit 10 page 3 ならまだある
// limit 10 page 4 なら最後のページ
func isLastPage(totalCounts int, limit int, page int) bool {
	return (totalCounts - limit*page) <= 0
}
