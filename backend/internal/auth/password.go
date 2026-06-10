package auth

import "volunteer-platform/internal/service"

// HashPassword создаёт salted hash пароля для хранения.
func HashPassword(p string) (string, error) { return service.HashPassword(p) }

// VerifyPassword проверяет пароль против сохранённого hash.
func VerifyPassword(p, h string) bool { return service.VerifyPassword(p, h) }
