package node

import (
	"context"
	"fmt"

	"code.vegaprotocol.io/data-node/accounts"
	"code.vegaprotocol.io/data-node/assets"
	"code.vegaprotocol.io/data-node/broker"
	"code.vegaprotocol.io/data-node/candles"
	"code.vegaprotocol.io/data-node/collateral"
	"code.vegaprotocol.io/data-node/config"
	"code.vegaprotocol.io/data-node/governance"
	"code.vegaprotocol.io/data-node/liquidity"
	"code.vegaprotocol.io/data-node/logging"
	"code.vegaprotocol.io/data-node/markets"
	"code.vegaprotocol.io/data-node/netparams"
	"code.vegaprotocol.io/data-node/netparams/checks"
	"code.vegaprotocol.io/data-node/netparams/dispatch"
	"code.vegaprotocol.io/data-node/notary"
	"code.vegaprotocol.io/data-node/oracles"
	oracleAdaptors "code.vegaprotocol.io/data-node/oracles/adaptors"
	"code.vegaprotocol.io/data-node/orders"
	"code.vegaprotocol.io/data-node/parties"
	"code.vegaprotocol.io/data-node/plugins"
	"code.vegaprotocol.io/data-node/pprof"
	"code.vegaprotocol.io/data-node/proto"
	"code.vegaprotocol.io/data-node/risk"
	"code.vegaprotocol.io/data-node/stats"
	"code.vegaprotocol.io/data-node/storage"
	"code.vegaprotocol.io/data-node/subscribers"
	"code.vegaprotocol.io/data-node/trades"
	"code.vegaprotocol.io/data-node/transfers"
	"code.vegaprotocol.io/data-node/types"
	"code.vegaprotocol.io/data-node/vegatime"
	"github.com/cenkalti/backoff"
	"github.com/prometheus/common/log"
)

func (l *NodeCommand) persistentPre(args []string) (err error) {
	// this shouldn't happen...
	if l.cancel != nil {
		l.cancel()
	}
	// ensure we cancel the context on error
	defer func() {
		if err != nil {
			l.cancel()
		}
	}()
	l.ctx, l.cancel = context.WithCancel(context.Background())

	conf := l.cfgwatchr.Get()

	if flagProvided("--no-stores") {
		conf.StoresEnabled = false
	}

	// reload logger with the setup from configuration
	l.Log = logging.NewLoggerFromConfig(conf.Logging)

	if conf.Pprof.Enabled {
		l.Log.Info("vega is starting with pprof profile, this is not a recommended setting for production")
		l.pproffhandlr, err = pprof.New(l.Log, conf.Pprof)
		if err != nil {
			return
		}
		l.cfgwatchr.OnConfigUpdate(
			func(cfg config.Config) { l.pproffhandlr.ReloadConf(cfg.Pprof) },
		)
	}

	l.Log.Info("Starting Vega",
		logging.String("config-path", l.configPath),
		logging.String("version", l.Version),
		logging.String("version-hash", l.VersionHash))

	// this doesn't fail
	l.timeService = vegatime.New(l.conf.Time)

	// Set ulimits
	if err = l.SetUlimits(); err != nil {
		l.Log.Warn("Unable to set ulimits",
			logging.Error(err))
	} else {
		l.Log.Debug("Set ulimits",
			logging.Uint64("nofile", l.conf.UlimitNOFile))
	}

	l.stats = stats.New(l.Log, l.conf.Stats, l.Version, l.VersionHash)

	// set up storage, this should be persistent
	if err := l.setupStorages(); err != nil {
		return err
	}
	l.setupSubscibers()

	if !l.conf.StoresEnabled {
		l.Log.Info("node setted up without badger store support")
	} else {
		l.Log.Info("node setted up with badger store support")
	}

	return nil
}

