// http/manager.go
package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type manager struct {
	echo   *echo.Echo
	server *http.Server
	config Config
}

func NewManager(config Config) HTTPService {
	e := echo.New()

	// Basic middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return &manager{
		echo:   e,
		config: config,
	}
}

func (m *manager) StartServer() error {
	m.server = &http.Server{
		Addr:         ":" + m.config.Port,
		Handler:      m.echo,
		ReadTimeout:  m.config.ReadTimeout,
		WriteTimeout: m.config.WriteTimeout,
	}

	fmt.Printf("Starting HTTP server on port %s\n", m.config.Port)
	return m.server.ListenAndServe()
}

func (m *manager) StopServer(ctx context.Context) error {
	if m.server != nil {
		return m.server.Shutdown(ctx)
	}
	return nil
}

func (m *manager) GetEngine() *echo.Echo {
	return m.echo
}

func (m *manager) RegisterRoutes(routes RouteRegistrar) {
	routes.RegisterRoutes(m.echo)
}
