package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"rest-api/utils"

	"github.com/go-chi/chi/v5"
)

type userData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (api *API) RegisterAuth(r chi.Router) {
	r.Post("/auth/login", api.loginHandler)
}

func (api *API) loginHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_request", "failed to read body")
		return
	}
	var user userData
	err = json.Unmarshal(body, &user)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}

	var (
		passwordHash string
		id           uint
	)
	row := api.Pool.QueryRow(
		r.Context(),
		"select id, password_hash from users where login = $1",
		user.Login,
	)
	err = row.Scan(
		&id,
		&passwordHash,
	)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, "user_not_found", "user does not exist")
		return
	}
	err = utils.ComparePassword(user.Password, passwordHash)
	if err != nil {
		utils.WriteJSONError(w, http.StatusUnauthorized, "invalid_password", "password is incorrect")
		return
	}

	token, err := utils.GenerateSessionToken()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "token_generation_failed", "failed to create token")
		return
	}

	err = api.Pool.QueryRow(
		r.Context(),
		"select  from users where login = $1",
		user.Login,
	).Scan()
	if err != nil {
		utils.WriteJSONError(w, http.StatusUnauthorized, "invalid_password", "password is incorrect")
		return
	}

	_, err = api.Pool.Exec(r.Context(), "insert into sessions(user_id, token) values ($1, $2)", id, token)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "session_save_failed", "failed to save session token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
