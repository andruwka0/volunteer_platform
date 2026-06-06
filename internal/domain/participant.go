package domain

import "time"

type EventParticipant struct {
	ID, EventID, UserID           int
	Status                        string
	DesiredRoleID, AssignedRoleID *int
	CreatedAt, UpdatedAt          time.Time
}

type EventParticipantRoleChoice struct {
	ID, ParticipantID, RoleID int
	CreatedAt                 time.Time
}
