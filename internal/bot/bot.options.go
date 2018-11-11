package bot

// Option an option function to customize Bot
type Option func(*Bot)

// WithPrivateKeyPath sets Bot's privateKeyPath
func WithPrivateKeyPath(privateKeyPath string) Option {
	return func(b *Bot) {
		b.privateKeyPath = privateKeyPath
	}
}

// WithAppID sets Bot's GitHub app id
func WithAppID(appID int) Option {
	return func(b *Bot) {
		b.appID = appID
	}
}

// WithWebhookSecretKey sets Bot's webhookSecretKey
func WithWebhookSecretKey(webhookSecretKey string) Option {
	return func(b *Bot) {
		b.webhookSecretKey = webhookSecretKey
	}
}
