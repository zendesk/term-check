// Package config provides objects containing configuration for specific parts of the applicaiton.
// It also encapsulates the logic needed to read the configuration from the environment.
package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/zendesk/term-check/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO: write Unmarshal() to require values

// BotConfig holds all config values necessary for the BotConfig
type BotConfig struct {
	AppID               int      `yaml:"appID"`
	TermList            []string `yaml:"termList"`
	CheckName           string   `yaml:"checkName"`
	CheckSuccessSummary string   `yaml:"checkSuccessSummary"`
	CheckFailureSummary string   `yaml:"checkFailureSummary"`
	CheckDetails        string   `yaml:"checkDetails"`
	AnnotationTitle     string   `yaml:"annotationTitle"`
	AnnotationBody      string   `yaml:"annotationBody"`
}

// ClientConfig holds all config values necessary for the client
type ClientConfig struct {
	AppID          int    `yaml:"appID"`
	PrivateKeyPath string `yaml:"privateKeyPath"`
}

// ServerConfig holds all config values necessary for the server
type ServerConfig struct {
	WebhookSecretKey string `yaml:"webhookSecretKey"`
}

// Config holds all config values for the applicaiton, separated by module
type Config struct {
	ForBot     *BotConfig
	ForClient  *ClientConfig
	ForServer  *ServerConfig
	configUtil *config.Config
	secretHash map[string]string
}

// New instantiates the Config object with configuration values from the environment for the BotConfig, client, and
// server
func New(configFilepath string) *Config {
	zerolog.TimeFieldFormat = ""

	sh, err := config.Secrets()
	if err != nil {
		panic(err)
	}

	c := Config{
		configUtil: config.New(),
		secretHash: sh,
	}

	config, err := ioutil.ReadFile(configFilepath)
	if err != nil {
		panic(err)
	}

	bc, err := c.getBotConfig(config)
	if err != nil {
		panic(err)
	}

	cc, err := c.getClientConfig(config)
	if err != nil {
		panic(err)
	}

	sc, err := c.getServerConfig(config)
	if err != nil {
		panic(err)
	}

	return &Config{
		ForBot:    bc,
		ForClient: cc,
		ForServer: sc,
	}
}

func (c *Config) getBotConfig(config []byte) (*BotConfig, error) {
	type driver struct {
		B BotConfig `yaml:"botConfig"`
	}

	d := driver{}
	yaml.Unmarshal(config, &d)
	bc := d.B

	if len(bc.TermList) == 0 {
		return &BotConfig{}, errors.New("TERM_LIST must contain at least one item")
	}

	return &bc, nil
}

func (c *Config) getClientConfig(config []byte) (*ClientConfig, error) {
	type driver struct {
		C ClientConfig `yaml:"clientConfig"`
	}

	d := driver{}
	yaml.Unmarshal(config, &d)
	cc := d.C

	return &cc, nil
}

func (c *Config) getServerConfig(config []byte) (*ServerConfig, error) {
	ws, ok := c.secretHash["WEBHOOK_SECRET_KEY"]

	if ok != true {
		return &ServerConfig{}, errors.New("WEBHOOK_SECRET_KEY not present in secrets hash")
	}

	return &ServerConfig{
		WebhookSecretKey: ws,
	}, nil
}

func panic(err error) {
	log.Panic().Err(err).Msg("Error encountered while parsing configuration")
}
