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
	var searchCondition domain.SearchCondition
	body := r.Body
	defer body.Close()
	err := json.NewDecoder(body).Decode(&searchCondition)
	if err != nil {
		var buf bytes.Buffer
		errRes := model.Error500(err)
		json.NewEncoder(&buf).Encode(errRes)
		w.Write(buf.Bytes())
		return
	}

}
