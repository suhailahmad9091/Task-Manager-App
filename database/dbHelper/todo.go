package dbHelper

import (
	"Todo/database"
	"Todo/models"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// checkRowsAffected turns an update that matched nothing into sql.ErrNoRows,
// so callers can tell "not found" apart from a successful write.
func checkRowsAffected(res sql.Result) error {
	rows, rowsErr := res.RowsAffected()
	if rowsErr != nil {
		return rowsErr
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func IsTodoExists(name, userID string) (bool, error) {
	SQL := `SELECT count(id) > 0 as is_exist
			  FROM todos
			  WHERE name = TRIM($1)     
			    AND user_id = $2        
			    AND archived_at IS NULL`

	var check bool
	chkErr := database.Todo.Get(&check, SQL, name, userID)
	return check, chkErr
}

func CreateTodo(body models.TodoRequest) error {
	SQL := `INSERT INTO todos (name, description, user_id)
			  VALUES (TRIM($1), TRIM($2), $3)`

	_, crtErr := database.Todo.Exec(SQL, body.Name, body.Description, body.UserID)
	return crtErr
}

func GetAllTodos(userID, keyword, completed string) ([]models.Todo, error) {
	SQL := `SELECT id, user_id, name, description, is_completed
				FROM todos
				WHERE user_id = $1
				  AND (
					$2 = '' OR (name ILIKE '%' || $2 || '%' OR description ILIKE '%' || $2 || '%')
					)
				  AND ($3 = '' OR is_completed = CAST($3 AS BOOLEAN))
				  AND archived_at IS NULL`

	todos := make([]models.Todo, 0)
	getErr := database.Todo.Select(&todos, SQL, userID, keyword, completed)
	return todos, getErr
}

// MarkCompleted returns sql.ErrNoRows when the todo does not exist, is already
// archived, or belongs to a different user.
func MarkCompleted(todoID, userID string) error {
	SQL := `UPDATE todos
              SET is_completed = true
              WHERE id = $1
                AND user_id = $2
                AND archived_at IS NULL`

	res, updErr := database.Todo.Exec(SQL, todoID, userID)
	if updErr != nil {
		return updErr
	}
	return checkRowsAffected(res)
}

// DeleteTodo returns sql.ErrNoRows when the todo does not exist, is already
// archived, or belongs to a different user.
func DeleteTodo(todoID, userID string) error {
	SQL := `UPDATE todos
			  SET archived_at = NOW()
			  WHERE id = $1
			    AND user_id = $2
			    AND archived_at IS NULL`

	res, delErr := database.Todo.Exec(SQL, todoID, userID)
	if delErr != nil {
		return delErr
	}
	return checkRowsAffected(res)
}

func DeleteAllTodos(tx *sqlx.Tx, userID string) error {
	SQL := `UPDATE todos
              SET archived_at = NOW()
              WHERE user_id = $1
                AND archived_at IS NULL`

	_, delErr := tx.Exec(SQL, userID)
	return delErr
}
