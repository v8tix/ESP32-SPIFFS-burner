package main

import (
	"ESP32-SPIFFS-burner/internal/application"
	"ESP32-SPIFFS-burner/internal/handlers"
	"ESP32-SPIFFS-burner/internal/middleware"
	"ESP32-SPIFFS-burner/internal/routes"
	"ESP32-SPIFFS-burner/internal/utils"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

const (
	defaultIdleTimeout    = time.Minute
	defaultReadTimeout    = 5 * time.Second
	defaultWriteTimeout   = 10 * time.Second
	defaultShutdownPeriod = 30 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	var cfg application.Config
	var wg sync.WaitGroup

	flag.IntVar(&cfg.HTTPPort, "port", 8080, "port to listen on for HTTP requests")
	showHelp := flag.Bool("help", false, "show help message")
	flag.Parse()

	if *showHelp {
		printUsage()
		os.Exit(0)
	}

	ut := utils.NewUtils(logger)
	mid := middleware.NewMiddleware(ut)
	handle := handlers.NewHandlers(ut)
	route := routes.NewRoutes(mid, handle)

	app := application.NewApplication(route, cfg, logger, &wg)
	return serveHTTP(app)
}

func printUsage() {
	fmt.Println("Usage: esp32-burner [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func serveHTTP(app *application.Application) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.HTTPPort),
		Handler:      app.Routes.Routes(),
		ErrorLog:     slog.NewLogLogger(app.Logger.Handler(), slog.LevelWarn),
		IdleTimeout:  defaultIdleTimeout,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	shutdownErrorChan := make(chan error)

	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)
		<-quitChan

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownPeriod)
		defer cancel()

		shutdownErrorChan <- srv.Shutdown(ctx)
	}()

	app.Logger.Info("starting server", slog.Group("server", "url", fmt.Sprintf("http://localhost%s", srv.Addr)))

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownErrorChan
	if err != nil {
		return err
	}

	app.Logger.Info("stopped server", slog.Group("server", "url", fmt.Sprintf("http://localhost%s", srv.Addr)))

	app.Wg.Wait()
	return nil
}
