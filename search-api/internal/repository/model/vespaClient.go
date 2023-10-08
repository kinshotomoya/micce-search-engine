package model

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (v *VespaClient) Do(request *VespaRequest) (*VespaResponse, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return nil, err
	}

	// TODO: ログレベル設定
	slog.Info("yql: " + request.Yql)

	req, _ := http.NewRequest("POST", v.vespaConfig.Url+"/search/", &buf)
	req.Header.Add("Content-Type", "application/json")
	res, err := v.httpClient.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	defer res.Body.Close()
	var vespaResponse VespaResponse
	err = json.NewDecoder(res.Body).Decode(&vespaResponse)
	if err != nil {
		return nil, err
	}
	return &vespaResponse, nil
}
