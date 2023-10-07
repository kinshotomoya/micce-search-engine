package model

import (
	"bytes"
	"encoding/json"
	"log/slog"
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

func (v *VespaClient) Do(request *VespaRequest) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return
	}

	// TODO: ログレベル設定
	slog.Info("yql: " + request.Yql)

	req, _ := http.NewRequest("POST", v.vespaConfig.Url, &buf)
	res, err := v.httpClient.Do(req)
	if err != nil {
		return
	}
	// TODO: 続き、vespaからのレスポンス処理
}
