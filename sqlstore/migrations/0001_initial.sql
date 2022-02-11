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

create table market_data (
    market text not null,
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
    market_trading_mode text,
    auction_trigger text,
    extension_trigger text,
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

-- +goose Down
drop table if exists market_data;
drop table if exists ledger;
drop table if exists accounts;
drop table if exists parties;
drop table if exists assets;
drop table if exists blocks;