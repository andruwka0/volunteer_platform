package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var apiPrefixes = []string{"/auth", "/events", "/admin"}

func main() {
	backendRaw := env("BACKEND_URL", "http://127.0.0.1:8080")
	backendURL, err := url.Parse(backendRaw)
	if err != nil {
		log.Fatalf("invalid BACKEND_URL %q: %v", backendRaw, err)
	}

	frontendAddr := env("FRONTEND_ADDR", "127.0.0.1:5173")
	publicRoot := env("FRONTEND_ROOT", ".")
	if _, err := os.Stat(filepath.Join(publicRoot, "index.html")); err != nil {
		log.Fatalf("frontend root %q must contain index.html: %v", publicRoot, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, proxyErr error) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Backend недоступен. Сначала запустите backend: cd backend && go run ./cmd/server",
			"details": proxyErr.Error(),
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAPIPath(r.URL.Path) {
			proxy.ServeHTTP(w, r)
			return
		}
		serveFrontend(publicRoot, w, r)
	})

	log.Printf("Frontend: http://%s", frontendAddr)
	log.Printf("Backend proxy: %s", backendURL.String())
	log.Fatal(http.ListenAndServe(frontendAddr, handler))
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func isAPIPath(path string) bool {
	for _, prefix := range apiPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func serveFrontend(root string, w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if path == "." || path == string(filepath.Separator) {
		path = "index.html"
	}

	fullPath := filepath.Join(root, path)
	rootAbs, rootErr := filepath.Abs(root)
	fileAbs, fileErr := filepath.Abs(fullPath)
	if rootErr != nil || fileErr != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if rel, err := filepath.Rel(rootAbs, fileAbs); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if info, err := os.Stat(fileAbs); err == nil && !info.IsDir() {
		http.ServeFile(w, r, fileAbs)
		return
	}

	http.ServeFile(w, r, filepath.Join(root, "index.html"))
}
