package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"rest-api/internal/models"
	"rest-api/utils"

	"github.com/go-chi/chi/v5"
)

func (api *API) RegisterUserMethods(r chi.Router) {
	r.Get("/users", api.getUsers)
	r.Get("/users/{id}", api.getUser)
	r.Post("/users", api.createUser)
}

func (api *API) getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := api.Pool.Query(
		api.Ctx,
		"select id, family, name, surname, is_admin, created_at, updated_at from users",
	)
	if err != nil {
		panic(err)
	}
	users := []models.UserResponse{}
	for rows.Next() {
		user := models.UserResponse{}
		err := rows.Scan(
			&user.Id,
			&user.Family,
			&user.Name,
			&user.Surname,
			&user.IsAdmin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		users = append(users, user)

	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(users)

}

func (api *API) createUser(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var user models.UserRequest
	err = json.Unmarshal(bytes, &user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	token, err := utils.GenerateSessionToken()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = api.Pool.Exec(
		api.Ctx,
		"insert into users (family, name, surname, password_hash, is_admin, session_token) values ($1, $2, $3, $4, $5, $6)",
		user.Family, user.Name, user.Surname, hash, false, token,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))
}

func (api *API) getUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	user := models.UserResponse{}
	err := api.Pool.QueryRow(
		api.Ctx,
		"select id, family, name, surname, is_admin, created_at, updated_at from users where id = $1",
		id,
	).Scan(
		&user.Id,
		&user.Family,
		&user.Name,
		&user.Surname,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
