package governance_test

import (
	"context"
	"testing"

	"code.vegaprotocol.io/data-node/governance"
	"code.vegaprotocol.io/data-node/governance/mocks"
	"code.vegaprotocol.io/data-node/logging"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testSvc struct {
	*governance.Svc
	ctrl  *gomock.Controller
	ctx   context.Context
	cfunc context.CancelFunc

	bus   *mocks.MockEventBus
	gov   *mocks.MockGovernanceDataSub
	votes *mocks.MockVoteSub
}

func newTestService(t *testing.T) *testSvc {
	ctrl := gomock.NewController(t)
	bus := mocks.NewMockEventBus(ctrl)
	gov := mocks.NewMockGovernanceDataSub(ctrl)
	votes := mocks.NewMockVoteSub(ctrl)

	ctx, cfunc := context.WithCancel(context.Background())

	result := &testSvc{
		ctrl:  ctrl,
		ctx:   ctx,
		cfunc: cfunc,
		bus:   bus,
		gov:   gov,
		votes: votes,
	}
	result.Svc = governance.NewService(logging.NewTestLogger(), governance.NewDefaultConfig(), bus, gov, votes)
	assert.NotNil(t, result.Svc)
	return result
}

func TestGovernanceService(t *testing.T) {
	svc := newTestService(t)
	defer svc.ctrl.Finish()

	cfg := svc.Config
	cfg.Level.Level = logging.DebugLevel
	svc.ReloadConf(cfg)
	assert.Equal(t, svc.Config.Level.Level, logging.DebugLevel)

	cfg.Level.Level = logging.InfoLevel
	svc.ReloadConf(cfg)
	assert.Equal(t, svc.Config.Level.Level, logging.InfoLevel)
}
