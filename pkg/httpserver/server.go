package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	server  *http.Server
	errChan chan error
}

func New(handler http.Handler, address string) *Server {
	httpServer := &http.Server{
		Handler: handler,
		Addr:    address,
	}

	s := &Server{
		server:  httpServer,
		errChan: make(chan error, 1),
	}

	return s
}

func (s *Server) Start() {
	go func() {
		s.errChan <- s.server.ListenAndServe()
		close(s.errChan)
	}()
}

func (s *Server) NotifyError() <-chan error {
	return s.errChan
}

func (s *Server) GracefulStop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
