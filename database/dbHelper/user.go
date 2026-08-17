package dbHelper

import (
	"Todo/database"
	"Todo/models"
	"Todo/utils"
	"time"

	"github.com/jmoiron/sqlx"
)

func IsUserExists(email string) (bool, error) {
	SQL := `SELECT count(id) > 0 as is_exist
			  FROM users
			  WHERE email = TRIM($1)
			    AND archived_at IS NULL`

	var check bool
	chkErr := database.Todo.Get(&check, SQL, email)
	return check, chkErr
}

func CreateUser(name, email, password string) error {
	SQL := `INSERT INTO users (name, email, password)
			  VALUES (TRIM($1), TRIM($2), $3)`

	_, crtErr := database.Todo.Exec(SQL, name, email, password)
	return crtErr
}

func CreateUserSession(userID string) (string, error) {
	var sessionID string
	SQL := `INSERT INTO user_session(user_id) 
              VALUES ($1) RETURNING id`
	crtErr := database.Todo.Get(&sessionID, SQL, userID)
	return sessionID, crtErr
}

func GetUserID(body models.LoginRequest) (string, error) {
	SQL := `SELECT u.id,
       			   u.password
			  FROM users u
			  WHERE u.email = TRIM($1)
			    AND u.archived_at IS NULL`

	var user models.LoginData
	if getErr := database.Todo.Get(&user, SQL, body.Email); getErr != nil {
		return "", getErr
	}
	if passwordErr := utils.CheckPassword(body.Password, user.PasswordHash); passwordErr != nil {
		return "", passwordErr
	}
	return user.ID, nil
}

func GetUser(userID string) (models.User, error) {
	var user models.User
	SQL := `SELECT id, name, email 
              FROM users 
              WHERE id = $1
                AND archived_at IS NULL`

	getErr := database.Todo.Get(&user, SQL, userID)
	return user, getErr
}

func GetArchivedAt(sessionID string) (*time.Time, error) {
	var archivedAt *time.Time

	SQL := `SELECT archived_at 
              FROM user_session 
              WHERE id = $1`

	getErr := database.Todo.Get(&archivedAt, SQL, sessionID)
	return archivedAt, getErr
}

func DeleteUserSession(sessionID string) error {
	SQL := `UPDATE user_session
			  SET archived_at = NOW()
			  WHERE id = $1
			    AND archived_at IS NULL`

	_, delErr := database.Todo.Exec(SQL, sessionID)
	return delErr
}

func DeleteAllUserSessions(tx *sqlx.Tx, userID string) error {
	SQL := `UPDATE user_session
			  SET archived_at = NOW()
			  WHERE user_id = $1
			    AND archived_at IS NULL`

	_, delErr := tx.Exec(SQL, userID)
	return delErr
}

func DeleteUser(tx *sqlx.Tx, userID string) error {
	SQL := `UPDATE users
			  SET archived_at = NOW()
			  WHERE id = $1
			    AND archived_at IS NULL`

	_, delErr := tx.Exec(SQL, userID)
	return delErr
}
