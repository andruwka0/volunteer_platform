package auth

import "testing"

// TestHashAndVerifyPassword проверяет hash/verify цикл пароля.
func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := HashPassword("Password123")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword("Password123", hash) {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword("wrong", hash) {
		t.Fatal("wrong password verified")
	}
}
