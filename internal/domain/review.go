package domain

import "time"

type Review struct {
	ID, EventID, UserID, Rating int
	Comment                     string
	CreatedAt, UpdatedAt        time.Time
}
