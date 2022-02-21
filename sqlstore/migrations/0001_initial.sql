-- +goose Up
create table blocks
(
    vega_time     TIMESTAMP WITH TIME ZONE NOT NULL PRIMARY KEY,
    height        BIGINT                   NOT NULL,
    hash          BYTEA                    NOT NULL
);

create table assets
(
    id             BYTEA NOT NULL PRIMARY KEY,
    name           TEXT NOT NULL UNIQUE,
    symbol         TEXT NOT NULL UNIQUE,
    total_supply   NUMERIC(32, 0),
    decimals       INT,
    quantum        INT,
    source         TEXT,
    erc20_contract TEXT,
    vega_time      TIMESTAMP WITH TIME ZONE NOT NULL REFERENCES blocks (vega_time)
);

create table parties
(
    id        BYTEA NOT NULL PRIMARY KEY,
    vega_time TIMESTAMP WITH TIME ZONE REFERENCES blocks (vega_time)
);

create table accounts
(
    id        SERIAL PRIMARY KEY,
    party_id  BYTEA,
    asset_id  BYTEA                    NOT NULL REFERENCES assets (id),
    market_id BYTEA,
    type      INT,
    vega_time TIMESTAMP WITH TIME ZONE NOT NULL REFERENCES blocks(vega_time),

    UNIQUE(party_id, asset_id, market_id, type)
);

create table balances
(
    account_id INT                      NOT NULL REFERENCES accounts(id),
    vega_time  TIMESTAMP WITH TIME ZONE NOT NULL REFERENCES blocks(vega_time),
    balance    NUMERIC(32, 0)           NOT NULL,

    PRIMARY KEY(vega_time, account_id)
);

create table ledger
(
    id              SERIAL                   PRIMARY KEY,
    account_from_id INT                      NOT NULL REFERENCES accounts(id),
    account_to_id   INT                      NOT NULL REFERENCES accounts(id),
    quantity        NUMERIC(32, 0)           NOT NULL,
    vega_time       TIMESTAMP WITH TIME ZONE NOT NULL REFERENCES blocks(vega_time),
    transfer_time   TIMESTAMP WITH TIME ZONE NOT NULL,
    reference       TEXT,
    type            TEXT
);


-- Create a function that always returns the first non-NULL value:
CREATE OR REPLACE FUNCTION public.first_agg (anyelement, anyelement)
  RETURNS anyelement
  LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE AS
'SELECT $1';

-- Then wrap an aggregate around it:
CREATE AGGREGATE public.first (anyelement) (
  SFUNC    = public.first_agg
, STYPE    = anyelement
, PARALLEL = safe
);

-- Create a function that always returns the last non-NULL value:
CREATE OR REPLACE FUNCTION public.last_agg (anyelement, anyelement)
  RETURNS anyelement
  LANGUAGE sql IMMUTABLE STRICT PARALLEL SAFE AS
'SELECT $2';

-- Then wrap an aggregate around it:
CREATE AGGREGATE public.last (anyelement) (
  SFUNC    = public.last_agg
, STYPE    = anyelement
, PARALLEL = safe
);

drop type if exists auction_trigger_type;
create type auction_trigger_type as enum('AUCTION_TRIGGER_UNSPECIFIED', 'AUCTION_TRIGGER_BATCH', 'AUCTION_TRIGGER_OPENING', 'AUCTION_TRIGGER_PRICE', 'AUCTION_TRIGGER_LIQUIDITY');

drop type if exists market_trading_mode_type;
create type market_trading_mode_type as enum('TRADING_MODE_UNSPECIFIED', 'TRADING_MODE_CONTINUOUS', 'TRADING_MODE_BATCH_AUCTION', 'TRADING_MODE_OPENING_AUCTION', 'TRADING_MODE_MONITORING_AUCTION');

create table market_data (
    market bytea not null,
    market_timestamp timestamp with time zone not null,
    vega_time timestamp with time zone not null references blocks(vega_time),
    mark_price numeric(32),
    best_bid_price numeric(32),
    best_bid_volume bigint,
    best_offer_price numeric(32),
    best_offer_volume bigint,
    best_static_bid_price numeric(32),
    best_static_bid_volume bigint,
    best_static_offer_price numeric(32),
    best_static_offer_volume bigint,
    mid_price numeric(32),
    static_mid_price numeric(32),
    open_interest bigint,
    auction_end bigint,
    auction_start bigint,
    indicative_price numeric(32),
    indicative_volume bigint,
    market_trading_mode market_trading_mode_type,
    auction_trigger auction_trigger_type,
    extension_trigger auction_trigger_type,
    target_stake numeric(32),
    supplied_stake numeric(32),
    price_monitoring_bounds jsonb,
    market_value_proxy text,
    liquidity_provider_fee_shares jsonb,
    primary key (
        market,
        market_timestamp
    )
);

create or replace view market_data_snapshot as
with cte_market_data_latest(market, market_timestamp) as (
    select market, max(market_timestamp)
    from market_data
    group by market
)
select md.market, md.market_timestamp, vega_time, mark_price, best_bid_price, best_bid_volume, best_offer_price, best_offer_volume,
       best_static_bid_price, best_static_bid_volume, best_static_offer_price, best_static_offer_volume,
       mid_price, static_mid_price, open_interest, auction_end, auction_start, indicative_price, indicative_volume,
       market_trading_mode, auction_trigger, extension_trigger, target_stake, supplied_stake, price_monitoring_bounds,
       market_value_proxy, liquidity_provider_fee_shares
from market_data md
join cte_market_data_latest mx
on md.market = mx.market
and md.market_timestamp = mx.market_timestamp
;


-- +goose Down
DROP AGGREGATE public.first(anyelement);
DROP AGGREGATE public.last(anyelement);
DROP FUNCTION public.first_agg(anyelement, anyelement);
DROP FUNCTION public.last_agg(anyelement, anyelement);

drop view if exists market_data_snapshot;
drop table if exists market_data;
drop type if exists market_trading_mode_type;
drop type if exists auction_trigger_type;
drop table if exists ledger;
drop table if exists balances;
drop table if exists accounts;
drop table if exists parties;
drop table if exists assets;
drop table if exists blocks;
