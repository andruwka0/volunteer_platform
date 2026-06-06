package service

import "testing"

// TestEventsResolveCoverPrefersCroppedThenUploadThenCurrent проверяет приоритеты выбора обложки.
func TestEventsResolveCoverPrefersCroppedThenUploadThenCurrent(t *testing.T) {
	events := Events{}
	if got := events.ResolveCover("old", "data", func(string) (string, error) { return "cropped", nil }, nil); got != "cropped" {
		t.Fatalf("cropped cover = %q, want cropped", got)
	}
	if got := events.ResolveCover("old", "data", func(string) (string, error) { return "", errImage }, func() (string, error) { return "upload", nil }); got != "upload" {
		t.Fatalf("fallback upload cover = %q, want upload", got)
	}
	if got := events.ResolveCover("old", "", nil, func() (string, error) { return "", errImage }); got != "old" {
		t.Fatalf("unchanged cover = %q, want old", got)
	}
}

type imageError string

// Error возвращает текст тестовой ошибки imageError.
func (e imageError) Error() string { return string(e) }

const errImage = imageError("image error")
