// examples/basic-api/main.go
package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/upnext-fng/fulcrum/security/jwt"
	"github.com/upnext-fng/fulcrum/security/middleware"
	"github.com/upnext-fng/fulcrum/security/password"
	"go.uber.org/fx"

	"github.com/upnext-fng/fulcrum/configuration"
	"github.com/upnext-fng/fulcrum/database"
	"github.com/upnext-fng/fulcrum/http"
	"github.com/upnext-fng/fulcrum/observability"
	"github.com/upnext-fng/fulcrum/security"
)

// Simple User model for testing
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

// API Routes
type APIRoutes struct {
	db       database.DatabaseService
	security security.SecurityService
	logger   observability.ObservabilityService
}

func NewAPIRoutes(
	db database.DatabaseService,
	security security.SecurityService,
	logger observability.ObservabilityService,
) *APIRoutes {
	return &APIRoutes{
		db:       db,
		security: security,
		logger:   logger,
	}
}

func (r *APIRoutes) RegisterRoutes(e *echo.Echo) {
	// Test endpoint - no auth required
	e.GET("/", r.Welcome)

	// Auth endpoints
	e.POST("/register", r.Register)
	e.POST("/login", r.Login)
	e.POST("/refresh", r.RefreshToken)

	// Protected endpoint
	protected := e.Group("/protected")
	protected.Use(r.security.JWTMiddleware())
	protected.GET("/profile", r.GetProfile)
	protected.GET("/validate", r.ValidateTokenDemo)
}

func (r *APIRoutes) Welcome(c echo.Context) error {
	r.logger.Logger().Info("Welcome endpoint called")
	return c.JSON(200, map[string]string{
		"message": "Welcome to the microservices components API!",
		"status":  "operational",
	})
}

func (r *APIRoutes) Register(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(400, "Invalid request body")
	}

	// Hash password
	hashedPassword, err := r.security.HashPassword(req.Password)
	if err != nil {
		r.logger.Logger().WithError(err).Error("Failed to hash password")
		return echo.NewHTTPError(500, "Failed to process password")
	}

	// Create user
	user := User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := r.db.Connection().Create(&user).Error; err != nil {
		r.logger.Logger().WithError(err).Error("Failed to create user")
		return echo.NewHTTPError(409, "Username already exists")
	}

	r.logger.Logger().WithField("user_id", user.ID).Info("User registered successfully")

	return c.JSON(201, map[string]interface{}{
		"message":  "User registered successfully",
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

func (r *APIRoutes) Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(400, "Invalid request body")
	}

	// Find user
	var user User
	if err := r.db.Connection().Where("username = ?", req.Username).First(&user).Error; err != nil {
		r.logger.Logger().WithField("username", req.Username).Warn("Login attempt with invalid username")
		return echo.NewHTTPError(401, "Invalid credentials")
	}

	r.logger.Logger().WithField("user_id", user.ID).Info("User found", user)

	// Verify password
	if err := r.security.VerifyPassword(user.Password, req.Password); err != nil {
		r.logger.Logger().WithField("user_id", user.ID).Warn("Login attempt with invalid password")
		return echo.NewHTTPError(401, "Invalid credentials")
	}

	// Generate token pair (access + refresh token) using strongly-typed interface
	tokenRequest := security.TokenRequest{
		UserClaims: security.UserClaims{
			UserID:   strconv.Itoa(int(user.ID)),
			Username: user.Username,
			Email:    user.Email,
			ClientID: "web",
		},
		Metadata: security.TokenMetadata{
			Purpose:     "login",
			Source:      "web",
			Environment: "production",
			Scopes:      []string{"read", "write"},
			CreatedAt:   time.Now(),
		},
		RefreshToken: true, // Enable refresh token generation
	}

	tokenResponse, err := r.security.GenerateToken(tokenRequest)
	if err != nil {
		r.logger.Logger().WithError(err).Error("Failed to generate token")
		return echo.NewHTTPError(500, "Failed to generate token")
	}

	r.logger.Logger().WithField("user_id", user.ID).Info("User logged in successfully")

	response := map[string]interface{}{
		"message":      "Login successful",
		"access_token": tokenResponse.AccessToken,
		"token_type":   tokenResponse.TokenType,
		"expires_in":   tokenResponse.ExpiresIn,
		"expires_at":   tokenResponse.ExpiresAt,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	}

	// Include refresh token if generated
	if tokenResponse.RefreshToken != "" {
		response["refresh_token"] = tokenResponse.RefreshToken
	}

	return c.JSON(200, response)
}

func (r *APIRoutes) RefreshToken(c echo.Context) error {
	// Parse request body
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(400, "Invalid request body")
	}

	if req.RefreshToken == "" {
		return echo.NewHTTPError(400, "Refresh token is required")
	}

	// Use the security service to refresh the access token
	tokenResponse, err := r.security.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		r.logger.Logger().WithError(err).Error("Failed to refresh access token")
		return echo.NewHTTPError(401, "Invalid or expired refresh token")
	}

	return c.JSON(200, map[string]interface{}{
		"access_token": tokenResponse.AccessToken,
		"token_type":   tokenResponse.TokenType,
		"expires_in":   tokenResponse.ExpiresIn,
		"expires_at":   tokenResponse.ExpiresAt,
		"message":      "Token refreshed successfully",
	})
}

func (r *APIRoutes) GetProfile(c echo.Context) error {
	// JWT middleware has already validated the token
	r.logger.Logger().Info("Protected endpoint accessed")

	return c.JSON(200, map[string]interface{}{
		"message": "This is a protected endpoint",
		"note":    "JWT token is valid",
		"data":    "User profile data would go here",
	})
}

