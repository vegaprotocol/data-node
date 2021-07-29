package checks

import (
	"errors"
	"fmt"

	"code.vegaprotocol.io/data-node/logging"
	types "code.vegaprotocol.io/protos/vega"
)

type Assets interface {
	IsEnabled(asset string) bool
}

func MarginScalingFactor() func(interface{}) error {
	return func(v interface{}) error {
		sf := v.(*types.ScalingFactors)
		if err := sf.Validate(); err != nil {
			return err
		}
		if sf.SearchLevel >= sf.InitialMargin || sf.InitialMargin >= sf.CollateralRelease {
			return errors.New("invalid scaling factors (searchLevel < initialMargin < collateralRelease)")
		}
		return nil
	}
}

func GovernanceAssetUpdate(
	log *logging.Logger,
	assets Assets,
) func(value string) error {
	return func(value string) error {
		if !assets.IsEnabled(value) {
			log.Debug("tried to push a governance update with an non-enabled asset",
				logging.String("asset-id", value))
			return fmt.Errorf("invalid asset %v", value)
		}

		return nil
	}
}

func EthereumConfig() func(interface{}) error {
	return func(v interface{}) error {
		ecfg := v.(*types.EthereumConfig)
		if len(ecfg.NetworkId) <= 0 {
			return errors.New("missing ethereum config network id")
		}
		if len(ecfg.ChainId) <= 0 {
			return errors.New("missing ethereum config chain id")
		}
		if len(ecfg.BridgeAddress) <= 0 {
			return errors.New("missing ethereum config bridge address")
		}
		if ecfg.Confirmations == 0 {
			return errors.New("ethereum config confirmation must be > 0")
		}
		return nil
	}
}
