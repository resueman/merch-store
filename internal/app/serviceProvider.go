package app

import (
	"context"
	"log"
	"os"

	"github.com/labstack/echo"
	"github.com/resueman/merch-store/config"
	v1 "github.com/resueman/merch-store/internal/delivery/handlers/http/v1"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase"
	"github.com/resueman/merch-store/pkg/closer"
	"github.com/resueman/merch-store/pkg/db"
	"github.com/resueman/merch-store/pkg/db/postgres"
)

type serviceProvider struct {
	configPath  string
	stopSignals []os.Signal

	config *config.Config
	closer *closer.Closer

	dbClient     db.Client
	repositories *repo.Repositories
	usecases     *usecase.Usecase
	handler      *echo.Echo
	//logger       *log.Logger
}

func newServiceProvider(configPath string, stopSignals ...os.Signal) *serviceProvider {
	return &serviceProvider{
		configPath:  configPath,
		stopSignals: stopSignals,
	}
}

func (p *serviceProvider) Config() *config.Config {
	if p.config != nil {
		return p.config
	}

	cfg, err := config.NewConfig(p.configPath)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	p.config = cfg

	return p.config
}

func (p *serviceProvider) Closer() *closer.Closer {
	if p.closer == nil {
		p.closer = closer.NewCloser(p.stopSignals...)
	}

	return p.closer
}

func (p *serviceProvider) DbClient(ctx context.Context) db.Client {
	if p.dbClient != nil {
		return p.dbClient
	}

	client, err := postgres.NewPostgresClient(ctx, p.Config().Postgres.DSN)
	if err != nil {
		log.Fatalf("failed to connect to pg: %v", err)
	}

	err = client.Primary().Ping(ctx)
	if err != nil {
		log.Fatalf("failed ping to pg: %v", err)
	}

	p.dbClient = client
	p.closer.Add(p.dbClient.Close)

	return p.dbClient
}

func (p *serviceProvider) Repositories(ctx context.Context) *repo.Repositories {
	if p.repositories == nil {
		p.repositories = repo.NewRepositories(p.DbClient(ctx))
	}

	return p.repositories
}

func (p *serviceProvider) Usecases(ctx context.Context) *usecase.Usecase {
	if p.usecases == nil {
		p.usecases = usecase.NewUsecase(p.Repositories(ctx))
	}

	return p.usecases
}

func (p *serviceProvider) Handler(ctx context.Context) *echo.Echo {
	if p.handler == nil {
		p.handler = echo.New()
		v1.NewRouter(p.handler, p.Usecases(ctx))
	}

	return p.handler
}

func (p *serviceProvider) Logger() *log.Logger {
	return nil
}
