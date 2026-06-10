package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

var ErrInvalidToken = errors.New("недействительный токен")

var (
	mu     sync.RWMutex
	tokens = make(map[string]int64)
)

func GenerateToken(userID int64) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	mu.Lock()
	defer mu.Unlock()
	tokens[token] = userID
	return token, nil
}

func ValidateToken(token string) (int64, error) {
	mu.Lock()
	defer mu.Unlock()
	userID, ok := tokens[token]
	if !ok {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
