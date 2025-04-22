package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gurebusan/simple-auth/internal/config"
	"github.com/gurebusan/simple-auth/internal/http-server/handlers"
	"github.com/gurebusan/simple-auth/internal/http-server/middlwares/logger"
	"github.com/gurebusan/simple-auth/internal/lib/logger/sl"
)

type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	router *chi.Mux
}

func New(cfg *config.Config, log *slog.Logger, handler *handlers.Handlers) *Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)

	router.Route("/", func(r chi.Router) {
		r.Post("/auth", handler.IssueTokens)
		r.Post("/refresh", handler.RefreshTokens)
	})
	return &Server{
		cfg:    cfg,
		log:    log,
		router: router,
	}
}

func (s *Server) Start() {
	srv := &http.Server{
		Addr:         s.cfg.HTTPServer.Address,
		Handler:      s.router,
		ReadTimeout:  s.cfg.HTTPServer.Timeout,
		WriteTimeout: s.cfg.HTTPServer.Timeout,
		IdleTimeout:  s.cfg.HTTPServer.IdleTimeout,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.log.Error("failed to start server", sl.Err(err))
		}
	}()
	s.log.Info("server started")

	<-done
	s.log.Info("stopping server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.log.Error("failed to stop server", sl.Err(err))

		return
	}
}
