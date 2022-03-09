package integration_test

import "testing"

func TestMarketBasic(t *testing.T) {
	query := "{ markets{ id, name, decimalPlaces, tradingMode, state } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestMarketFees(t *testing.T) {
	query := "{ markets{ fees { factors { makerFee, infrastructureFee, liquidityFee } } } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestTradableInstrument(t *testing.T) {
	query := "{ markets{ tradableInstrument{ instrument{ id, code, name, metadata{ tags } } } } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestMarginCalculator(t *testing.T) {
	query := "{ markets{ tradableInstrument{ marginCalculator{ scalingFactors{ searchLevel, initialMargin,collateralRelease } } } } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestPriceMonitoringSettings(t *testing.T) {
	query := "{ markets{ priceMonitoringSettings{ parameters{ triggers{ horizonSecs, probability } } } } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestLiquidityMonitoring(t *testing.T) {
	query := "{ markets{ liquidityMonitoringParameters{ targetStakeParameters{ timeWindow, scalingFactor } triggeringRatio} } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestProposal(t *testing.T) {
	query := "{ markets{ proposal{ id, reference, party { id }, state, datetime, rejectionReason} } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestProposalTerms(t *testing.T) {
	query := "{ markets{ proposal{ terms{ closingDatetime, enactmentDatetime } } } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestProposalYes(t *testing.T) {
	query := "{ markets{ proposal{ votes{ yes{ totalNumber totalWeight totalTokens} } } } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestProposalYesVotes(t *testing.T) {
	query := `
	{ markets{ proposal{ votes{ yes{ votes{
		value, party { id }, datetime, proposalId, governanceTokenBalance, governanceTokenWeight
	} } } } } }`
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestProposalNo(t *testing.T) {
	query := "{ markets{ proposal{ votes{ no{ totalNumber totalWeight totalTokens} } } } }"
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestProposalNoVotes(t *testing.T) {
	query := `
	{ markets{ proposal{ votes{ no{ votes{
		value, party { id }, datetime, proposalId, governanceTokenBalance, governanceTokenWeight
	} } } } } }`
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestMarketOrders(t *testing.T) {
	query := `
	{ markets{ orders {
		id, price, side, timeInForce, size, remaining,
		status, reference, type, rejectionReason, version,
		party{id}, market{id}, trades{id}
		createdAt, expiresAt,  updatedAt,
		peggedOrder { reference, offset },
	} } }
	`
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestMarketOrdersLiquidityProvision(t *testing.T) {
	query := `
	{ markets{ orders{ liquidityProvision{
		commitmentAmount, fee, status, version, reference
		createdAt, updatedAt
		market {id}
		# TODO: sells buys
	} } } }
	`
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestMarketTrades(t *testing.T) {
	query := `
	{markets { trades {
		id, price, size, createdAt, market{id}, type
		buyOrder, sellOrder,
		buyer{id}, seller{id}, aggressor
		buyerFee {makerFee, infrastructureFee, liquidityFee }
		sellerFee {makerFee, infrastructureFee, liquidityFee }
		buyerAuctionBatch, sellerAuctionBatch
	}}}
	`
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

func TestMarketAccounts(t *testing.T) {
	query := `
	{ markets{ accounts {
		balance
		# asset {id}   # TODO - put back once this doesn't break
		type
		market {id}
	} } }
	`
	var new, old struct{ Markets []Market }
	assertGraphQLQueriesReturnSame(t, query, &new, &old)
}

// TODO: Market depth / data / candles / liquidity provisions / timestamps
