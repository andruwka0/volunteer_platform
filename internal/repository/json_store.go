package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"volunteer-platform/internal/domain"
)

type JSONStore struct {
	mu            sync.Mutex
	path          string
	Next          map[string]int
	Users         []domain.User
	Events        []domain.Event
	Reviews       []domain.Review
	Rules         []domain.Rule
	Participants  []domain.EventParticipant
	Organizers    []domain.EventOrganizer
	EventRoles    []domain.EventRole
	RoleChoices   []domain.EventParticipantRoleChoice
	Notifications []domain.Notification
}

// OpenJSONStore открывает JSON-хранилище и нормализует старые sqlite-пути.
func OpenJSONStore(databaseURL string) (*JSONStore, error) {
	p := dbPath(databaseURL)
	if strings.HasSuffix(p, ".db") {
		p = strings.TrimSuffix(p, ".db") + ".json"
	}
	s := &JSONStore{path: p, Next: map[string]int{}}
	b, err := os.ReadFile(p)
	if err == nil {
		_ = json.Unmarshal(b, s)
	}
	if s.Next == nil {
		s.Next = map[string]int{}
	}
	return s, nil
}

// dbPath убирает legacy sqlite-префиксы из строки подключения.
func dbPath(s string) string {
	return strings.TrimPrefix(strings.TrimPrefix(s, "sqlite:///"), "sqlite://")
}

// Lock открывает критическую секцию JSONStore для составных операций.
func (s *JSONStore) Lock() { s.mu.Lock() }

// Unlock закрывает критическую секцию JSONStore.
func (s *JSONStore) Unlock() { s.mu.Unlock() }

// SaveUnlocked сохраняет JSONStore на диск без самостоятельного lock.
func (s *JSONStore) SaveUnlocked() {
	b, _ := json.MarshalIndent(s, "", "  ")
	if dir := filepath.Dir(s.path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
	_ = os.WriteFile(s.path, b, 0644)
}

// NextIDUnlocked выдаёт следующий ID без самостоятельного lock.
func (s *JSONStore) NextIDUnlocked(k string) int { s.Next[k]++; return s.Next[k] }

// Save потокобезопасно сохраняет JSONStore на диск.
func (s *JSONStore) Save() {
	s.Lock()
	defer s.Unlock()
	s.SaveUnlocked()
}

// NextID потокобезопасно выдаёт следующий ID для указанной сущности.
func (s *JSONStore) NextID(k string) int {
	s.Lock()
	defer s.Unlock()
	return s.NextIDUnlocked(k)
}
