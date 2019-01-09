package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LookupEnv to allow mocking of os.LookupEnv
type LookupEnv func(string) (string, bool)

// FatalLog to allow mocking of log.Fatal()
type FatalLog func(format string, a ...interface{})

// Config has custom configuration methods
type Config struct {
	osEnv LookupEnv
	fatal FatalLog
}

// New instantiates the Config object
func New(options ...Option) *Config {
	c := Config{
		osEnv: os.LookupEnv,
		fatal: log.Fatal().Msgf,
	}
	for i := range options {
		options[i](&c)
	}
	return &c
}

// Env looks env value for passed in key, logging and failing if not set
func (c *Config) Env(name string, fallback string) string {
	zerolog.TimeFieldFormat = ""

	v, ok := c.osEnv(name)
	if !ok {
		if fallback != "" {
			return fallback
		}
		c.fatal("Environment variable is not set: %s", name)
	}
	return v
}
