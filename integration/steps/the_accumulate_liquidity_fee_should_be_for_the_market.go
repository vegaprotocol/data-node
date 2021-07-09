package steps

import (
	"fmt"

	"code.vegaprotocol.io/data-node/integration/stubs"
)

func TheAccumulatedLiquidityFeesShouldBeForTheMarket(
	broker *stubs.BrokerStub,
	amountStr, market string,
) error {
	amount, err := U64(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount: %w", err)
	}

	acc, err := broker.GetMarketLiquidityFeePoolAccount(market)
	if err != nil {
		return err
	}

	if acc.Balance != amount {
		return errInvalidAmountInLiquidityFee(market, amount, acc.Balance)
	}

	return nil
}

func errInvalidAmountInLiquidityFee(market string, expected, got uint64) error {
	return fmt.Errorf("invalid amount in liquidity fee pool for market %s want %d got %d", market, expected, got)
}
