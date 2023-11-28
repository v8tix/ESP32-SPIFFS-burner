package routes

import (
	"ESP32-SPIFFS-burner/assets"
	"ESP32-SPIFFS-burner/internal/handlers"
	"ESP32-SPIFFS-burner/internal/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Routes struct {
	middleware.Middleware
	handlers.Handlers
}

func NewRoutes(middleware middleware.Middleware, handlers handlers.Handlers) Routes {
	return Routes{Middleware: middleware, Handlers: handlers}
}

func (r Routes) Routes() http.Handler {
	mux := chi.NewRouter()
	mux.NotFound(r.NotFound)

	mux.Use(r.RecoverPanic)
	mux.Use(r.SecurityHeaders)

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	mux.Handle("/static/*", fileServer)

	mux.Get("/", r.Home)
	mux.Post("/", r.Home)

	return mux
}
