package main

import (
	"strconv"

	"github.com/ragurney/term-check/internal/bot"
	"github.com/ragurney/term-check/pkg/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = ""

	c := config.New()

	id, err := strconv.Atoi(c.Env("APP_ID", ""))
	if err != nil {
		log.Fatal().Err(err)
	}
	pk := c.Env("PRIVATE_KEY_PATH", "")
	ws := c.Env("WEBHOOK_SECRET_KEY", "")

	log.Info().Msg("Starting service...")

	bot.New(
		bot.WithAppID(id),
		bot.WithPrivateKeyPath(pk),
		bot.WithWebhookSecretKey(ws),
	).Start()
}