func (r *APIRoutes) ValidateTokenDemo(c echo.Context) error {
	// Extract token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return echo.NewHTTPError(400, "Missing Authorization header")
	}

	// Parse Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return echo.NewHTTPError(400, "Invalid Authorization header format")
	}
	token := parts[1]

	// Validate token using strongly-typed interface
	validationRequest := security.ValidationRequest{
		Token:          token,
		RequiredScopes: []string{"read", "write"},
	}

	validationResponse, err := r.security.ValidateToken(validationRequest)
	if err != nil {
		r.logger.Logger().WithError(err).Error("Token validation failed")
		return echo.NewHTTPError(401, "Token validation failed")
	}

	if !validationResponse.Valid {
		return echo.NewHTTPError(401, "Invalid token")
	}

	return c.JSON(200, map[string]interface{}{
		"message":       "Token validation successful",
		"valid":         validationResponse.Valid,
		"user_claims":   validationResponse.UserClaims,
		"metadata":      validationResponse.Metadata,
		"custom_claims": validationResponse.CustomClaims,
		"expires_at":    validationResponse.ExpiresAt,
		"issued_at":     validationResponse.IssuedAt,
	})
}

// Application Configuration
type AppConfig struct {
	Database      database.Config      `mapstructure:"database"`
	HTTP          http.Config          `mapstructure:"http"`
	Security      security.Config      `mapstructure:"security"`
	Observability observability.Config `mapstructure:"observability"`
}

// Load configuration with sensible defaults
func LoadConfig() AppConfig {
	config := AppConfig{
		// Database defaults
		Database: database.Config{
			Host:     "10.6.2.29",
			Port:     5432,
			Database: "postgres",
			Username: "postgres",
			Password: "Postgres!23456",
			SSLMode:  "disable",
		},
		// HTTP defaults
		HTTP: http.Config{
			Port:         "8888",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		// Security defaults
		Security: security.Config{
			JWT: jwt.Config{
				AccessTokenTTL: time.Hour,
			},
			Password: password.Config{},
			Middleware: middleware.Config{
				Skipper: func(e echo.Context) bool {
					return true
				},
				JWTConfig: jwt.Config{},
			},
		},
		// Observability defaults
		Observability: observability.Config{
			LogLevel:  "info",
			LogFormat: "json",
		},
	}

	// Try to load from config file or environment
	configService := configuration.NewConfigurationService(configuration.Config{
		ConfigPath: ".",
		EnvPrefix:  "APP",
	})
	configService.LoadConfig(&config)

	return config
}

// Application startup and lifecycle
func StartApplication(
	httpService http.HTTPService,
	dbService database.DatabaseService,
	obsService observability.ObservabilityService,
	apiRoutes *APIRoutes,
	lifecycle fx.Lifecycle,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			obsService.Logger().Info("Starting microservices application...")

			// Auto-migrate database schema
			if err := dbService.Connection().AutoMigrate(&User{}); err != nil {
				obsService.Logger().WithError(err).Fatal("Failed to migrate database")
				return err
			}
			obsService.Logger().Info("Database migration completed")

			// Test database connection
			if err := dbService.HealthCheck(); err != nil {
				obsService.Logger().WithError(err).Fatal("Database health check failed")
				return err
			}
			obsService.Logger().Info("Database connection verified")

			// Register routes
			httpService.RegisterRoutes(apiRoutes)
			obsService.Logger().Info("API routes registered")

			// Add middleware
			httpService.GetEngine().Use(obsService.RequestLoggerMiddleware())

			// Add health endpoint
			httpService.GetEngine().GET("/health", obsService.HealthEndpoint())

			// Start HTTP server in goroutine
			go func() {
				if err := httpService.StartServer(); err != nil {
					obsService.Logger().WithError(err).Fatal("HTTP server failed")
				}
			}()

			//obsService.Logger().Info("🚀 Application started successfully on port " + httpService.GetEngine().(*echo.Echo).Server.Addr)
			obsService.Logger().Info("📝 Available endpoints:")
			obsService.Logger().Info("  GET  /           - Welcome message")
			obsService.Logger().Info("  GET  /health     - Health check")
			obsService.Logger().Info("  POST /register   - User registration")
			obsService.Logger().Info("  POST /login      - User authentication (returns access + refresh token)")
			obsService.Logger().Info("  POST /refresh    - Refresh access token using refresh token")
			obsService.Logger().Info("  GET  /protected/profile - Protected endpoint (requires JWT)")
			obsService.Logger().Info("  GET  /protected/validate - Token validation demo")

			return nil
		},
		OnStop: func(ctx context.Context) error {
			obsService.Logger().Info("Shutting down application...")

			// Stop HTTP server
			if err := httpService.StopServer(ctx); err != nil {
				obsService.Logger().WithError(err).Error("Error stopping HTTP server")
			}

			// Close database connection
			if err := dbService.Close(); err != nil {
				obsService.Logger().WithError(err).Error("Error closing database")
			}

			obsService.Logger().Info("Application shutdown completed")
			return nil
		},
	})
}

func main() {
	config := LoadConfig()

	fx.New(
		// Provide module configurations
		fx.Provide(func() database.Config { return config.Database }),
		fx.Provide(func() http.Config { return config.HTTP }),
		fx.Provide(func() security.Config { return config.Security }),
		fx.Provide(func() observability.Config { return config.Observability }),

		// Infrastructure modules
		database.Module,
		http.Module,
		security.Module,
		observability.Module,

		// Application services
		fx.Provide(NewAPIRoutes),

		// Application startup
		fx.Invoke(StartApplication),
	).Run()
}
