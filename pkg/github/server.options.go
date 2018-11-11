package github

// ServerOption an option function to customize Server
type ServerOption func(*Server)

// WithWebhookSecretKey sets Server's webhookSecretKey
func WithWebhookSecretKey(webhookSecretKey string) ServerOption {
	return func(s *Server) {
		s.webhookSecretKey = webhookSecretKey
	}
}

// WithEventHandler sets Server's eventHandler
func WithEventHandler(eventHandler EventHandler) ServerOption {
	return func(s *Server) {
		s.eventHandler = eventHandler
	}
}
