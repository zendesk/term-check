package github

import (
	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v18/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
)

// Client -
type Client struct {
	privateKeyPath string
	appID          int
}

// NewClient creates a new instance of Client, taking in Client options and creating a GitHub client
func NewClient(options ...ClientOption) *Client {
	zerolog.TimeFieldFormat = ""

	c := Client{}
	for _, option := range options {
		option(&c)
	}
	return &c
}

// CreateClient - TODO: rename this
func (c *Client) CreateClient(installationID int) *github.Client {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, c.appID, installationID, c.privateKeyPath)
	if err != nil {
		log.Fatal().Msg("Failed to parse private key from file.")
	}

	return github.NewClient(&http.Client{Transport: itr})
}
