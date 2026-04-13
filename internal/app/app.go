package app

import (
	"context"
	"log"
	"net/http"

	"example.com/laneblog/internal/admin"
	"example.com/laneblog/internal/config"
	"example.com/laneblog/internal/frontend"
	apihttp "example.com/laneblog/internal/http"
	"example.com/laneblog/internal/store"
)

type App struct {
	cfg         config.Config
	server      *http.Server
	storeCloser interface{ Close() error }
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	dataStore, err := newStore(cfg)
	if err != nil {
		return nil, err
	}

	adminService := admin.NewService(dataStore)
	frontService := frontend.NewService(dataStore)
	handler := apihttp.NewRouter(cfg, adminService, frontService)
	server := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: handler,
	}

	return &App{
		cfg:         cfg,
		server:      server,
		storeCloser: closerFromStore(dataStore),
	}, nil
}

func newStore(cfg config.Config) (admin.Store, error) {
	return store.NewMySQLStore(cfg.Store.MySQLDSN)
}

func closerFromStore(store admin.Store) interface{ Close() error } {
	closer, ok := store.(interface{ Close() error })
	if !ok {
		return nil
	}
	return closer
}

func (a *App) Run() error {
	log.Printf("server is listening on %s", a.server.Addr)

	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Printf("shutting down server on %s", a.server.Addr)
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	if a.storeCloser != nil {
		return a.storeCloser.Close()
	}
	return nil
}

func (a *App) Config() config.Config {
	return a.cfg
}
