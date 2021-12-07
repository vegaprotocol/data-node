package gql_test

import (
	"context"
	"testing"

	gql "code.vegaprotocol.io/data-node/gateway/graphql"
	"code.vegaprotocol.io/protos/vega"
	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	skip        *int
	first       *int
	last        *int
	expected    []*vega.RewardDetails
	description string
}

type rewards []*vega.RewardDetails

func TestPaginateRewards(t *testing.T) {
	r1 := &vega.RewardDetails{Amount: "1"}
	r2 := &vega.RewardDetails{Amount: "2"}
	r3 := &vega.RewardDetails{Amount: "3"}
	testRewards := rewards{r1, r2, r3}

	one, two, three, four := 1, 2, 3, 4

	tc1 := TestCase{first: &two, expected: rewards{r1, r2}, description: "First Two"}
	tc2 := TestCase{first: &two, skip: &one, expected: rewards{r2, r3}, description: "Skip one, take two"}
	tc3 := TestCase{first: &two, skip: &four, expected: rewards{}, description: "Skip past end"}
	tc4 := TestCase{first: &four, expected: rewards{r1, r2, r3}, description: "First > length"}
	tc5 := TestCase{skip: &three, expected: rewards{}, description: "Skip everything"}
	tc6 := TestCase{last: &two, expected: rewards{r2, r3}, description: "Last Two"}
	tc7 := TestCase{last: &one, skip: &one, expected: rewards{r2}, description: "Last but one"}
	tc8 := TestCase{last: &one, skip: &four, expected: rewards{}, description: "Skip before beginning"}
	tc9 := TestCase{last: &four, skip: &one, expected: rewards{r1, r2}, description: "Last before beginning"}

	cases := []TestCase{tc1, tc2, tc3, tc4, tc5, tc6, tc7, tc8, tc9}
	for _, tc := range cases {
		rewards := vega.RewardPerAssetDetail{Details: testRewards}
		resolver := gql.VegaResolverRoot{}
		rewardResolver := resolver.RewardPerAssetDetail()
		actual, err := rewardResolver.Rewards(context.Background(), &rewards, tc.skip, tc.first, tc.last)
		assert.NoError(t, err)
		assert.Equal(t, tc.expected, actual, tc.description)
	}
}
