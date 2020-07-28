package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rverst/zgelf"
	"os"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
	}

	t, err := zgelf.NewUdpTransport("192.168.0.11", 12345)
	if err != nil {
		log.Fatal().Err(err).Msg("can not initialize udpTransport")
	}

	gWriter := zgelf.New("W-RV", "", t)
	defer gWriter.Close()

	multi := zerolog.MultiLevelWriter(consoleWriter, gWriter)

	log.Logger = zerolog.New(multi) //.With().Caller().Timestamp().Str("source", "example.org").Logger()
	log.Info().Msg("Hello World")
}
