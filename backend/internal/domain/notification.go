package domain

import "time"

type Notification struct {
	ID, UserID                    int
	Type, Title, Message, LinkURL string
	IsRead                        bool
	CreatedAt                     time.Time
}
