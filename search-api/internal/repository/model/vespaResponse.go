package model

type VespaResponse struct {
	Root struct {
		Id        string  `json:"id"`
		Relevance float64 `json:"relevance"`
		Fields    struct {
			TotalCount int `json:"totalCount"`
		} `json:"fields"`
		Coverage struct {
			Coverage    int  `json:"coverage"`
			Documents   int  `json:"documents"`
			Full        bool `json:"full"`
			Nodes       int  `json:"nodes"`
			Results     int  `json:"results"`
			ResultsFull int  `json:"resultsFull"`
		} `json:"coverage"`
		Children []struct {
			Id        string  `json:"id"`
			Relevance float64 `json:"relevance"`
			Source    string  `json:"source"`
			Fields    struct {
				Sddocname       string `json:"sddocname"`
				Documentid      string `json:"documentid"`
				SpotId          string `json:"spot_id"`
				Name            string `json:"name"`
				SpotGeoLocation struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"spot_geo_location"`
				Category           string `json:"category"`
				HasInstagramImages bool   `json:"has_instagram_images"`
			} `json:"fields"`
		} `json:"children"`
	} `json:"root"`
}
