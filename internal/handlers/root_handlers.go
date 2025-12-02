package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *API) RegisterRoot(r chi.Router) {
	r.Get("/", api.rootHandler)
}

func (api *API) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("wrong url"))
}
