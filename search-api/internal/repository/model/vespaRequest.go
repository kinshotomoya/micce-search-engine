package model

type VespaRequest struct {
	Yql     string `json:"yql"`
	Ranking string `json:"ranking"`
}

func NewVespaRequest(yql string, ranking string) *VespaRequest {
	return &VespaRequest{
		Yql:     yql,
		Ranking: ranking,
	}
}
