package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/resueman/merch-store/pkg/httpserver"
)

type App struct {
	provider *serviceProvider
}

func NewApp(configPath string, stopSignals ...os.Signal) *App {
	app := &App{
		provider: newServiceProvider(configPath, stopSignals...),
	}

	return app
}

func (a *App) Run() {
	defer func() {
		a.provider.Closer().CloseAll()
		a.provider.Closer().Wait()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	e := echo.New()
	httpServer := httpserver.New(a.provider.Handler(ctx, e), a.provider.Config().HTTPServer.Port)
	a.provider.Closer().Add(func() error {
		log.Info("stopping http server gracefully...")

		return httpServer.GracefulStop()
	})

	log.Info("starting http server...")
	httpServer.Start()

	log.Info("listening...")

	select {
	case <-a.provider.Closer().Notify():
		log.Info("got stop signal...")
	case err := <-httpServer.NotifyError():
		log.Error(fmt.Errorf("http server stopped by error: %w", err))
	}

	<-a.provider.Closer().Done()
	log.Info("all resources released")
}
