package middleware

import (
	"github.com/upnext-fng/fulcrum/security/jwt"
)

func NewService(config Config, jwtService jwt.Service) Service {
	return NewManager(config, jwtService)
}
