package log

import (
	"github.com/arangodb/go-driver"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Report(err error, msg string, a ...interface{}) error {
	log.Error().Err(err).Msgf(msg, a...)
	return status.Error(codes.Internal, "internal error")
}

func ArangoMeta(meta driver.DocumentMeta, msg string) {
	log.Info().Str("id", meta.ID.String()).Str("rev", meta.Rev).Str("old_rev", meta.OldRev).Msg(msg)
}
