package logger

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type Config struct {
	Level  string `toml:"level"`
	Output string `toml:"output"`
}

func New(cfg Config) zerolog.Logger {
	level, err := zerolog.ParseLevel(strings.ToLower(strings.TrimSpace(cfg.Level)))
	if err != nil {
		level = zerolog.InfoLevel
	}
	if level == zerolog.Disabled {
		return zerolog.Nop()
	}
	var output io.Writer
	switch strings.ToLower(strings.TrimSpace(cfg.Output)) {
	case "console":
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	case "stdout":
		output = os.Stdout
	default:
		output = os.Stdout
	}
	zerolog.DefaultContextLogger = nil
	zerolog.SetGlobalLevel(level)
	return zerolog.New(output).With().Timestamp().Logger()
}

// FromContext fetch logger from given context or return disabled logger if not exist.
func FromContext(ctx context.Context) *zerolog.Logger { return zerolog.Ctx(ctx) }
