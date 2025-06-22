package domain

import "time"

type Person struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" binding:"required, min=1, max=50"`
	Surname     string    `json:"surname" db:"surname" binding:"required, min=1, max=100"`
	Patronymic  *string   `json:"patronymic,omitempty" db:"patronymic" binding:"omitempty, min=1, max=100"`
	Age         int       `json:"age,omitempty" db:"age" binding:"omitempty, min=0, max=120"`
	Gender      string    `json:"gender,omitempty" db:"gender" binding:"omitempty,oneof=male female other"`
	Nationality string    `json:"nationality,omitempty" db:"nationality" binding:"omitempty,min=2,max=100"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}
