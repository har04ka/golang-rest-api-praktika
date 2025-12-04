package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

	err = utils.ValidateLoginRequest(user.Login, user.Password)
	if err != nil {
		valErr, ok := err.(*utils.ValidationError)
		if ok {
			utils.WriteJSONValidationError(w, valErr.Field, valErr.Message)
		} else {
			utils.WriteJSONError(w, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	var (
		passwordHash string
		id           int64
		isAdmin      bool
		family       string
		name         string
		surname      string
	)
	row := api.Pool.QueryRow(
		r.Context(),
		"select id, password_hash, is_admin, family, name, surname from users where login = $1",
		user.Login,
	)
	err = row.Scan(
		&id,
		&passwordHash,
		&isAdmin,
		&family,
		&name,
		&surname,
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

	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)

	loginResponse := models.LoginResponse{
		Status: "ok",
		Token:  token,
		User: models.UserProfileResponse{
			Id:      int(id),
			Family:  family,
			Name:    name,
			Surname: surname,
			IsAdmin: isAdmin,
		},
	}

	utils.WriteJSON(w, http.StatusOK, loginResponse)
}

func (api *API) aboutMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)

	if !ok {
		utils.WriteJSONError(w, http.StatusUnauthorized, "not_authorized", "trying to get about info without auth")
		return
	}

	var user models.UserProfileResponse
	err := api.Pool.QueryRow(
		r.Context(),
		"select id, family, name, surname, is_admin from users where id = $1",
		userID,
	).Scan(
		&user.Id,
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
	utils.WriteJSON(w, http.StatusOK, user)
}

func (api *API) logoutHandler(w http.ResponseWriter, r *http.Request) {
	var tokenValue string

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenValue = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			utils.ClearSessionCookie(w)
			utils.WriteJSONSuccess(w, http.StatusOK)
			return
		}
		tokenValue = cookie.Value
	}

	if tokenValue == "" {
		utils.ClearSessionCookie(w)
		utils.WriteJSONSuccess(w, http.StatusOK)
		return
	}

	_, err := api.Pool.Exec(
		r.Context(),
		"delete from sessions where token_hash = $1",
		utils.HashTokenHMAC(tokenValue),
	)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "session_delete_failed", "failed to delete session")
		return
	}

	utils.ClearSessionCookie(w)
	utils.WriteJSONSuccess(w, http.StatusOK)
}
