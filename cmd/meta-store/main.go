package main

import (
	"github.com/geomodular/meta-store/pkg/server"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	if err := run(); err != nil {
		log.Error().Err(err).Msg("")
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	log.Info().Msg("starting server")
	if err := server.Run(); err != nil {
		return errors.Wrap(err, "failed running servers")
	}
	return nil
}
