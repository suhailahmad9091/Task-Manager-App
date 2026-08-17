package handlers

import (
	"Todo/database"
	"Todo/database/dbHelper"
	"Todo/middlewares"
	"Todo/models"
	"Todo/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	var body models.TodoRequest
	userCtx := middlewares.UserContext(r)
	body.UserID = userCtx.UserID

	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body")
		return
	}

	if valErr := utils.ValidateStruct(body); valErr != nil {
		utils.RespondError(w, http.StatusBadRequest, valErr, "input validation failed")
		return
	}

	exists, existsErr := dbHelper.IsTodoExists(body.Name, body.UserID)
	if existsErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existsErr, "failed to check todo existence")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusBadRequest, nil, "todo already exists")
		return
	}

	if saveErr := dbHelper.CreateTodo(body); saveErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, saveErr, "failed to create todo")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"todo created successfully"})
}

func GetAllTodos(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	completed := r.URL.Query().Get("completed")

	userCtx := middlewares.UserContext(r)
	userID := userCtx.UserID

	todos, getErr := dbHelper.GetAllTodos(userID, keyword, completed)
	if getErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, getErr, "failed to get todos")
		return
	}

	utils.RespondJSON(w, http.StatusOK, todos)
}

func MarkCompleted(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "todoId")

	userCtx := middlewares.UserContext(r)
	userID := userCtx.UserID

	updErr := dbHelper.MarkCompleted(todoID, userID)
	if updErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, updErr, "failed to mark todo completed")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"todo marked completed successfully"})
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	todoID := chi.URLParam(r, "todoId")

	userCtx := middlewares.UserContext(r)
	userID := userCtx.UserID

	delErr := dbHelper.DeleteTodo(todoID, userID)
	if delErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, delErr, "failed to delete todo")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"todo deleted successfully"})
}

func DeleteAllTodos(w http.ResponseWriter, r *http.Request) {
	userCtx := middlewares.UserContext(r)
	userID := userCtx.UserID

	delErr := database.Tx(func(tx *sqlx.Tx) error {
		return dbHelper.DeleteAllTodos(tx, userID)
	})
	if delErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, delErr, "failed to delete todos")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"all todos deleted successfully"})
}
