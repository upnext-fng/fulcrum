// observability/provider.go
package observability

func NewObservabilityService(config Config) ObservabilityService {
	return NewManager(config)
}
