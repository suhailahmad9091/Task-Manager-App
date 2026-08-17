package handlers

import (
	"Todo/database"
	"Todo/database/dbHelper"
	"Todo/middlewares"
	"Todo/models"
	"Todo/utils"
	"net/http"

	"github.com/jmoiron/sqlx"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var body models.RegisterRequest

	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body")
		return
	}

	if valErr := utils.ValidateStruct(body); valErr != nil {
		utils.RespondError(w, http.StatusBadRequest, valErr, "input validation failed")
		return
	}

	exists, existsErr := dbHelper.IsUserExists(body.Email)
	if existsErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, existsErr, "failed to check user existence")
		return
	}
	if exists {
		utils.RespondError(w, http.StatusBadRequest, nil, "user already exists")
		return
	}

	hashedPassword, hasErr := utils.HashPassword(body.Password)
	if hasErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, hasErr, "failed to secure password")
		return
	}

	if saveErr := dbHelper.CreateUser(body.Name, body.Email, hashedPassword); saveErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, saveErr, "failed to save user")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"user created successfully"})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var body models.LoginRequest

	if parseErr := utils.ParseBody(r.Body, &body); parseErr != nil {
		utils.RespondError(w, http.StatusBadRequest, parseErr, "failed to parse request body")
		return
	}

	if valErr := utils.ValidateStruct(body); valErr != nil {
		utils.RespondError(w, http.StatusBadRequest, valErr, "input validation failed")
		return
	}

	userID, userErr := dbHelper.GetUserID(body)
	if userErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, userErr, "failed to find user")
		return
	}

	if userID == "" {
		utils.RespondError(w, http.StatusBadRequest, nil, "user not found")
		return
	}

	sessionID, crtErr := dbHelper.CreateUserSession(userID)
	if crtErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, crtErr, "failed to create user session")
		return
	}

	token, genErr := utils.GenerateJWT(userID, sessionID)
	if genErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, genErr, "failed to generate token")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}{"login successful", token})
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	userCtx := middlewares.UserContext(r)
	userID := userCtx.UserID

	user, getErr := dbHelper.GetUser(userID)
	if getErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, getErr, "failed to get user")
		return
	}

	utils.RespondJSON(w, http.StatusOK, user)
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	userCtx := middlewares.UserContext(r)
	sessionID := userCtx.SessionID

	if delErr := dbHelper.DeleteUserSession(sessionID); delErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, delErr, "failed to delete user session")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"logout successful"})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	userCtx := middlewares.UserContext(r)
	userID := userCtx.UserID

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		if delErr := dbHelper.DeleteUser(tx, userID); delErr != nil {
			return delErr
		}

		if delErr := dbHelper.DeleteAllTodos(tx, userID); delErr != nil {
			return delErr
		}

		return dbHelper.DeleteAllUserSessions(tx, userID)
	})
	if txErr != nil {
		utils.RespondError(w, http.StatusInternalServerError, txErr, "failed to delete user account")
		return
	}

	utils.RespondJSON(w, http.StatusOK, struct {
		Message string `json:"message"`
	}{"account deleted successfully"})
}
