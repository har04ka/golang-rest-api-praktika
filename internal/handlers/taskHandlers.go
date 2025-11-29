package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *API) registerTasks(r chi.Router) {
	r.Get("/tasks", api.getTasks)
}

func (api *API) getTasks(w http.ResponseWriter, r *http.Request) {

}
