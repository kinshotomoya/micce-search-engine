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

/*
参考：
https://docs.vespa.ai/en/document-v1-api-guide.html#upserts
*/
func (v VespaClient) Upsert(document Document) {
	body, err := createBody(document)
	fmt.Println(body)
	if err != nil {
		log.Print("fatal create request body")
		return
	}
	url := fmt.Sprintf("%s/document/v1/default/spot/docid/%s?create=true", v.Config.Url, document.Id)
	req, _ := http.NewRequest("PUT", url, body)
	req.Header.Add("Content-Type", "application/json")

	res, err := v.Client.Do(req)
	if err != nil {
		log.Print("fatal put document to vespa")
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

	var lat *Assign
	if document.Latitude == nil {
		lat = nil
	} else {
		lat = &Assign{
			Assign: document.Latitude,
		}
	}

	var lon *Assign
	if document.Longitude == nil {
		lon = nil
	} else {
		lon = &Assign{
			Assign: document.Longitude,
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
			Name:      name,
			KoreaName: koreaName,
			Latitude:  lat,
			Longitude: lon,
			Category:  category,
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
