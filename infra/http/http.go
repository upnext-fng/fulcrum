package http

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/upnext-fng/fulcrum/infra/http/middleware"
	"github.com/upnext-fng/fulcrum/logger"
)

type EchoServer struct {
	engine   *echo.Echo
	handlers []func(engine *echo.Echo)
	logger   *logger.Logger
	config   Config
	server   *http.Server
}

func NewHTTPServer(logger *logger.Logger, config Config) *EchoServer {
	return &EchoServer{
		engine:   echo.New(),
		handlers: []func(engine *echo.Echo){},
		logger:   logger,
		config:   config,
	}
}

func (s *EchoServer) RouteRegistration(hdl func(engine *echo.Echo)) {
	s.handlers = append(s.handlers, hdl)
}

func (s *EchoServer) Start(ctx context.Context) error {

	if err := s.Configure(); err != nil {
		return err
	}

	for _, hdl := range s.handlers {
		hdl(s.engine)
	}

	s.server = &http.Server{
		Addr:           s.config.Address(),
		Handler:        s.engine,
		ReadTimeout:    s.config.Timeouts.ReadTimeout,
		WriteTimeout:   s.config.Timeouts.WriteTimeout,
		IdleTimeout:    s.config.Timeouts.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	return s.server.ListenAndServe()
}

func (s *EchoServer) Configure() error {
	// we will configure server and middleware here
	if s.engine == nil {
		s.engine = echo.New()
	}

	s.engine.Use(middleware.RequestID())
	s.engine.Use(middleware.RequestLogger("http"))
	s.engine.Use(middleware.Timeout(30))
	s.engine.Use(middleware.CORS(s.config.CORS.AllowOrigins, s.config.CORS.AllowMethods))

	s.engine.Logger = NewEchoLogger(s.logger)
	return nil
}

func (s *EchoServer) Stop(ctx context.Context) error {
	if s.server != nil {
		_ = s.server.Shutdown(ctx)
	}

	return nil
}
