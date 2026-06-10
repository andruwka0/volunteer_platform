package router

import (
	"net/http"

	"volunteer-platform/internal/handler"
)

// New собирает HTTP router поверх готового набора handlers.
func New(h *handler.App) http.Handler { return h.Routes() }