func (l *NodeCommand) setupSubscibers() {
	l.transferSub = subscribers.NewTransferResponse(l.ctx, l.transferResponseStore, true)
	l.marketEventSub = subscribers.NewMarketEvent(l.ctx, l.conf.Subscribers, l.Log, false)
	l.orderSub = subscribers.NewOrderEvent(l.ctx, l.conf.Subscribers, l.Log, l.orderStore, true)
	l.accountSub = subscribers.NewAccountSub(l.ctx, l.accounts, true)
	l.partySub = subscribers.NewPartySub(l.ctx, l.partyStore, true)
	l.tradeSub = subscribers.NewTradeSub(l.ctx, l.tradeStore, true)
	l.marginLevelSub = subscribers.NewMarginLevelSub(l.ctx, l.riskStore, true)
	l.governanceSub = subscribers.NewGovernanceDataSub(l.ctx, true)
	l.voteSub = subscribers.NewVoteSub(l.ctx, false, true)
	l.marketDataSub = subscribers.NewMarketDataSub(l.ctx, l.marketDataStore, true)
	l.newMarketSub = subscribers.NewMarketSub(l.ctx, l.marketStore, true)
	l.marketUpdatedSub = subscribers.NewMarketUpdatedSub(l.ctx, l.marketStore, true)
	l.candleSub = subscribers.NewCandleSub(l.ctx, l.candleStore, true)
	l.marketDepthSub = subscribers.NewMarketDepthBuilder(l.ctx, l.Log, true)
	l.riskFactorSub = subscribers.NewRiskFactorSub(l.ctx, l.riskStore, true)
}

func (l *NodeCommand) setupStorages() (err error) {
	l.marketDataStore = storage.NewMarketData(l.Log, l.conf.Storage)
	l.riskStore = storage.NewRisks(l.Log, l.conf.Storage)

	// always enabled market,parties etc stores as they are in memory or boths use them
	if l.marketStore, err = storage.NewMarkets(l.Log, l.conf.Storage, l.cancel); err != nil {
		return
	}

	if l.partyStore, err = storage.NewParties(l.conf.Storage); err != nil {
		return
	}
	if l.transferResponseStore, err = storage.NewTransferResponses(l.Log, l.conf.Storage); err != nil {
		return
	}

	// if stores are not enabled, initialise the noop stores and do nothing else
	if !l.conf.StoresEnabled {
		l.orderStore = storage.NewNoopOrders(l.Log, l.conf.Storage)
		l.tradeStore = storage.NewNoopTrades(l.Log, l.conf.Storage)
		l.accounts = storage.NewNoopAccounts(l.Log, l.conf.Storage)
		l.candleStore = storage.NewNoopCandles(l.Log, l.conf.Storage)
		return
	}

	if l.candleStore, err = storage.NewCandles(l.Log, l.conf.Storage, l.cancel); err != nil {
		return
	}

	if l.orderStore, err = storage.NewOrders(l.Log, l.conf.Storage, l.cancel); err != nil {
		return
	}
	if l.tradeStore, err = storage.NewTrades(l.Log, l.conf.Storage, l.cancel); err != nil {
		return
	}
	if l.accounts, err = storage.NewAccounts(l.Log, l.conf.Storage, l.cancel); err != nil {
		return
	}

	l.cfgwatchr.OnConfigUpdate(
		func(cfg config.Config) { l.accounts.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.tradeStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.orderStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.candleStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.transferResponseStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.partyStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.riskStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.marketDataStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.marketStore.ReloadConf(cfg.Storage) },
		func(cfg config.Config) { l.stats.ReloadConf(cfg.Stats) },
	)

	return
}

