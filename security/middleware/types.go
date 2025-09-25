package middleware

import (
	"github.com/labstack/echo/v4"
)

type AuthContext struct {
	UserID string
	Token  string
	echo.Context
}

type Func func(echo.HandlerFunc) echo.HandlerFunc

type ErrorHandler func(echo.Context, error) error
