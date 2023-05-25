package server

import (
	"context"
	"net/http"
	"time"

	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/app"
	"github.com/Pos1t1veM1ndset/anti-bruteforce/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	server *http.Server
}

func New(app app.RequestValidator, cfg config.Config, logg zap.SugaredLogger) *Server {
	return &Server{
		server: &http.Server{
			Addr:              cfg.Server.Host + cfg.Server.Port,
			Handler:           loggingMiddleware(newHandler(app), logg),
			ReadHeaderTimeout: time.Second * 3,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	<-ctx.Done()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
