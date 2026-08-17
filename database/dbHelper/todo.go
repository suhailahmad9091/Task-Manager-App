package dbHelper

import (
	"Todo/database"
	"Todo/models"

	"github.com/jmoiron/sqlx"
)

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

func MarkCompleted(todoID, userID string) error {
	SQL := `UPDATE todos
              SET is_completed = true        
              WHERE id = $1                  
                AND user_id = $2             
                AND archived_at IS NULL`

	_, updErr := database.Todo.Exec(SQL, todoID, userID)
	return updErr
}

func DeleteTodo(todoID, userID string) error {
	SQL := `UPDATE todos
			  SET archived_at = NOW()        
			  WHERE id = $1                  
			    AND user_id = $2             
			    AND archived_at IS NULL`

	_, delErr := database.Todo.Exec(SQL, todoID, userID)
	return delErr
}

func DeleteAllTodos(tx *sqlx.Tx, userID string) error {
	SQL := `UPDATE todos
              SET archived_at = NOW()
              WHERE user_id = $1
                AND archived_at IS NULL`

	_, delErr := tx.Exec(SQL, userID)
	return delErr
}
