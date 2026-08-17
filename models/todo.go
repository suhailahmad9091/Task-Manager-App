package models

type TodoRequest struct {
	UserID      string `json:"userId"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type Todo struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	IsCompleted bool   `json:"isCompleted" db:"is_completed"`
	UserID      string `json:"userId" db:"user_id"`
}
