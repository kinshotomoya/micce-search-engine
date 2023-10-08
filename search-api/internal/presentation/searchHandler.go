package presentation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"search-api/internal/domain"
	"search-api/internal/presentation/model"
	"search-api/internal/repository"
)

type Handler struct {
	vespaRepository *repository.VespaRepository
}

func NewHandler(vespaRepository *repository.VespaRepository) *Handler {
	return &Handler{
		vespaRepository: vespaRepository,
	}
}

func (handler *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	searchCondition, err := domain.NewSearchCondition(body)
	defer body.Close()
	if err != nil {
		responseError(w, 400, err)
		return
	}

	res, err := handler.vespaRepository.Search(searchCondition)
	if err != nil {
		responseError(w, 500, err)
		return
	}

	response := model.NewResponse(res, searchCondition)
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(&response)
	if err != nil {
		responseError(w, 500, err)
		return
	}

	w.Write(buf.Bytes())

}

func responseError(w http.ResponseWriter, status int, err error) {
	switch status {
	case 400:
		var buf bytes.Buffer
		errRes := model.Error400(err)
		json.NewEncoder(&buf).Encode(errRes)
		w.Write(buf.Bytes())
	default:
		var buf bytes.Buffer
		errRes := model.Error500(err)
		json.NewEncoder(&buf).Encode(errRes)
		w.Write(buf.Bytes())
	}

}
