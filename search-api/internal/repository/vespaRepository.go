package repository

import (
	"search-api/internal/domain"
	"search-api/internal/repository/model"
	"search-api/internal/repository/query"
)

type VespaRepository struct {
	vespaClient *model.VespaClient
}

func NewVespaRepository(client *model.VespaClient) *VespaRepository {
	return &VespaRepository{
		vespaClient: client,
	}
}

func (v *VespaRepository) Search(searchCondition *domain.SearchCondition) (*model.VespaResponse, error) {
	builder := query.NewQueryBuilder("spot", "*")
	yql := builder.BuildQuery(searchCondition)
	request := model.NewVespaRequest(yql, "spot")
	res, err := v.vespaClient.Do(request)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *VespaRepository) Close() {
	v.vespaClient.Close()

}
