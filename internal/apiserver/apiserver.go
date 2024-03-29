package apiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Config struct {
	BindAddress string
}

type APIServer struct {
	router  *chi.Mux
	cfg     Config
	server  *http.Server
	service service
}

func New(cfg Config, service service) *APIServer {
	router := chi.NewRouter()

	return &APIServer{
		cfg:     cfg,
		service: service,
		router:  router,
		server: &http.Server{
			Addr:              cfg.BindAddress,
			ReadHeaderTimeout: 5 * time.Second,
			Handler:           router,
		},
	}
}

func (s *APIServer) Run(ctx context.Context) error {
	defer zap.L().Info("server stopped")

	s.configRouter()

	go func() {
		<-ctx.Done()

		gfCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		//nolint: contextcheck
		if err := s.server.Shutdown(gfCtx); err != nil {
			zap.L().With(zap.Error(err)).Warn("failed to gracefully shutdown server")

			return
		}
	}()

	zap.L().Info("sever starting", zap.String("port", s.cfg.BindAddress))

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("s.server.ListenAndServe(): %w", err)
	}

	return nil
}

func (s *APIServer) configRouter() {
	s.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Post("/good/create", s.createGood)
			r.Patch("/good/update", s.updateGoods)
			r.Delete("/good/remove", s.removeGoods)
			r.Get("/good/list", s.getGoods)
			r.Patch("/good/reprioritize", s.reprioritizeGood)
		})
	})
}
