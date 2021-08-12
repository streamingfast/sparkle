


SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


CREATE SCHEMA sgd3;


ALTER SCHEMA sgd3 OWNER TO graph;

SET default_tablespace = '';

SET default_table_access_method = heap;


CREATE TABLE sgd3.bundle (
    id text NOT NULL,
    eth_price numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.bundle OWNER TO graph;


CREATE SEQUENCE sgd3.bundle_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.bundle_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.bundle_vid_seq OWNED BY sgd3.bundle.vid;



CREATE TABLE sgd3.burn (
    id text NOT NULL,
    transaction text NOT NULL,
    "timestamp" numeric NOT NULL,
    pair text NOT NULL,
    liquidity numeric NOT NULL,
    sender bytea,
    amount_0 numeric,
    amount_1 numeric,
    "to" bytea,
    log_index numeric,
    amount_usd numeric,
    needs_complete boolean NOT NULL,
    fee_to bytea,
    fee_liquidity numeric,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.burn OWNER TO graph;


CREATE SEQUENCE sgd3.burn_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.burn_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.burn_vid_seq OWNED BY sgd3.burn.vid;



CREATE TABLE sgd3.cursor (
    id integer NOT NULL,
    cursor text
);


ALTER TABLE sgd3.cursor OWNER TO graph;


CREATE TABLE sgd3.mint (
    id text NOT NULL,
    transaction text NOT NULL,
    "timestamp" numeric NOT NULL,
    pair text NOT NULL,
    "to" bytea NOT NULL,
    liquidity numeric NOT NULL,
    sender bytea,
    amount_0 numeric,
    amount_1 numeric,
    log_index numeric,
    amount_usd numeric,
    fee_to bytea,
    fee_liquidity numeric,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.mint OWNER TO graph;


CREATE SEQUENCE sgd3.mint_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.mint_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.mint_vid_seq OWNED BY sgd3.mint.vid;



CREATE TABLE sgd3.pair (
    id text NOT NULL,
    token_0 text NOT NULL,
    token_1 text NOT NULL,
    reserve_0 numeric NOT NULL,
    reserve_1 numeric NOT NULL,
    total_supply numeric NOT NULL,
    reserve_eth numeric NOT NULL,
    reserve_usd numeric NOT NULL,
    tracked_reserve_eth numeric NOT NULL,
    token_0_price numeric NOT NULL,
    token_1_price numeric NOT NULL,
    volume_token_0 numeric NOT NULL,
    volume_token_1 numeric NOT NULL,
    volume_usd numeric NOT NULL,
    untracked_volume_usd numeric NOT NULL,
    tx_count numeric NOT NULL,
    created_at_timestamp numeric NOT NULL,
    created_at_block_number numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
)
WITH (autovacuum_vacuum_scale_factor='0.001', autovacuum_analyze_scale_factor='0.001');
ALTER TABLE ONLY sgd3.pair ALTER COLUMN tracked_reserve_eth SET STATISTICS 1000;


ALTER TABLE sgd3.pair OWNER TO graph;


CREATE TABLE sgd3.pair_day_data (
    id text NOT NULL,
    date integer NOT NULL,
    pair_address bytea NOT NULL,
    token_0 text NOT NULL,
    token_1 text NOT NULL,
    reserve_0 numeric NOT NULL,
    reserve_1 numeric NOT NULL,
    total_supply numeric NOT NULL,
    reserve_usd numeric NOT NULL,
    daily_volume_token_0 numeric NOT NULL,
    daily_volume_token_1 numeric NOT NULL,
    daily_volume_usd numeric NOT NULL,
    daily_txns numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.pair_day_data OWNER TO graph;


CREATE SEQUENCE sgd3.pair_day_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.pair_day_data_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.pair_day_data_vid_seq OWNED BY sgd3.pair_day_data.vid;



CREATE TABLE sgd3.pair_hour_data (
    id text NOT NULL,
    hour_start_unix integer NOT NULL,
    pair text NOT NULL,
    reserve_0 numeric NOT NULL,
    reserve_1 numeric NOT NULL,
    reserve_usd numeric NOT NULL,
    hourly_volume_token_0 numeric NOT NULL,
    hourly_volume_token_1 numeric NOT NULL,
    hourly_volume_usd numeric NOT NULL,
    hourly_txns numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.pair_hour_data OWNER TO graph;


CREATE SEQUENCE sgd3.pair_hour_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.pair_hour_data_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.pair_hour_data_vid_seq OWNED BY sgd3.pair_hour_data.vid;



CREATE TABLE sgd3.pair_ref (
    id text NOT NULL,
    pair_id text NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.pair_ref OWNER TO graph;


CREATE SEQUENCE sgd3.pair_ref_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.pair_ref_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.pair_ref_vid_seq OWNED BY sgd3.pair_ref.vid;



CREATE TABLE sgd3.pair_test (
    id text NOT NULL,
    token_0 text NOT NULL,
    token_1 text NOT NULL,
    reserve_0 bigint NOT NULL,
    reserve_1 bigint NOT NULL,
    total_supply numeric NOT NULL,
    reserve_eth numeric NOT NULL,
    reserve_usd numeric NOT NULL,
    tracked_reserve_eth numeric NOT NULL,
    token_0_price numeric NOT NULL,
    token_1_price numeric NOT NULL,
    volume_token_0 numeric NOT NULL,
    volume_token_1 numeric NOT NULL,
    volume_usd numeric NOT NULL,
    untracked_volume_usd numeric NOT NULL,
    tx_count numeric NOT NULL,
    created_at_timestamp numeric NOT NULL,
    created_at_block_number numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
)
WITH (autovacuum_vacuum_scale_factor='0.001', autovacuum_analyze_scale_factor='0.001');


ALTER TABLE sgd3.pair_test OWNER TO graph;


CREATE SEQUENCE sgd3.pair_test_reserve_0_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.pair_test_reserve_0_seq OWNER TO graph;


ALTER SEQUENCE sgd3.pair_test_reserve_0_seq OWNED BY sgd3.pair_test.reserve_0;



CREATE SEQUENCE sgd3.pair_test_reserve_1_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.pair_test_reserve_1_seq OWNER TO graph;


ALTER SEQUENCE sgd3.pair_test_reserve_1_seq OWNED BY sgd3.pair_test.reserve_1;



CREATE SEQUENCE sgd3.pair_test_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.pair_test_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.pair_test_vid_seq OWNED BY sgd3.pair_test.vid;



CREATE SEQUENCE sgd3.pair_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.pair_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.pair_vid_seq OWNED BY sgd3.pair.vid;



CREATE TABLE sgd3."poi2$" (
    digest bytea NOT NULL,
    id text NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3."poi2$" OWNER TO graph;


CREATE SEQUENCE sgd3."poi2$_vid_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3."poi2$_vid_seq" OWNER TO graph;


ALTER SEQUENCE sgd3."poi2$_vid_seq" OWNED BY sgd3."poi2$".vid;



CREATE TABLE sgd3.swap (
    id text NOT NULL,
    transaction text NOT NULL,
    "timestamp" numeric NOT NULL,
    pair text NOT NULL,
    sender bytea NOT NULL,
    "from" bytea NOT NULL,
    amount_0_in numeric NOT NULL,
    amount_1_in numeric NOT NULL,
    amount_0_out numeric NOT NULL,
    amount_1_out numeric NOT NULL,
    "to" bytea NOT NULL,
    log_index numeric,
    amount_usd numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.swap OWNER TO graph;


CREATE SEQUENCE sgd3.swap_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.swap_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.swap_vid_seq OWNED BY sgd3.swap.vid;



CREATE TABLE sgd3.token (
    id text NOT NULL,
    symbol text NOT NULL,
    name text NOT NULL,
    decimals numeric NOT NULL,
    total_supply numeric NOT NULL,
    trade_volume numeric NOT NULL,
    trade_volume_usd numeric NOT NULL,
    untracked_volume_usd numeric NOT NULL,
    tx_count numeric NOT NULL,
    total_liquidity numeric NOT NULL,
    derived_eth numeric,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.token OWNER TO graph;


CREATE TABLE sgd3.token_address (
    id text NOT NULL,
    vid bigint NOT NULL
);


ALTER TABLE sgd3.token_address OWNER TO graph;


CREATE SEQUENCE sgd3.token_address_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.token_address_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.token_address_vid_seq OWNED BY sgd3.token_address.vid;



CREATE TABLE sgd3.token_day_data (
    id text NOT NULL,
    date integer NOT NULL,
    token text NOT NULL,
    daily_volume_token numeric NOT NULL,
    daily_volume_eth numeric NOT NULL,
    daily_volume_usd numeric NOT NULL,
    daily_txns numeric NOT NULL,
    total_liquidity_token numeric NOT NULL,
    total_liquidity_eth numeric NOT NULL,
    total_liquidity_usd numeric NOT NULL,
    price_usd numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.token_day_data OWNER TO graph;


CREATE SEQUENCE sgd3.token_day_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.token_day_data_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.token_day_data_vid_seq OWNED BY sgd3.token_day_data.vid;



CREATE SEQUENCE sgd3.token_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.token_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.token_vid_seq OWNED BY sgd3.token.vid;



CREATE TABLE sgd3.transaction (
    id text NOT NULL,
    block_number numeric NOT NULL,
    "timestamp" numeric NOT NULL,
    mints text[] NOT NULL,
    burns text[] NOT NULL,
    swaps text[] NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.transaction OWNER TO graph;


CREATE SEQUENCE sgd3.transaction_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.transaction_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.transaction_vid_seq OWNED BY sgd3.transaction.vid;



CREATE TABLE sgd3.uniswap_day_data (
    id text NOT NULL,
    date integer NOT NULL,
    daily_volume_eth numeric NOT NULL,
    daily_volume_usd numeric NOT NULL,
    daily_volume_untracked numeric NOT NULL,
    total_volume_eth numeric NOT NULL,
    total_liquidity_eth numeric NOT NULL,
    total_volume_usd numeric NOT NULL,
    total_liquidity_usd numeric NOT NULL,
    tx_count numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.uniswap_day_data OWNER TO graph;


CREATE SEQUENCE sgd3.uniswap_day_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.uniswap_day_data_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.uniswap_day_data_vid_seq OWNED BY sgd3.uniswap_day_data.vid;



CREATE TABLE sgd3.uniswap_factory (
    id text NOT NULL,
    pair_count integer NOT NULL,
    total_volume_usd numeric NOT NULL,
    total_volume_eth numeric NOT NULL,
    untracked_volume_usd numeric NOT NULL,
    total_liquidity_usd numeric NOT NULL,
    total_liquidity_eth numeric NOT NULL,
    tx_count numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);


ALTER TABLE sgd3.uniswap_factory OWNER TO graph;


CREATE SEQUENCE sgd3.uniswap_factory_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE sgd3.uniswap_factory_vid_seq OWNER TO graph;


ALTER SEQUENCE sgd3.uniswap_factory_vid_seq OWNED BY sgd3.uniswap_factory.vid;



ALTER TABLE ONLY sgd3.bundle ALTER COLUMN vid SET DEFAULT nextval('sgd3.bundle_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.burn ALTER COLUMN vid SET DEFAULT nextval('sgd3.burn_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.mint ALTER COLUMN vid SET DEFAULT nextval('sgd3.mint_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.pair ALTER COLUMN vid SET DEFAULT nextval('sgd3.pair_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.pair_day_data ALTER COLUMN vid SET DEFAULT nextval('sgd3.pair_day_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.pair_hour_data ALTER COLUMN vid SET DEFAULT nextval('sgd3.pair_hour_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.pair_ref ALTER COLUMN vid SET DEFAULT nextval('sgd3.pair_ref_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.pair_test ALTER COLUMN reserve_0 SET DEFAULT nextval('sgd3.pair_test_reserve_0_seq'::regclass);



ALTER TABLE ONLY sgd3.pair_test ALTER COLUMN reserve_1 SET DEFAULT nextval('sgd3.pair_test_reserve_1_seq'::regclass);



ALTER TABLE ONLY sgd3.pair_test ALTER COLUMN vid SET DEFAULT nextval('sgd3.pair_test_vid_seq'::regclass);



ALTER TABLE ONLY sgd3."poi2$" ALTER COLUMN vid SET DEFAULT nextval('sgd3."poi2$_vid_seq"'::regclass);



ALTER TABLE ONLY sgd3.swap ALTER COLUMN vid SET DEFAULT nextval('sgd3.swap_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.token ALTER COLUMN vid SET DEFAULT nextval('sgd3.token_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.token_address ALTER COLUMN vid SET DEFAULT nextval('sgd3.token_address_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.token_day_data ALTER COLUMN vid SET DEFAULT nextval('sgd3.token_day_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.transaction ALTER COLUMN vid SET DEFAULT nextval('sgd3.transaction_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.uniswap_day_data ALTER COLUMN vid SET DEFAULT nextval('sgd3.uniswap_day_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.uniswap_factory ALTER COLUMN vid SET DEFAULT nextval('sgd3.uniswap_factory_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.bundle
    ADD CONSTRAINT bundle_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.bundle
    ADD CONSTRAINT bundle_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.burn
    ADD CONSTRAINT burn_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.burn
    ADD CONSTRAINT burn_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.cursor
    ADD CONSTRAINT cursor_pkey PRIMARY KEY (id);



ALTER TABLE ONLY sgd3.mint
    ADD CONSTRAINT mint_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.mint
    ADD CONSTRAINT mint_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.pair_day_data
    ADD CONSTRAINT pair_day_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.pair_day_data
    ADD CONSTRAINT pair_day_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.pair_hour_data
    ADD CONSTRAINT pair_hour_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.pair_hour_data
    ADD CONSTRAINT pair_hour_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.pair
    ADD CONSTRAINT pair_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.pair
    ADD CONSTRAINT pair_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.pair_ref
    ADD CONSTRAINT pair_ref_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.pair_ref
    ADD CONSTRAINT pair_ref_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.pair_test
    ADD CONSTRAINT pair_test_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.pair_test
    ADD CONSTRAINT pair_test_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3."poi2$"
    ADD CONSTRAINT "poi2$_id_block_range_excl" EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3."poi2$"
    ADD CONSTRAINT "poi2$_pkey" PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.swap
    ADD CONSTRAINT swap_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.token_address
    ADD CONSTRAINT token_address_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.token_day_data
    ADD CONSTRAINT token_day_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.token_day_data
    ADD CONSTRAINT token_day_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.token
    ADD CONSTRAINT token_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.token
    ADD CONSTRAINT token_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.transaction
    ADD CONSTRAINT transaction_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.uniswap_day_data
    ADD CONSTRAINT uniswap_day_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.uniswap_day_data
    ADD CONSTRAINT uniswap_day_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.uniswap_factory
    ADD CONSTRAINT uniswap_factory_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.uniswap_factory
    ADD CONSTRAINT uniswap_factory_pkey PRIMARY KEY (vid);



CREATE INDEX attr_0_0_uniswap_factory_id ON sgd3.uniswap_factory USING btree (id);
CREATE INDEX attr_0_1_uniswap_factory_pair_count ON sgd3.uniswap_factory USING btree (pair_count);
CREATE INDEX attr_0_2_uniswap_factory_total_volume_usd ON sgd3.uniswap_factory USING btree (total_volume_usd);
CREATE INDEX attr_0_3_uniswap_factory_total_volume_eth ON sgd3.uniswap_factory USING btree (total_volume_eth);
CREATE INDEX attr_0_4_uniswap_factory_untracked_volume_usd ON sgd3.uniswap_factory USING btree (untracked_volume_usd);
CREATE INDEX attr_0_5_uniswap_factory_total_liquidity_usd ON sgd3.uniswap_factory USING btree (total_liquidity_usd);
CREATE INDEX attr_0_6_uniswap_factory_total_liquidity_eth ON sgd3.uniswap_factory USING btree (total_liquidity_eth);
CREATE INDEX attr_0_7_uniswap_factory_tx_count ON sgd3.uniswap_factory USING btree (tx_count);
CREATE INDEX attr_10_0_pair_hour_data_id ON sgd3.pair_hour_data USING btree (id);
CREATE INDEX attr_10_1_pair_hour_data_hour_start_unix ON sgd3.pair_hour_data USING btree (hour_start_unix);
CREATE INDEX attr_10_2_pair_hour_data_pair ON sgd3.pair_hour_data USING gist (pair, block_range);
CREATE INDEX attr_10_3_pair_hour_data_reserve_0 ON sgd3.pair_hour_data USING btree (reserve_0);
CREATE INDEX attr_10_4_pair_hour_data_reserve_1 ON sgd3.pair_hour_data USING btree (reserve_1);
CREATE INDEX attr_10_5_pair_hour_data_reserve_usd ON sgd3.pair_hour_data USING btree (reserve_usd);
CREATE INDEX attr_10_6_pair_hour_data_hourly_volume_token_0 ON sgd3.pair_hour_data USING btree (hourly_volume_token_0);
CREATE INDEX attr_10_7_pair_hour_data_hourly_volume_token_1 ON sgd3.pair_hour_data USING btree (hourly_volume_token_1);
CREATE INDEX attr_10_8_pair_hour_data_hourly_volume_usd ON sgd3.pair_hour_data USING btree (hourly_volume_usd);
CREATE INDEX attr_10_9_pair_hour_data_hourly_txns ON sgd3.pair_hour_data USING btree (hourly_txns);
CREATE INDEX attr_11_0_pair_day_data_id ON sgd3.pair_day_data USING btree (id);
CREATE INDEX attr_11_10_pair_day_data_daily_volume_token_1 ON sgd3.pair_day_data USING btree (daily_volume_token_1);
CREATE INDEX attr_11_11_pair_day_data_daily_volume_usd ON sgd3.pair_day_data USING btree (daily_volume_usd);
CREATE INDEX attr_11_12_pair_day_data_daily_txns ON sgd3.pair_day_data USING btree (daily_txns);
CREATE INDEX attr_11_1_pair_day_data_date ON sgd3.pair_day_data USING btree (date);
CREATE INDEX attr_11_2_pair_day_data_pair_address ON sgd3.pair_day_data USING btree (pair_address);
CREATE INDEX attr_11_3_pair_day_data_token_0 ON sgd3.pair_day_data USING gist (token_0, block_range);
CREATE INDEX attr_11_4_pair_day_data_token_1 ON sgd3.pair_day_data USING gist (token_1, block_range);
CREATE INDEX attr_11_5_pair_day_data_reserve_0 ON sgd3.pair_day_data USING btree (reserve_0);
CREATE INDEX attr_11_6_pair_day_data_reserve_1 ON sgd3.pair_day_data USING btree (reserve_1);
CREATE INDEX attr_11_7_pair_day_data_total_supply ON sgd3.pair_day_data USING btree (total_supply);
CREATE INDEX attr_11_8_pair_day_data_reserve_usd ON sgd3.pair_day_data USING btree (reserve_usd);
CREATE INDEX attr_11_9_pair_day_data_daily_volume_token_0 ON sgd3.pair_day_data USING btree (daily_volume_token_0);



CREATE INDEX attr_12_0_token_day_data_id ON sgd3.token_day_data USING btree (id);



CREATE INDEX attr_12_10_token_day_data_price_usd ON sgd3.token_day_data USING btree (price_usd);



CREATE INDEX attr_12_1_token_day_data_date ON sgd3.token_day_data USING btree (date);



CREATE INDEX attr_12_2_token_day_data_token ON sgd3.token_day_data USING gist (token, block_range);



CREATE INDEX attr_12_3_token_day_data_daily_volume_token ON sgd3.token_day_data USING btree (daily_volume_token);



CREATE INDEX attr_12_4_token_day_data_daily_volume_eth ON sgd3.token_day_data USING btree (daily_volume_eth);



CREATE INDEX attr_12_5_token_day_data_daily_volume_usd ON sgd3.token_day_data USING btree (daily_volume_usd);



CREATE INDEX attr_12_6_token_day_data_daily_txns ON sgd3.token_day_data USING btree (daily_txns);



CREATE INDEX attr_12_7_token_day_data_total_liquidity_token ON sgd3.token_day_data USING btree (total_liquidity_token);



CREATE INDEX attr_12_8_token_day_data_total_liquidity_eth ON sgd3.token_day_data USING btree (total_liquidity_eth);



CREATE INDEX attr_12_9_token_day_data_total_liquidity_usd ON sgd3.token_day_data USING btree (total_liquidity_usd);



CREATE INDEX "attr_13_0_poi2$_digest" ON sgd3."poi2$" USING btree (digest);



CREATE INDEX "attr_13_1_poi2$_id" ON sgd3."poi2$" USING btree ("left"(id, 256));



CREATE INDEX attr_1_0_token_address_id ON sgd3.token_address USING btree (id);



CREATE INDEX attr_1_0_token_id ON sgd3.token USING btree (id);



CREATE INDEX attr_1_10_token_derived_eth ON sgd3.token USING btree (derived_eth);



CREATE INDEX attr_1_1_token_symbol ON sgd3.token USING btree ("left"(symbol, 256));



CREATE INDEX attr_1_2_token_name ON sgd3.token USING btree ("left"(name, 256));



CREATE INDEX attr_1_3_token_decimals ON sgd3.token USING btree (decimals);



CREATE INDEX attr_1_4_token_total_supply ON sgd3.token USING btree (total_supply);



CREATE INDEX attr_1_5_token_trade_volume ON sgd3.token USING btree (trade_volume);



CREATE INDEX attr_1_6_token_trade_volume_usd ON sgd3.token USING btree (trade_volume_usd);



CREATE INDEX attr_1_7_token_untracked_volume_usd ON sgd3.token USING btree (untracked_volume_usd);



CREATE INDEX attr_1_8_token_tx_count ON sgd3.token USING btree (tx_count);



CREATE INDEX attr_1_9_token_total_liquidity ON sgd3.token USING btree (total_liquidity);



CREATE INDEX attr_2_0_pair_ref_id ON sgd3.pair_ref USING btree (id);



CREATE INDEX attr_2_1_pair_ref_pair_id ON sgd3.pair_ref USING btree ("left"(pair_id, 256));



CREATE INDEX attr_3_0_pair_id ON sgd3.pair USING btree (id);



CREATE INDEX attr_3_10_pair_token_1_price ON sgd3.pair USING btree (token_1_price);



CREATE INDEX attr_3_11_pair_volume_token_0 ON sgd3.pair USING btree (volume_token_0);



CREATE INDEX attr_3_12_pair_volume_token_1 ON sgd3.pair USING btree (volume_token_1);



CREATE INDEX attr_3_13_pair_volume_usd ON sgd3.pair USING btree (volume_usd);



CREATE INDEX attr_3_14_pair_untracked_volume_usd ON sgd3.pair USING btree (untracked_volume_usd);



CREATE INDEX attr_3_15_pair_tx_count ON sgd3.pair USING btree (tx_count);



CREATE INDEX attr_3_16_pair_created_at_timestamp ON sgd3.pair USING btree (created_at_timestamp);



CREATE INDEX attr_3_17_pair_created_at_block_number ON sgd3.pair USING btree (created_at_block_number);



CREATE INDEX attr_3_1_pair_token_0 ON sgd3.pair USING gist (token_0, block_range);



CREATE INDEX attr_3_2_pair_token_1 ON sgd3.pair USING gist (token_1, block_range);



CREATE INDEX attr_3_3_pair_reserve_0 ON sgd3.pair USING btree (reserve_0);



CREATE INDEX attr_3_4_pair_reserve_1 ON sgd3.pair USING btree (reserve_1);



CREATE INDEX attr_3_5_pair_total_supply ON sgd3.pair USING btree (total_supply);



CREATE INDEX attr_3_6_pair_reserve_eth ON sgd3.pair USING btree (reserve_eth);



CREATE INDEX attr_3_7_pair_reserve_usd ON sgd3.pair USING btree (reserve_usd);



CREATE INDEX attr_3_8_pair_tracked_reserve_eth ON sgd3.pair USING btree (tracked_reserve_eth);



CREATE INDEX attr_3_9_pair_token_0_price ON sgd3.pair USING btree (token_0_price);



CREATE INDEX attr_4_0_transaction_id ON sgd3.transaction USING btree (id);



CREATE INDEX attr_4_1_transaction_block_number ON sgd3.transaction USING btree (block_number);



CREATE INDEX attr_4_2_transaction_timestamp ON sgd3.transaction USING btree ("timestamp");



CREATE INDEX attr_4_3_transaction_mints ON sgd3.transaction USING gin (mints);



CREATE INDEX attr_4_4_transaction_burns ON sgd3.transaction USING gin (burns);



CREATE INDEX attr_4_5_transaction_swaps ON sgd3.transaction USING gin (swaps);



CREATE INDEX attr_5_0_mint_id ON sgd3.mint USING btree (id);



CREATE INDEX attr_5_10_mint_amount_usd ON sgd3.mint USING btree (amount_usd);



CREATE INDEX attr_5_11_mint_fee_to ON sgd3.mint USING btree (fee_to);



CREATE INDEX attr_5_12_mint_fee_liquidity ON sgd3.mint USING btree (fee_liquidity);



CREATE INDEX attr_5_1_mint_transaction ON sgd3.mint USING gist (transaction, block_range);



CREATE INDEX attr_5_2_mint_timestamp ON sgd3.mint USING btree ("timestamp");



CREATE INDEX attr_5_3_mint_pair ON sgd3.mint USING gist (pair, block_range);



CREATE INDEX attr_5_4_mint_to ON sgd3.mint USING btree ("to");



CREATE INDEX attr_5_5_mint_liquidity ON sgd3.mint USING btree (liquidity);



CREATE INDEX attr_5_6_mint_sender ON sgd3.mint USING btree (sender);



CREATE INDEX attr_5_7_mint_amount_0 ON sgd3.mint USING btree (amount_0);



CREATE INDEX attr_5_8_mint_amount_1 ON sgd3.mint USING btree (amount_1);



CREATE INDEX attr_5_9_mint_log_index ON sgd3.mint USING btree (log_index);



CREATE INDEX attr_6_0_burn_id ON sgd3.burn USING btree (id);



CREATE INDEX attr_6_10_burn_amount_usd ON sgd3.burn USING btree (amount_usd);



CREATE INDEX attr_6_11_burn_needs_complete ON sgd3.burn USING btree (needs_complete);



CREATE INDEX attr_6_12_burn_fee_to ON sgd3.burn USING btree (fee_to);



CREATE INDEX attr_6_13_burn_fee_liquidity ON sgd3.burn USING btree (fee_liquidity);



CREATE INDEX attr_6_1_burn_transaction ON sgd3.burn USING gist (transaction, block_range);



CREATE INDEX attr_6_2_burn_timestamp ON sgd3.burn USING btree ("timestamp");



CREATE INDEX attr_6_3_burn_pair ON sgd3.burn USING gist (pair, block_range);



CREATE INDEX attr_6_4_burn_liquidity ON sgd3.burn USING btree (liquidity);



CREATE INDEX attr_6_5_burn_sender ON sgd3.burn USING btree (sender);



CREATE INDEX attr_6_6_burn_amount_0 ON sgd3.burn USING btree (amount_0);



CREATE INDEX attr_6_7_burn_amount_1 ON sgd3.burn USING btree (amount_1);



CREATE INDEX attr_6_8_burn_to ON sgd3.burn USING btree ("to");



CREATE INDEX attr_6_9_burn_log_index ON sgd3.burn USING btree (log_index);



CREATE INDEX attr_7_0_swap_id ON sgd3.swap USING btree (id);



CREATE INDEX attr_7_10_swap_to ON sgd3.swap USING btree ("to");



CREATE INDEX attr_7_11_swap_log_index ON sgd3.swap USING btree (log_index);



CREATE INDEX attr_7_12_swap_amount_usd ON sgd3.swap USING btree (amount_usd);



CREATE INDEX attr_7_1_swap_transaction ON sgd3.swap USING gist (transaction, block_range);



CREATE INDEX attr_7_2_swap_timestamp ON sgd3.swap USING btree ("timestamp");



CREATE INDEX attr_7_3_swap_pair ON sgd3.swap USING gist (pair, block_range);



CREATE INDEX attr_7_4_swap_sender ON sgd3.swap USING btree (sender);



CREATE INDEX attr_7_5_swap_from ON sgd3.swap USING btree ("from");



CREATE INDEX attr_7_6_swap_amount_0_in ON sgd3.swap USING btree (amount_0_in);



CREATE INDEX attr_7_7_swap_amount_1_in ON sgd3.swap USING btree (amount_1_in);



CREATE INDEX attr_7_8_swap_amount_0_out ON sgd3.swap USING btree (amount_0_out);



CREATE INDEX attr_7_9_swap_amount_1_out ON sgd3.swap USING btree (amount_1_out);



CREATE INDEX attr_8_0_bundle_id ON sgd3.bundle USING btree (id);



CREATE INDEX attr_8_1_bundle_eth_price ON sgd3.bundle USING btree (eth_price);



CREATE INDEX attr_9_0_uniswap_day_data_id ON sgd3.uniswap_day_data USING btree (id);



CREATE INDEX attr_9_1_uniswap_day_data_date ON sgd3.uniswap_day_data USING btree (date);



CREATE INDEX attr_9_2_uniswap_day_data_daily_volume_eth ON sgd3.uniswap_day_data USING btree (daily_volume_eth);



CREATE INDEX attr_9_3_uniswap_day_data_daily_volume_usd ON sgd3.uniswap_day_data USING btree (daily_volume_usd);



CREATE INDEX attr_9_4_uniswap_day_data_daily_volume_untracked ON sgd3.uniswap_day_data USING btree (daily_volume_untracked);



CREATE INDEX attr_9_5_uniswap_day_data_total_volume_eth ON sgd3.uniswap_day_data USING btree (total_volume_eth);



CREATE INDEX attr_9_6_uniswap_day_data_total_liquidity_eth ON sgd3.uniswap_day_data USING btree (total_liquidity_eth);



CREATE INDEX attr_9_7_uniswap_day_data_total_volume_usd ON sgd3.uniswap_day_data USING btree (total_volume_usd);



CREATE INDEX attr_9_8_uniswap_day_data_total_liquidity_usd ON sgd3.uniswap_day_data USING btree (total_liquidity_usd);



CREATE INDEX attr_9_9_uniswap_day_data_tx_count ON sgd3.uniswap_day_data USING btree (tx_count);



CREATE INDEX brin_bundle ON sgd3.bundle USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_burn ON sgd3.burn USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_mint ON sgd3.mint USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_pair ON sgd3.pair USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_pair_day_data ON sgd3.pair_day_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_pair_hour_data ON sgd3.pair_hour_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_pair_ref ON sgd3.pair_ref USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX "brin_poi2$" ON sgd3."poi2$" USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_swap ON sgd3.swap USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_token ON sgd3.token USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_token_day_data ON sgd3.token_day_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_transaction ON sgd3.transaction USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_uniswap_day_data ON sgd3.uniswap_day_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_uniswap_factory ON sgd3.uniswap_factory USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX bundle_block_range_closed ON sgd3.bundle USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX burn_block_range_closed ON sgd3.burn USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX mint_block_range_closed ON sgd3.mint USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX pair_block_range_closed ON sgd3.pair USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX pair_day_data_block_range_closed ON sgd3.pair_day_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX pair_hour_data_block_range_closed ON sgd3.pair_hour_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX pair_ref_block_range_closed ON sgd3.pair_ref USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX "poi2$_block_range_closed" ON sgd3."poi2$" USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX swap_block_range_closed ON sgd3.swap USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX token_block_range_closed ON sgd3.token USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX token_day_data_block_range_closed ON sgd3.token_day_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX transaction_block_range_closed ON sgd3.transaction USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX uniswap_day_data_block_range_closed ON sgd3.uniswap_day_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX uniswap_factory_block_range_closed ON sgd3.uniswap_factory USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);




