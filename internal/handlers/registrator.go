package handlers

import (
	"github.com/go-chi/chi/v5"
)

func (api *API) RegisterAll(c chi.Router) {
	api.RegisterRoot(c)
	api.RegisterUserMethods(c)
	api.RegisterAuth(c)
	api.registerTasks(c)
}
