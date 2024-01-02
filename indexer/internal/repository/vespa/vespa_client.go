package vespa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"indexer/internal"
	"io"
	"log"
	"net/http"
)

type VespaClient struct {
	Client *http.Client
	Config *VespaConfig
}

func (v *VespaClient) Close() {
	log.Println("vespa client closed")
	v.Client.CloseIdleConnections()
}

/*
参考：
https://docs.vespa.ai/en/document-v1-api-guide.html#upserts
*/

func (v *VespaClient) Upsert(document Document) error {
	body, err := createBody(document)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal create request body: %s", err.Error()))
		return err
	}
	url := fmt.Sprintf("%s/document/v1/default/spot/docid/%s?create=true", v.Config.Url, document.Id)
	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Add("Content-Type", "application/json")

	res, err := v.Client.Do(req)
	if err != nil {
		internal.Logger.Error(fmt.Sprintf("fatal put document to vespa: %s", err.Error()))
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 && res.StatusCode != 201 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}
	return nil

}

func createBody(document Document) (io.Reader, error) {
	var name *Assign
	if document.Name == nil {
		name = nil
	} else {
		name = &Assign{
			Assign: document.Name,
		}
	}

	var koreaName *Assign
	if document.KoreaName == nil {
		koreaName = nil
	} else {
		koreaName = &Assign{
			Assign: document.KoreaName,
		}
	}

	var geoLocation *Assign
	if document.Latitude == nil || document.Longitude == nil {
		geoLocation = nil
	} else {
		geoLocation = &Assign{
			Assign: map[string]float64{
				"lat": *document.Latitude,
				"lng": *document.Longitude,
			},
		}
	}

	var category *Assign
	if document.Category == nil {
		category = nil
	} else {
		category = &Assign{
			Assign: document.Category,
		}
	}

	body := &Body{
		Fields: Fields{
			Id: Assign{
				Assign: document.Id,
			},
			Name:            name,
			KoreaName:       koreaName,
			SpotGeoLocation: geoLocation,
			Category:        category,
			HasInstagramImages: Assign{
				Assign: document.HasInstagramImages,
			},
		},
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		return nil, err
	}
	return &buf, nil

}
