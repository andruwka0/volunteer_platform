package upload

import (
	"encoding/base64"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SaveUpload сохраняет multipart-файл в static uploads и возвращает URL.
func SaveUpload(file multipart.File, h *multipart.FileHeader, prefix string) (string, error) {
	if file == nil || h == nil || h.Filename == "" {
		return "", nil
	}
	ext := strings.ToLower(filepath.Ext(h.Filename))
	if ext == "" {
		ext = ".bin"
	}
	name := prefix + "_" + strings.ReplaceAll(time.Now().Format("20060102150405.000000000"), ".", "") + ext
	dir := uploadDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	out, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}
	return "/static/uploads/" + name, nil
}

// SaveDataURL сохраняет base64 data URL в static uploads и возвращает URL.
func SaveDataURL(dataURL, prefix string) (string, error) {
	idx := strings.Index(dataURL, ",")
	if idx < 0 {
		return "", errors.New("invalid data url")
	}
	raw, err := base64.StdEncoding.DecodeString(dataURL[idx+1:])
	if err != nil {
		return "", err
	}
	name := prefix + "_" + strings.ReplaceAll(time.Now().Format("20060102150405.000000000"), ".", "") + ".png"
	dir := uploadDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, name), raw, 0644); err != nil {
		return "", err
	}
	return "/static/uploads/" + name, nil
}

// uploadDir находит директорию uploads относительно корня проекта.
func uploadDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return filepath.Join("app", "static", "uploads")
	}
	for {
		candidate := filepath.Join(dir, "app", "static")
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return filepath.Join(candidate, "uploads")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Join("app", "static", "uploads")
		}
		dir = parent
	}
}

// SaveAvatarFromDataURL сохраняет avatar data URL как upload-файл.
func SaveAvatarFromDataURL(dataURL string) (string, error) { return SaveDataURL(dataURL, "avatar") }
