package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"rest-api/internal/middlewares"
	"rest-api/internal/models"
	"rest-api/utils"

	"github.com/go-chi/chi/v5"
)

type userData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (api *API) RegisterAuth(r chi.Router) {
	r.Group(func(gr chi.Router) {
		gr.Use(middlewares.AuthCheck(api.Pool))
		gr.Get("/auth/me", api.aboutMe)
	})
	r.Post("/auth/login", api.loginHandler)
	r.Post("/auth/logout", api.logoutHandler)
}

func (api *API) loginHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
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
		utils.WriteJSONError(w, http.StatusUnauthorized, "invalid_credentials", "user does not exist or password is incorrect")
		return
	}
	err = utils.ComparePassword(user.Password, passwordHash)
	if err != nil {
		utils.WriteJSONError(w, http.StatusUnauthorized, "invalid_credentials", "user does not exist or password is incorrect")
		return
	}

	token, err := utils.GenerateSessionToken()
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "token_generation_failed", "failed to create token")
		return
	}

	tokenHash := utils.HashTokenHMAC(token)
	_, err = api.Pool.Exec(r.Context(), "insert into sessions(user_id, token_hash) values ($1, $2)", id, tokenHash)
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

func (api *API) aboutMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)

	if !ok {
		w.Header().Set("Content-Type", "application/json")
		utils.WriteJSONError(w, http.StatusUnauthorized, "not_authorized", "trying to get about info without auth")
		return
	}

	var user models.UserResponse
	err := api.Pool.QueryRow(
		r.Context(),
		"select family, name, surname, is_admin from users where id = $1",
		userID,
	).Scan(
		&user.Family,
		&user.Name,
		&user.Surname,
		&user.IsAdmin,
	)
	if err != nil {
		fmt.Println("database : ", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to fetch user info")
		return
	}
	jsonWriter := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	err = jsonWriter.Encode(user)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "encode_error", "failed to encode json")
		return
	}
}

func (api *API) logoutHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("session_token")
	if err != nil {
		utils.ClearSessionCookie(w)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	_, err = api.Pool.Exec(
		r.Context(),
		"delete from sessions where token_hash = $1",
		utils.HashTokenHMAC(token.Value),
	)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "session_save_failed", "failed to save session token")
		return
	}

	utils.ClearSessionCookie(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
