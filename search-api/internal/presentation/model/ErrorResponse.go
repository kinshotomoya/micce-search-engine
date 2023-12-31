package model

type ErrorResponse struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func Error500(err error) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: 500,
		Message:    err.Error(),
	}
}

func Error400(err error) *ErrorResponse {
	return &ErrorResponse{
		StatusCode: 400,
		Message:    err.Error(),
	}
}
