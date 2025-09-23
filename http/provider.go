// http/provider.go
package http

func NewHTTPService(config Config) HTTPService {
	return NewManager(config)
}
