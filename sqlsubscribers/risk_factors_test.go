package sqlsubscribers_test

import (
	"context"
	"testing"
	"time"

	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/sqlsubscribers"
	"code.vegaprotocol.io/data-node/sqlsubscribers/mocks"
	"code.vegaprotocol.io/vega/events"
	"code.vegaprotocol.io/vega/types"
	"code.vegaprotocol.io/vega/types/num"
	"github.com/golang/mock/gomock"
)

func TestRiskFactor_Push(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockRiskFactorStore(ctrl)

	store.EXPECT().Upsert(context.Background(), gomock.Any()).Times(1)
	subscriber := sqlsubscribers.NewRiskFactor(store, logging.NewTestLogger())
	subscriber.Push(context.Background(), events.NewTime(context.Background(), time.Now()))
	subscriber.Push(context.Background(), events.NewRiskFactorEvent(context.Background(), types.RiskFactor{
		Market: "deadbeef",
		Short:  num.DecimalFromInt64(1000),
		Long:   num.DecimalFromInt64(1000),
	}))
}
