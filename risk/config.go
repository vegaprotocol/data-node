package risk

import (
	"code.vegaprotocol.io/data-node/config/encoding"
	"code.vegaprotocol.io/data-node/logging"
)

// namedLogger is the identifier for package and should ideally match the package name
// this is simply emitted as a hierarchical label e.g. 'api.grpc'.
const namedLogger = "risk"

// Config represent the configuration of the Risk engine
type Config struct {
	Level encoding.LogLevel `long:"log-level"`

	LogMarginUpdate encoding.Bool `long:"log-margin-update"`
}

// NewDefaultConfig creates an instance of the package specific configuration, given a
// pointer to a logger instance to be used for logging within the package.
func NewDefaultConfig() Config {
	return Config{
		Level:           encoding.LogLevel{Level: logging.InfoLevel},
		LogMarginUpdate: true,
	}
}
