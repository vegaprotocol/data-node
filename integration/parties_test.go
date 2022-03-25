package integration_test

import "testing"

func TestParties(t *testing.T) {
	queries := map[string]string{
		"Deposits":      "{ parties { id deposits{ id, party { id }, amount, asset { id }, status, createdTimestamp, creditedTimestamp, txHash } } }",
		"Withdrawals":   "{ parties { id withdrawals { id, party { id }, amount, asset { id }, status, ref, expiry, txHash, createdTimestamp, withdrawnTimestamp } } }",
		"Delegations":   "{ parties{ id delegations{ node { id }, party{ id }, epoch, amount } } }",
		"Proposals":     "{ parties{ id proposals{ id votes{ yes { totalNumber } no { totalNumber } } } } }",
		"Votes":         "{ parties{ id votes{ proposalId vote{ value } } } }",
		"Margin Levels": "{ parties { id margins { market { id }, asset { id }, party { id }, maintenanceLevel, searchLevel, initialLevel, collateralReleaseLevel, timestamp } } }",
	}

	for name, query := range queries {
		t.Run(name, func(t *testing.T) {
			var new, old struct{ Parties []Party }
			assertGraphQLQueriesReturnSame(t, query, &new, &old)
		})
	}
}
