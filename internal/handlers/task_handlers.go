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
	"github.com/jackc/pgx/v5"
)

func (api *API) RegisterTasks(r chi.Router) {
	r.Group(func(gr chi.Router) {
		gr.Use(middlewares.AuthCheck(api.Pool))
		gr.Use(middlewares.AddUserStatus(api.Pool))

		gr.Get("/tasks", api.getTasks)
		gr.Post("/tasks/{id}/complete", api.completeTaskHandler)

		gr.Group(func(admin chi.Router) {
			admin.Use(middlewares.UserStatusCheck(api.Pool))
			admin.Post("/tasks", api.createTaskHandler)
			admin.Post("/tasks/{id}/users", api.bindUserHandler)
		})
	})
}

func (api *API) getTasks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok || userID == 0 {
		utils.WriteJSONError(w, http.StatusUnauthorized, "not_authorized", "you are not authorized")
		return
	}

	isAdmin, _ := r.Context().Value(middlewares.IsAdminKey).(bool)

	var (
		rows pgx.Rows
		err  error
	)
	if isAdmin {

		rows, err = api.Pool.Query(
			r.Context(),
			"select id, title, description, created_at, is_completed from tasks",
		)
	} else {
		rows, err = api.Pool.Query(
			r.Context(),
			`select t.id, t.title, t.description, t.created_at, t.is_completed
			 from tasks t
			 join task_users tu on tu.task_id = t.id
			 where tu.user_id = $1`,
			userID,
		)
	}

	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to fetch tasks")
		return
	}
	defer rows.Close()

	tasks := []models.Task{}
	for rows.Next() {
		var task models.Task
		err := rows.Scan(
			&task.Id,
			&task.Title,
			&task.Description,
			&task.CreatedAt,
			&task.IsCompleted,
		)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to scan task row")
			return
		}
		tasks = append(tasks, task)
	}

	utils.WriteJSON(w, http.StatusOK, tasks)
}

func (api *API) createTaskHandler(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_request", "failed to read body")
		return
	}

	var task models.TaskRequest
	err = json.Unmarshal(body, &task)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}

	err = utils.ValidateTaskRequest(task.Title, task.Description)
	if err != nil {
		valErr, ok := err.(*utils.ValidationError)
		if ok {
			utils.WriteJSONValidationError(w, valErr.Field, valErr.Message)
		} else {
			utils.WriteJSONError(w, http.StatusBadRequest, "validation_error", err.Error())
		}
		return
	}

	_, err = api.Pool.Exec(
		r.Context(),
		"insert into tasks(title, description, is_completed) values ($1, $2, $3)",
		task.Title, task.Description, task.Is_completed,
	)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "task_creation_failed", "failed to create task")
		return
	}

	utils.WriteJSONSuccess(w, http.StatusCreated)

}

func (api *API) bindUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	taskId, err := strconv.Atoi(idStr)
	if err != nil || taskId <= 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_task_id", "task id must be a positive integer")
		return
	}

	var req struct {
		UserIds []int `json:"user_ids"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_request", "failed to read body")
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_json", "request body is not valid JSON")
		return
	}

	if len(req.UserIds) == 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "validation_error", "user_ids array cannot be empty")
		return
	}

	var taskExists bool
	err = api.Pool.QueryRow(
		r.Context(),
		"select exists(select 1 from tasks where id = $1)",
		taskId,
	).Scan(&taskExists)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to check task existence")
		return
	}
	if !taskExists {
		utils.WriteJSONError(w, http.StatusNotFound, "not_found", "task with this id does not exist")
		return
	}

	for _, userId := range req.UserIds {
		if userId <= 0 {
			utils.WriteJSONError(w, http.StatusBadRequest, "invalid_user_id", "user id must be a positive integer")
			return
		}

		var userExists bool
		err = api.Pool.QueryRow(
			r.Context(),
			"select exists(select 1 from users where id = $1)",
			userId,
		).Scan(&userExists)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to check user existence")
			return
		}
		if !userExists {
			utils.WriteJSONError(w, http.StatusNotFound, "not_found", "user with id does not exist")
			return
		}

		var alreadyBound bool
		err = api.Pool.QueryRow(
			r.Context(),
			"select exists(select 1 from task_users where task_id = $1 and user_id = $2)",
			taskId, userId,
		).Scan(&alreadyBound)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to check existing binding")
			return
		}
		if alreadyBound {
			continue
		}

		_, err = api.Pool.Exec(
			r.Context(),
			"insert into task_users(task_id, user_id) values ($1, $2)",
			taskId, userId,
		)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to bind user")
			return
		}
	}

	utils.WriteJSONSuccess(w, http.StatusOK)
}

func (api *API) completeTaskHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(int64)
	if !ok || userID == 0 {
		utils.WriteJSONError(w, http.StatusUnauthorized, "not_authorized", "you are not authorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	taskID, err := strconv.Atoi(idStr)
	if err != nil || taskID <= 0 {
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid_task_id", "task id must be a positive integer")
		return
	}

	isAdmin, _ := r.Context().Value(middlewares.IsAdminKey).(bool)

	if !isAdmin {

		var hasAccess bool
		err = api.Pool.QueryRow(
			r.Context(),
			"select exists(select 1 from task_users where task_id = $1 and user_id = $2)",
			taskID,
			userID,
		).Scan(&hasAccess)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to check task access")
			return
		}

		if !hasAccess {
			utils.WriteJSONError(w, http.StatusForbidden, "forbidden", "you do not have access to this task")
			return
		}
	}

	_, err = api.Pool.Exec(
		r.Context(),
		"update tasks set is_completed = true where id = $1",
		taskID,
	)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to complete task")
		return
	}

	utils.WriteJSONSuccess(w, http.StatusOK)
}
