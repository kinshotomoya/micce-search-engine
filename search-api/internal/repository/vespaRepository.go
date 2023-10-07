package repository

import (
	"search-api/internal/domain"
	"search-api/internal/repository/model"
)

type VespaRepository struct {
	vespaClient *model.VespaClient
}

func NewVespaRepository(client *model.VespaClient) *VespaRepository {
	return &VespaRepository{
		vespaClient: client,
	}
}

func (v VespaRepository) Search(searchCondition domain.SearchCondition) {

}
