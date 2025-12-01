package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (api *API) RegisterTasks(r chi.Router) {
	r.Group(func(gr chi.Router) {

	})
}

func (api *API) getTasks(w http.ResponseWriter, r *http.Request) {

}

func (api *API) createTaskHandler(w http.ResponseWriter, r http.Request) {

}
