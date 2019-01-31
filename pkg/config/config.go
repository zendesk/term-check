// Package config provides some basic helpers to read environment variables, as well as secrets that are set
// by Samson.
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

// Secrets returns a map of secret names and values
func Secrets() (map[string]string, error) {
	s, err := readSecrets("/secrets") // TODO: extract to flag?
	return s, err
}

func readSecrets(d string) (map[string]string, error) {
	s := make(map[string]string)
	files, err := ioutil.ReadDir(d)

	if err != nil {
		return s, err
	}

	for _, file := range files {
		n := file.Name()
		if strings.HasPrefix(n, ".") {
			continue
		}
		data, err := ioutil.ReadFile(filepath.Join(d, n))
		if err != nil {
			return s, err
		}
		s[n] = strings.TrimSpace(string(data))
	}

	return s, err
}
