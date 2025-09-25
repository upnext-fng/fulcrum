package middleware

import "github.com/labstack/echo/v4"

type Config struct {
	Skipper func(echo.Context) bool
}
