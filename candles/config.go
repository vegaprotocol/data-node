package candles

import (
	"code.vegaprotocol.io/data-node/config/encoding"
	"code.vegaprotocol.io/data-node/logging"
)

// namedLogger is the identifier for package and should ideally match the package name
// this is simply emitted as a hierarchical label e.g. 'api.grpc'.
const namedLogger = "candles"

// Config represent the configuration of the candle package
type Config struct {
	Level                       encoding.LogLevel `long:"log-level"`
	CandleEventStreamBufferSize int               `long:"candle-event-stream-buffer-size" description:"buffer size used by the candle events stream for the per client per candle channel"`
	PeriodFetchSize             int               `long:"period-fetch-size" description:"the number of forward candle periods to fetch"`
}

// NewDefaultConfig creates an instance of the package specific configuration, given a
// pointer to a logger instance to be used for logging within the package.
func NewDefaultConfig() Config {
	return Config{
		Level:                       encoding.LogLevel{Level: logging.InfoLevel},
		CandleEventStreamBufferSize: 100,
		PeriodFetchSize:             100,
	}
}
