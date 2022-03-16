package integration_test

import "testing"

const (
	partyDepositsQuery = `{ 
	parties {
		id
		deposits{
		  id,
		  party {
			id
		  }
		  amount
		  asset {
			id
		  }
		  status
		  createdTimestamp
		  creditedTimestamp
		  txHash
		}
	} 
}`
)

func TestPartyAPIs(t *testing.T) {
	testCases := []struct {
		name  string
		query string
	}{
		{
			name:  "Party Deposits",
			query: partyDepositsQuery,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var new, old struct{ Party []Party }
			assertGraphQLQueriesReturnSame(tt, tc.query, &new, &old)
		})
	}
}
