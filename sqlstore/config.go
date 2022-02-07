package sqlstore

import (
	"code.vegaprotocol.io/data-node/config/encoding"
)

type Config struct {
	Enabled       bool
	Host          string
	Port          int
	Username      string
	Password      string
	Database      string
	WipeOnStartup bool
	Level         encoding.LogLevel `long:"log-level"`
}

func NewDefaultConfig() Config {
	return Config{
		Enabled:       false,
		Host:          "localhost",
		Port:          5432,
		Username:      "vega",
		Password:      "vega",
		Database:      "vega",
		WipeOnStartup: true,
	}
}
