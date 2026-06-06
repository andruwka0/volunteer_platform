package config

import (
	"os"
	"strings"
)

type Config struct {
	AppName           string
	DatabaseURL       string
	SecretKey         string
	SessionCookieName string
	Addr              string
	Debug             bool
	TemplatesDir      string
	StaticDir         string
	UploadDir         string
}

// Load загружает конфигурацию из .env и переменных окружения.
func Load() Config {
	loadEnv()
	return Config{
		AppName:           env("APP_NAME", "Volunteer Rating Platform"),
		DatabaseURL:       env("DATABASE_URL", "./volunteer_platform.json"),
		SecretKey:         env("SECRET_KEY", "change-me-in-production"),
		SessionCookieName: env("SESSION_COOKIE_NAME", "volunteer_session"),
		Addr:              env("ADDR", ":8000"),
		Debug:             env("DEBUG", "") == "1" || strings.EqualFold(env("DEBUG", ""), "true"),
		TemplatesDir:      env("TEMPLATES_DIR", "app/templates"),
		StaticDir:         env("STATIC_DIR", "app/static"),
		UploadDir:         env("UPLOAD_DIR", "app/static/uploads"),
	}
}

// env возвращает значение переменной окружения или fallback.
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

// loadEnv подмешивает значения из локального .env, не перезаписывая окружение.
func loadEnv() {
	b, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, ln := range strings.Split(string(b), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" || strings.HasPrefix(ln, "#") || !strings.Contains(ln, "=") {
			continue
		}
		p := strings.SplitN(ln, "=", 2)
		if os.Getenv(p[0]) == "" {
			os.Setenv(p[0], strings.Trim(p[1], "\"'"))
		}
	}
}
