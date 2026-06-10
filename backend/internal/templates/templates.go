package templates

import (
	"html/template"
	"path/filepath"
)

type FuncsProvider interface{ Funcs() template.FuncMap }

// Load загружает конфигурацию из .env и переменных окружения.
func Load(root string, funcs template.FuncMap) *template.Template {
	t := template.New("").Funcs(funcs)
	t = template.Must(t.ParseGlob(filepath.Join(root, "app", "templates", "admin", "*.html")))
	t = template.Must(t.ParseGlob(filepath.Join(root, "app", "templates", "*.html")))
	return t
}
