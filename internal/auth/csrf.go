package auth

import "sync"

type CSRFStore struct {
	mu     sync.Mutex
	tokens map[string]string
}

// NewCSRFStore создаёт in-memory хранилище CSRF-токенов.
func NewCSRFStore() *CSRFStore { return &CSRFStore{tokens: map[string]string{}} }

// Get возвращает значение из in-memory хранилища по ключу.
func (c *CSRFStore) Get(sessionID string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tokens[sessionID]
}

// Set сохраняет значение в in-memory хранилище.
func (c *CSRFStore) Set(sessionID, token string) {
	c.mu.Lock()
	c.tokens[sessionID] = token
	c.mu.Unlock()
}

// Validate проверяет CSRF-токен для указанной сессии.
func (c *CSRFStore) Validate(sessionID, token string) bool {
	return token != "" && token == c.Get(sessionID)
}
