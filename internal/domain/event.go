package domain

import "time"

type Event struct {
	ID                                                                        int
	Title, ShortDescription, FullDescription, Location, CoverImageURL, Status string
	StartDate, EndDate, RatingDeadline, CreatedAt, UpdatedAt                  time.Time
	RegistrationDeadline                                                      *time.Time
	MaxParticipants                                                           *int
	ReserveParticipants, SPPoints, FinalReviewsCount                          int
	FinalRating                                                               *float64
	FinalRatingFixedAt                                                        *time.Time
	CreatedByID                                                               int
	ParticipantsCount, ReserveCount                                           int
	RecruitmentStatus                                                         string
}

type EventOrganizer struct {
	ID, EventID, UserID int
	CreatedAt           time.Time
}

type EventRole struct {
	ID, EventID                      int
	Title, Description, Requirements string
	Capacity                         *int
	IsActive                         bool
	SortOrder                        int
}
