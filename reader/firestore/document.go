package firestore

import "cloud.google.com/go/firestore"

type Document struct {
	Id        *string  `json:"id"`
	Name      *string  `json:"name"`
	KoreaName *string  `json:"korea_name"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func CreateDocument(snapShot *firestore.DocumentSnapshot) Document {
	var id *string
	if snapShot.Data()["id"] == nil {
		id = nil
	} else {
		tmp := snapShot.Data()["id"].(string)
		id = &tmp
	}

	var name *string
	if snapShot.Data()["name"] == nil {
		name = nil
	} else {
		tmp := snapShot.Data()["name"].(string)
		name = &tmp
	}

	var koreaName *string
	if snapShot.Data()["koreaName"] == nil {
		koreaName = nil
	} else {
		tmp := snapShot.Data()["koreaName"].(string)
		koreaName = &tmp
	}

	var latitude *float64
	if snapShot.Data()["latitude"] == nil {
		latitude = nil
	} else {
		tmp := snapShot.Data()["latitude"].(float64)
		latitude = &tmp
	}

	var longitude *float64
	if snapShot.Data()["longitude"] == nil {
		longitude = nil
	} else {
		tmp := snapShot.Data()["longitude"].(float64)
		longitude = &tmp
	}

	return Document{
		Id:        id,
		Name:      name,
		KoreaName: koreaName,
		Latitude:  latitude,
		Longitude: longitude,
	}
}
