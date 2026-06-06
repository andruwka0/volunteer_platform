package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashPassword создаёт salted hash пароля для хранения.
func HashPassword(p string) (string, error) {
	salt := make([]byte, 8)
	_, _ = rand.Read(salt)
	h := sha256.Sum256(append(salt, []byte(p)...))
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(h[:]), nil
}

// VerifyPassword проверяет пароль против сохранённого hash.
func VerifyPassword(p, h string) bool {
	parts := strings.Split(h, ":")
	if len(parts) != 2 {
		return false
	}
	salt, _ := hex.DecodeString(parts[0])
	sum := sha256.Sum256(append(salt, []byte(p)...))
	return hex.EncodeToString(sum[:]) == parts[1]
}
