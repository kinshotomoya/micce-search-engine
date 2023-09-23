package vespa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type VespaClient struct {
	Client *http.Client
	Config *VespaConfig
}

func (v VespaClient) Close() {
	log.Println("vespa client closed")
	v.Client.CloseIdleConnections()
}

/*
参考：
https://docs.vespa.ai/en/document-v1-api-guide.html#upserts
*/
func (v VespaClient) Upsert(document Document) {
	body, err := createBody(document)
	fmt.Println(body)
	if err != nil {
		log.Printf("fatal create request body: %s", err.Error())
		return
	}
	url := fmt.Sprintf("%s/document/v1/default/spot/docid/%s?create=true", v.Config.Url, document.Id)
	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Add("Content-Type", "application/json")

	res, err := v.Client.Do(req)
	if err != nil {
		log.Printf("fatal put document to vespa: %s", err.Error())
		return
	}
	defer res.Body.Close()

	bytt, _ := io.ReadAll(res.Body)
	fmt.Println(bytt)

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
