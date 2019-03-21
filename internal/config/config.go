// Package config provides objects containing configuration for specific parts of the applicaiton.
// It also encapsulates the logic needed to read the configuration from the environment.
package config

import (
	"context"
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"

	"github.com/google/go-github/v18/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zendesk/term-check/pkg/config"
)

const repoConfigFileLocation = "./.github/inclusive_lang.yaml"

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

// RepoConfig is an object holding all configuration values for one repo
// ignore - array of paths following `.gitignore` rules to ignore in the term check
type RepoConfig struct {
	Ignore []string `yaml:"ignore"`
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

// GetRepoConfig retreives the configuration for a repository
func GetRepoConfig(ctx context.Context, repo *github.Repository, head string, client *github.Client) *RepoConfig {
	config := RepoConfig{}
	var rawConfig string

	fc, _, resp, err := client.Repositories.GetContents(
		ctx,
		repo.GetOwner().GetLogin(),
		repo.GetName(),
		repoConfigFileLocation,
		&github.RepositoryContentGetOptions{Ref: head},
	)
	if err == nil {
		rawConfig, err = fc.GetContent()
	}

	// Store empty configuration if error or file is not there
	if err == nil && resp.StatusCode == http.StatusOK {
		yaml.Unmarshal([]byte(rawConfig), &config)
	}

	return &config
}

func panic(err error) {
	log.Panic().Err(err).Msg("Error encountered while parsing configuration")
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
