package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/logger/handlers/slogpretty"
	"url-shortener/internal/lib/sl"
	"url-shortener/internal/storage/sqlite"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	//"golang.org/x/tools/godoc/redirect"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	// reading CONFIG_PATH from .env for MustLoad func
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	cfg := config.MustLoad()

	// initializating logger
	log := setupLogger(cfg.Env)

	log.With(slog.String("env", cfg.Env))                           // string to add for different level of logger
	log.Info("starting url-shortener", slog.String("version", "1")) // start message for level info
	log.Debug("debug messagges are enabled")                        // start message for level debug

	storage, err := sqlite.New(cfg.Storage_Path)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}
	_ = storage

	// init router
	router := chi.NewRouter()
	// connecting middleware to router
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)

	//Если вы решили завести себе такой middleware, разместить его рекомендую в internal/http-server/middleware
	// router.Use(middleware.New())

	router.Use(middleware.Recoverer) // if some panic happens - app need to be recovered
	router.Use(middleware.URLFormat) // for pretty ulr-s display

	router.Route("/ulr", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string{
			cfg.HTTP_Server.User: cfg.HTTP_Server.Password,
		}))
		r.Post("/", save.New(log, storage))
		//TODO: add Delete /url/{id}
	})

	router.Post("/url", save.New(log, storage))
	router.Get("/{alias}", redirect.New(log, storage))
	// TODO: write handler for Delete
	router.Delete("/url/{alias}", redirect.New(log, storage))

	log.Info("starting server", slog.String("adress", cfg.Adress))

	srv := &http.Server{
		Addr:         cfg.Adress,
		Handler:      router,
		ReadTimeout:  cfg.HTTP_Server.TimeOut, // timeout for processing request
		WriteTimeout: cfg.HTTP_Server.TimeOut, // same
		IdleTimeout:  cfg.HTTP_Server.Idle_Timeout,
	}

	if err := srv.ListenAndServe(); err != nil { // blocking func - no read until stopped
		log.Error("failed to start server")
	}

	log.Error("server stopped")

}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log

}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
