package worker

import (
	"context"
	"log"
	"time"

	"github.com/andruwka0/volunteer_platform/internal/domain"
	"github.com/andruwka0/volunteer_platform/internal/store"
)

func StartEventStatusChecker(ctx context.Context, store *store.Store, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[Worker] Event status checker started (interval: %v)", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Worker] Stopped by context")
			return
		case <-ticker.C:
			now := time.Now()

			events, err := store.GetAllEvents()
			if err != nil {
				log.Printf("[Worker] Error getting events: %v", err)
				continue
			}

			for _, event := range events {
				// ACTIVE + время вышло → FINISHED
				if event.Status == domain.EventActive && now.After(event.EndDate) {
					if err := store.UpdateEventStatus(event.ID, domain.EventFinished); err != nil {
						log.Printf("[Worker] Failed to finish event %d: %v", event.ID, err)
					} else {
						log.Printf("[Worker] Auto-finished event %d: %s", event.ID, event.Title)
					}
				}

				// RECRUITING + время начала прошло → CANCELLED
				if event.Status == domain.EventRecruiting && now.After(event.StartDate) {
					if err := store.UpdateEventStatus(event.ID, domain.EventCancelled); err != nil {
						log.Printf("[Worker] Failed to cancel event %d: %v", event.ID, err)
					} else {
						log.Printf("[Worker] Auto-cancelled event %d", event.ID)
					}
				}
			}
		}
	}
}
