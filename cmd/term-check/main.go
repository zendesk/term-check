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
		log.Panic().Err(err)
	}

	pk := c.Env("PRIVATE_KEY_PATH", "")

	secrets, err := config.Secrets()

	if err != nil {
		log.Panic().Err(err)
	}

	ws, ok := secrets["WEBHOOK_SECRET_KEY"]

	if ok != true {
		log.Panic().Msg("Could not read PRIVATE_KEY_PATH from secrets.")
	}

	log.Info().Msg("Starting service...")

	bot.New(
		bot.WithAppID(id),
		bot.WithPrivateKeyPath(pk),
		bot.WithWebhookSecretKey(ws),
	).Start()
}
