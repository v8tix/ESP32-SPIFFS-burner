package application

import (
	"ESP32-SPIFFS-burner/internal/routes"
	"log/slog"
	"sync"
)

type Config struct {
	BaseURL  string
	HTTPPort int
}

type Application struct {
	routes.Routes
	Config Config
	Logger *slog.Logger
	Wg     *sync.WaitGroup
}

func NewApplication(routes routes.Routes, config Config, logger *slog.Logger, wg *sync.WaitGroup) *Application {
	return &Application{Routes: routes, Config: config, Logger: logger, Wg: wg}
}
