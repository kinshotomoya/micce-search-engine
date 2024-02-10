package presentation

import "net/http"

func (handler *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
