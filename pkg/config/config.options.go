package config

// Option represents an option to config object
type Option func(*Config)

// WithLookupEnv assigns lookupEnv function to config object
func WithLookupEnv(l LookupEnv) Option {
	return func(c *Config) {
		c.osEnv = l
	}
}

// WithFatalLog assigns a fatal logging function to config object
func WithFatalLog(l FatalLog) Option {
	return func(c *Config) {
		c.fatal = l
	}
}