func (l *NodeCommand) loadAsset(id string, v *proto.AssetDetails) error {
	aid, err := l.assets.NewAsset(id, types.AssetDetailsFromProto(v))
	if err != nil {
		return fmt.Errorf("error instanciating asset %v", err)
	}

	asset, err := l.assets.Get(aid)
	if err != nil {
		return fmt.Errorf("unable to get asset %v", err)
	}

	// just a simple backoff here
	err = backoff.Retry(
		func() error {
			err := asset.Validate()
			if !asset.IsValid() {
				return err
			}
			return nil
		},
		backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5),
	)
	if err != nil {
		return fmt.Errorf("unable to instantiate new asset err=%v, asset-source=%s", err, v.String())
	}
	if err := l.assets.Enable(aid); err != nil {
		l.Log.Error("invalid genesis asset",
			logging.String("asset-details", v.String()),
			logging.Error(err))
		return fmt.Errorf("unable to enable asset: %v", err)
	}

	assetD := asset.Type()
	if err := l.collateral.EnableAsset(context.Background(), *assetD); err != nil {
		return fmt.Errorf("unable to enable asset in collateral: %v", err)
	}

	l.Log.Info("new asset added successfully",
		logging.String("asset", asset.String()))

	// FIXME: this will be remove once we stop loading market from config
	// here we replace the mkts assets symbols with ids
	for _, v := range l.mktscfg {
		sym := v.TradableInstrument.Instrument.GetFuture().SettlementAsset
		if sym == assetD.Details.Symbol {
			v.TradableInstrument.Instrument.GetFuture().SettlementAsset = assetD.Id
		}
	}

	return nil
}

// we've already set everything up WRT arguments etc... just bootstrap the node
func (l *NodeCommand) preRun(_ []string) (err error) {
	// ensure that context is cancelled if we return an error here
	defer func() {
		if err != nil {
			l.cancel()
		}
	}()

	// plugins
	l.settlePlugin = plugins.NewPositions(l.ctx)
	l.notaryPlugin = plugins.NewNotary(l.ctx)
	l.assetPlugin = plugins.NewAsset(l.ctx)
	l.withdrawalPlugin = plugins.NewWithdrawal(l.ctx)
	l.depositPlugin = plugins.NewDeposit(l.ctx)
	l.netParamsService = netparams.NewService(l.ctx)
	l.liquidityService = liquidity.NewService(l.ctx, l.Log, l.conf.Liquidity)
	l.oracleService = oracles.NewService(l.ctx)

	l.broker = broker.New(l.ctx)
	l.broker.SubscribeBatch(
		l.marketEventSub, l.transferSub, l.orderSub, l.accountSub,
		l.partySub, l.tradeSub, l.marginLevelSub, l.governanceSub,
		l.voteSub, l.marketDataSub, l.notaryPlugin, l.settlePlugin,
		l.newMarketSub, l.assetPlugin, l.candleSub, l.withdrawalPlugin,
		l.depositPlugin, l.marketDepthSub, l.riskFactorSub, l.netParamsService,
		l.liquidityService, l.marketUpdatedSub, l.oracleService)

	now, _ := l.timeService.GetTimeNow()

	l.assets, err = assets.New(l.Log, l.conf.Assets, l.timeService)
	if err != nil {
		return err
	}

	l.collateral = collateral.New(l.Log, l.conf.Collateral, l.broker, now)
	l.oracle = oracles.NewEngine(l.Log, l.conf.Oracles, now, l.broker)
	l.timeService.NotifyOnTick(l.oracle.UpdateCurrentTime)
	l.oracleAdaptors = oracleAdaptors.New()

	l.netParams = netparams.New(l.Log, l.conf.NetworkParameters, l.broker)

	// TODO Remove governance engine
	l.governance, err = governance.NewEngine(l.Log, l.conf.Governance, l.collateral, l.broker, l.assets, nil, l.netParams, now)
	if err != nil {
		log.Error("unable to initialise governance", logging.Error(err))
		return err
	}

	// start services
	if l.candleService, err = candles.NewService(l.Log, l.conf.Candles, l.candleStore); err != nil {
		return
	}

	if l.orderService, err = orders.NewService(l.Log, l.conf.Orders, l.orderStore, l.timeService); err != nil {
		return
	}

	if l.tradeService, err = trades.NewService(l.Log, l.conf.Trades, l.tradeStore, l.settlePlugin); err != nil {
		return
	}
	if l.marketService, err = markets.NewService(l.Log, l.conf.Markets, l.marketStore, l.orderStore, l.marketDataStore, l.marketDepthSub); err != nil {
		return
	}
	l.riskService = risk.NewService(l.Log, l.conf.Risk, l.riskStore, l.marketStore, l.marketDataStore)
	l.governanceService = governance.NewService(l.Log, l.conf.Governance, l.broker, l.governanceSub, l.voteSub, l.netParams)

	// last assignment to err, no need to check here, if something went wrong, we'll know about it
	l.partyService, err = parties.NewService(l.Log, l.conf.Parties, l.partyStore)
	l.accountsService = accounts.NewService(l.Log, l.conf.Accounts, l.accounts)
	l.transfersService = transfers.NewService(l.Log, l.conf.Transfers, l.transferResponseStore)
	l.notaryService = notary.NewService(l.Log, l.conf.Notary, l.notaryPlugin)
	l.assetService = assets.NewService(l.Log, l.conf.Assets, l.assetPlugin)
	l.eventService = subscribers.NewService(l.broker)

	// setup config reloads for all engines / services /etc
	l.setupConfigWatchers()
	l.timeService.NotifyOnTick(l.cfgwatchr.OnTimeUpdate)

	// setup some network parameters runtime validations
	// and network parameters updates dispatches
	return l.setupNetParameters()
}

