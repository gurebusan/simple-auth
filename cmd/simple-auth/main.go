package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gurebusan/simple-auth/internal/config"
	"github.com/gurebusan/simple-auth/internal/http-server/handlers"
	"github.com/gurebusan/simple-auth/internal/http-server/server"
	"github.com/gurebusan/simple-auth/internal/lib/logger/handlers/slogpretty"
	"github.com/gurebusan/simple-auth/internal/lib/logger/sl"
	"github.com/gurebusan/simple-auth/internal/lib/notifier"
	"github.com/gurebusan/simple-auth/internal/lib/token/manager"
	"github.com/gurebusan/simple-auth/internal/service"
	"github.com/gurebusan/simple-auth/internal/storage/postgres"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting simple-auth", slog.String("env", cfg.Env))
	log.Info("initializing server", slog.String("address", cfg.HTTPServer.Address))

	ctx := context.Background()

	storage, err := postgres.New(ctx, cfg.StorageDSN)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	manager := manager.NewTokenManager(cfg.Token.Secret)
	notifier := notifier.NewMockNotifier(log, cfg)
	service := service.New(ctx, storage, manager, notifier, cfg)
	handlers := handlers.New(log, cfg, service)
	server := server.New(cfg, log, handlers)

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))
	server.Start()

	log.Info("server stopped")
	storage.Close()
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
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
