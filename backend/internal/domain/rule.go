package domain

import "time"

type Rule struct {
	ID                   int
	Title, Content       string
	OrderIndex           int
	CreatedAt, UpdatedAt time.Time
}
