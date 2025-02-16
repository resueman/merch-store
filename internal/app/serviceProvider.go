package app

import (
	"context"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/resueman/merch-store/config"
	v1 "github.com/resueman/merch-store/internal/delivery/handlers/http/v1"
	"github.com/resueman/merch-store/internal/delivery/middleware"
	"github.com/resueman/merch-store/internal/repo"
	"github.com/resueman/merch-store/internal/usecase"
	"github.com/resueman/merch-store/pkg/closer"
	"github.com/resueman/merch-store/pkg/db"
	"github.com/resueman/merch-store/pkg/db/postgres"
	"github.com/resueman/merch-store/pkg/password"
)

type serviceProvider struct {
	configPath  string
	stopSignals []os.Signal

	config *config.Config
	closer *closer.Closer

	dbClient        db.Client
	txManager       db.TxManager
	passwordManager *password.BcryptManager
	repositories    *repo.Repositories
	usecases        *usecase.Usecase
	authMiddleware  *middleware.AuthMiddleware
	handler         *echo.Echo
	//logger       *log.Logger
}

func NewServiceProvider(configPath string, stopSignals ...os.Signal) *serviceProvider {
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
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	err = client.Primary().Ping(ctx)
	if err != nil {
		log.Fatalf("failed ping to postgres: %v", err)
	}

	p.dbClient = client
	p.Closer().Add(func() error {
		log.Info("stopping db client...")

		return p.dbClient.Close()
	})

	return p.dbClient
}

func (p *serviceProvider) TxManager(ctx context.Context) db.TxManager {
	if p.txManager == nil {
		timeout := time.Duration(p.Config().TxManager.TimeoutMs) * time.Millisecond
		maxRetries := p.Config().TxManager.MaxRetries
		p.txManager = postgres.NewTxManager(p.DbClient(ctx), timeout, maxRetries)
	}

	return p.txManager
}

func (p *serviceProvider) Repositories(ctx context.Context) *repo.Repositories {
	if p.repositories == nil {
		p.repositories = repo.NewRepositories(p.DbClient(ctx))
	}

	return p.repositories
}

func (p *serviceProvider) PasswordManager() *password.BcryptManager {
	if p.passwordManager == nil {
		salt := "1234567890" // надо генерировать для каждого пользователя, но пока так, переделаю, если успею
		p.passwordManager = password.NewPasswordManager(salt)
	}

	return p.passwordManager
}

func (p *serviceProvider) Usecases(ctx context.Context) *usecase.Usecase {
	if p.usecases == nil {
		secret := p.Config().JWT.Secret
		ttl := time.Duration(p.Config().JWT.TTLMin) * time.Minute

		p.usecases = usecase.NewUsecase(p.Repositories(ctx), p.TxManager(ctx), p.PasswordManager(), secret, ttl)
	}

	return p.usecases
}

func (p *serviceProvider) AuthMiddleware(ctx context.Context) *middleware.AuthMiddleware {
	if p.authMiddleware == nil {
		p.authMiddleware = middleware.NewAuthMiddleware(p.Usecases(ctx))
	}

	return p.authMiddleware
}

func (p *serviceProvider) Handler(ctx context.Context, e *echo.Echo) *echo.Echo {
	if p.handler == nil {
		p.handler = e
		v1.NewRouter(p.handler, p.Usecases(ctx), p.AuthMiddleware(ctx))
	}

	return p.handler
}

func (p *serviceProvider) Logger() *log.Logger {
	return nil
}
