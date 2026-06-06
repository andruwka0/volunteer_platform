package auth

import "sync"

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]int
}

// NewSessionStore создаёт in-memory хранилище сессий.
func NewSessionStore() *SessionStore { return &SessionStore{sessions: map[string]int{}} }

// Get возвращает значение из in-memory хранилища по ключу.
func (s *SessionStore) Get(id string) int { s.mu.Lock(); defer s.mu.Unlock(); return s.sessions[id] }

// Set сохраняет значение в in-memory хранилище.
func (s *SessionStore) Set(id string, userID int) {
	s.mu.Lock()
	s.sessions[id] = userID
	s.mu.Unlock()
}

// Delete удаляет значение из in-memory хранилища.
func (s *SessionStore) Delete(id string) { s.mu.Lock(); delete(s.sessions, id); s.mu.Unlock() }
