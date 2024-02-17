package repository

import (
	"search-api/internal/domain"
	"search-api/internal/repository/model"
	"search-api/internal/repository/query"
)

type VespaRepository struct {
	vespaClient     *model.VespaClient
	bboltRepository *BboltRepository
}

func NewVespaRepository(client *model.VespaClient, bbolt *BboltRepository) *VespaRepository {
	return &VespaRepository{
		vespaClient:     client,
		bboltRepository: bbolt,
	}
}

func (v *VespaRepository) Search(searchCondition *domain.SearchCondition) (*model.VespaResponse, error) {
	var synonymKeyword *string
	if searchCondition.SpotName != nil {
		b := v.bboltRepository.GetValue([]byte(*searchCondition.SpotName))
		if b != nil {
			str := string(b)
			synonymKeyword = &str
		}
	}

	builder := query.NewQueryBuilder("spot", "*")
	yql := builder.BuildQuery(searchCondition, synonymKeyword)
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
