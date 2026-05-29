package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"rest-api/internal/middlewares"
	"rest-api/internal/models"
	"rest-api/utils"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (api *API) RegisterUserMethods(r chi.Router) {
	r.Group(func(gr chi.Router) {
		gr.Use(middlewares.AuthCheck(api.Pool))
		gr.Use(middlewares.AddUserStatus(api.Pool))
		gr.Get("/users", api.getUsers)
		gr.Get("/users/me/completed", api.getMyCompletedCount)
		gr.Get("/users/{id}", api.getUser)
	})
	r.Group(func(admin chi.Router) {
		// require authentication first so UserStatusCheck can read UserIDKey from context
		admin.Use(middlewares.AuthCheck(api.Pool))
		admin.Use(middlewares.AddUserStatus(api.Pool))
		admin.Use(middlewares.UserStatusCheck(api.Pool))
		admin.Post("/users", api.createUser)
	})
}

func (api *API) getMyCompletedCount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok || userID == 0 {
		utils.WriteJSONError(w, http.StatusUnauthorized, "not_authorized", "you are not authorized")
		return
	}

	var count int64
	err := api.Pool.QueryRow(
		r.Context(),
		`select count(*) from tasks t join task_users tu on tu.task_id = t.id where tu.user_id = $1 and t.is_completed = true`,
		userID,
	).Scan(&count)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to fetch completed count")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int64{"completed_tasks": count})
}

func (api *API) getUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := api.Pool.Query(
		r.Context(),
		"select id, family, name, surname from users",
	)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to fetch users")
		return
	}
	defer rows.Close()

	users := []models.UserPublicResponse{}
	for rows.Next() {
		user := models.UserPublicResponse{}
		err := rows.Scan(
			&user.Id,
			&user.Family,
			&user.Name,
			&user.Surname,
		)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to scan user row")
			return
		}
		users = append(users, user)
	}

	utils.WriteJSON(w, http.StatusOK, users)
}

func (api *API) createUser(w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_request", "failed to read body")
		return
	}

	var user models.UserRequest
	err = json.Unmarshal(bytes, &user)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}

	err = utils.ValidateUserRequest(user.Login, user.Family, user.Name, user.Surname, user.Password)
	if err != nil {
		valErr, ok := err.(*utils.ValidationError)
		if ok {
			utils.WriteJSONValidationError(w, valErr.Field, valErr.Message)
		} else {
			utils.WriteJSONError(w, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "hash_error", "failed to hash password")
		return
	}

	_, err = api.Pool.Exec(
		r.Context(),
		"insert into users (login, family, name, surname, password_hash, is_admin) values ($1, $2, $3, $4, $5, $6)",
		user.Login, user.Family, user.Name, user.Surname, hash, false,
	)

	if err != nil {
		utils.WriteJSONError(w, http.StatusConflict, "user_is_exist", "user with this login is already exist")
		return
	}
	utils.WriteJSONSuccess(w, http.StatusCreated)
}

func (api *API) getUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_id", "user id is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_id", "user id must be a positive integer")
		return
	}

	user := models.UserPublicResponse{}
	err = api.Pool.QueryRow(
		r.Context(),
		"select id, family, name, surname from users where id = $1",
		id,
	).Scan(
		&user.Id,
		&user.Family,
		&user.Name,
		&user.Surname,
	)
	if err != nil {
		utils.WriteJSONError(w, http.StatusNotFound, "not_found", "user with this id does not exist")
		return
	}
	utils.WriteJSON(w, http.StatusOK, user)
}
