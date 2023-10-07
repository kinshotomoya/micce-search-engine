package model

import (
	"net/http"
)

type VespaClient struct {
	httpClient  *http.Client
	vespaConfig *VespaConfig
}

func NewVespaClient(client *http.Client, vespaUrl string) *VespaClient {
	return &VespaClient{
		httpClient: client,
		vespaConfig: &VespaConfig{
			Url:     vespaUrl,
			Timeout: 1000,
		},
	}
}
