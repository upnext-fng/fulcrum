package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/upnext-fng/fulcrum/infra/http"
	"github.com/upnext-fng/fulcrum/logger"
)

func main() {
	log := logger.NewLog("main", logger.WithDevelopment(true))
	log.Info("Hello World")

	server := http.NewHTTPServer(log, http.DefaultConfig())

	server.RouteRegistration(func(engine *echo.Echo) {
		engine.GET("/", func(c echo.Context) error {
			return c.String(200, "Hello World")
		})

		auth := engine.Group("/auth")
		{
			auth.GET("/login", func(c echo.Context) error {
				return c.String(200, "Login")
			})
		}

	})

	if err := server.Start(context.TODO()); err != nil {
		log.WithErr(err).Info("Failed to start server")
		os.Exit(1)
	}

	log.Info("Server is running")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
}