func (l *NodeCommand) setupNetParameters() error {
	// now we are going to setup some network parameters which can be done
	// through runtime checks
	// e.g: changing the governance asset require the Assets and Collateral engines, so we can ensure any changes there are made for a valid asset
	if err := l.netParams.AddRules(
		netparams.ParamStringRules(
			netparams.GovernanceVoteAsset,
			checks.GovernanceAssetUpdate(l.Log, l.assets, l.collateral),
		),
	); err != nil {
		return err
	}

	// now add some watcher for our netparams
	return l.netParams.Watch(
		netparams.WatchParam{
			Param:   netparams.GovernanceVoteAsset,
			Watcher: dispatch.GovernanceAssetUpdate(l.Log, l.assets),
		},
	)
}

func (l *NodeCommand) setupConfigWatchers() {
	l.cfgwatchr.OnConfigUpdate(
		func(cfg config.Config) { l.notary.ReloadConf(cfg.Notary) },
		func(cfg config.Config) { l.evtfwd.ReloadConf(cfg.EvtForward) },
		func(cfg config.Config) { l.assets.ReloadConf(cfg.Assets) },
		func(cfg config.Config) { l.banking.ReloadConf(cfg.Banking) },
		func(cfg config.Config) { l.governance.ReloadConf(cfg.Governance) },

		// services
		func(cfg config.Config) { l.candleService.ReloadConf(cfg.Candles) },
		func(cfg config.Config) { l.orderService.ReloadConf(cfg.Orders) },
		func(cfg config.Config) { l.liquidityService.ReloadConf(cfg.Liquidity) },
		func(cfg config.Config) { l.tradeService.ReloadConf(cfg.Trades) },
		func(cfg config.Config) { l.marketService.ReloadConf(cfg.Markets) },
		func(cfg config.Config) { l.riskService.ReloadConf(cfg.Risk) },
		func(cfg config.Config) { l.governanceService.ReloadConf(cfg.Governance) },
		func(cfg config.Config) { l.assetService.ReloadConf(cfg.Assets) },
		func(cfg config.Config) { l.notaryService.ReloadConf(cfg.Notary) },
		func(cfg config.Config) { l.transfersService.ReloadConf(cfg.Transfers) },
		func(cfg config.Config) { l.accountsService.ReloadConf(cfg.Accounts) },
		func(cfg config.Config) { l.partyService.ReloadConf(cfg.Parties) },
	)
}
