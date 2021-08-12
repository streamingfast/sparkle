\c "graph-node";

CREATE SCHEMA sgd1;

CREATE TABLE sgd1.bundle (
    id text NOT NULL,
    eth_price numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);

CREATE SEQUENCE sgd1.bundle_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.bundle_vid_seq OWNED BY sgd1.bundle.vid;

CREATE TABLE sgd1.burn (
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



CREATE SEQUENCE sgd1.burn_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.burn_vid_seq OWNED BY sgd1.burn.vid;



CREATE TABLE sgd1.mint (
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



CREATE SEQUENCE sgd1.mint_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.mint_vid_seq OWNED BY sgd1.mint.vid;



CREATE TABLE sgd1.pair (
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
);



CREATE TABLE sgd1.pair_day_data (
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



CREATE SEQUENCE sgd1.pair_day_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.pair_day_data_vid_seq OWNED BY sgd1.pair_day_data.vid;



CREATE TABLE sgd1.pair_hour_data (
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



CREATE SEQUENCE sgd1.pair_hour_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.pair_hour_data_vid_seq OWNED BY sgd1.pair_hour_data.vid;



CREATE SEQUENCE sgd1.pair_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.pair_vid_seq OWNED BY sgd1.pair.vid;



CREATE TABLE sgd1."poi2$" (
    digest bytea NOT NULL,
    id text NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE SEQUENCE sgd1."poi2$_vid_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1."poi2$_vid_seq" OWNED BY sgd1."poi2$".vid;



CREATE TABLE sgd1.swap (
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



CREATE SEQUENCE sgd1.swap_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.swap_vid_seq OWNED BY sgd1.swap.vid;



CREATE TABLE sgd1.token (
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



CREATE TABLE sgd1.token_day_data (
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



CREATE SEQUENCE sgd1.token_day_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.token_day_data_vid_seq OWNED BY sgd1.token_day_data.vid;



CREATE SEQUENCE sgd1.token_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.token_vid_seq OWNED BY sgd1.token.vid;



CREATE TABLE sgd1.transaction (
    id text NOT NULL,
    block_number numeric NOT NULL,
    "timestamp" numeric NOT NULL,
    mints text[] NOT NULL,
    burns text[] NOT NULL,
    swaps text[] NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE SEQUENCE sgd1.transaction_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.transaction_vid_seq OWNED BY sgd1.transaction.vid;



CREATE TABLE sgd1.uniswap_day_data (
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



CREATE SEQUENCE sgd1.uniswap_day_data_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.uniswap_day_data_vid_seq OWNED BY sgd1.uniswap_day_data.vid;



CREATE TABLE sgd1.uniswap_factory (
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



CREATE SEQUENCE sgd1.uniswap_factory_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd1.uniswap_factory_vid_seq OWNED BY sgd1.uniswap_factory.vid;



ALTER TABLE ONLY sgd1.bundle ALTER COLUMN vid SET DEFAULT nextval('sgd1.bundle_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.burn ALTER COLUMN vid SET DEFAULT nextval('sgd1.burn_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.mint ALTER COLUMN vid SET DEFAULT nextval('sgd1.mint_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.pair ALTER COLUMN vid SET DEFAULT nextval('sgd1.pair_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.pair_day_data ALTER COLUMN vid SET DEFAULT nextval('sgd1.pair_day_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.pair_hour_data ALTER COLUMN vid SET DEFAULT nextval('sgd1.pair_hour_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd1."poi2$" ALTER COLUMN vid SET DEFAULT nextval('sgd1."poi2$_vid_seq"'::regclass);



ALTER TABLE ONLY sgd1.swap ALTER COLUMN vid SET DEFAULT nextval('sgd1.swap_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.token ALTER COLUMN vid SET DEFAULT nextval('sgd1.token_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.token_day_data ALTER COLUMN vid SET DEFAULT nextval('sgd1.token_day_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.transaction ALTER COLUMN vid SET DEFAULT nextval('sgd1.transaction_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.uniswap_day_data ALTER COLUMN vid SET DEFAULT nextval('sgd1.uniswap_day_data_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.uniswap_factory ALTER COLUMN vid SET DEFAULT nextval('sgd1.uniswap_factory_vid_seq'::regclass);



ALTER TABLE ONLY sgd1.bundle
    ADD CONSTRAINT bundle_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.bundle
    ADD CONSTRAINT bundle_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.burn
    ADD CONSTRAINT burn_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.burn
    ADD CONSTRAINT burn_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.mint
    ADD CONSTRAINT mint_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.mint
    ADD CONSTRAINT mint_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.pair_day_data
    ADD CONSTRAINT pair_day_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.pair_day_data
    ADD CONSTRAINT pair_day_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.pair_hour_data
    ADD CONSTRAINT pair_hour_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.pair_hour_data
    ADD CONSTRAINT pair_hour_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.pair
    ADD CONSTRAINT pair_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.pair
    ADD CONSTRAINT pair_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1."poi2$"
    ADD CONSTRAINT "poi2$_id_block_range_excl" EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1."poi2$"
    ADD CONSTRAINT "poi2$_pkey" PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.swap
    ADD CONSTRAINT swap_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.swap
    ADD CONSTRAINT swap_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.token_day_data
    ADD CONSTRAINT token_day_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.token_day_data
    ADD CONSTRAINT token_day_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.token
    ADD CONSTRAINT token_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.token
    ADD CONSTRAINT token_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.transaction
    ADD CONSTRAINT transaction_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.transaction
    ADD CONSTRAINT transaction_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.uniswap_day_data
    ADD CONSTRAINT uniswap_day_data_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.uniswap_day_data
    ADD CONSTRAINT uniswap_day_data_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd1.uniswap_factory
    ADD CONSTRAINT uniswap_factory_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd1.uniswap_factory
    ADD CONSTRAINT uniswap_factory_pkey PRIMARY KEY (vid);



CREATE INDEX attr_0_0_uniswap_factory_id ON sgd1.uniswap_factory USING btree (id);



CREATE INDEX attr_0_1_uniswap_factory_pair_count ON sgd1.uniswap_factory USING btree (pair_count);



CREATE INDEX attr_0_2_uniswap_factory_total_volume_usd ON sgd1.uniswap_factory USING btree (total_volume_usd);



CREATE INDEX attr_0_3_uniswap_factory_total_volume_eth ON sgd1.uniswap_factory USING btree (total_volume_eth);



CREATE INDEX attr_0_4_uniswap_factory_untracked_volume_usd ON sgd1.uniswap_factory USING btree (untracked_volume_usd);



CREATE INDEX attr_0_5_uniswap_factory_total_liquidity_usd ON sgd1.uniswap_factory USING btree (total_liquidity_usd);



CREATE INDEX attr_0_6_uniswap_factory_total_liquidity_eth ON sgd1.uniswap_factory USING btree (total_liquidity_eth);



CREATE INDEX attr_0_7_uniswap_factory_tx_count ON sgd1.uniswap_factory USING btree (tx_count);



CREATE INDEX attr_10_0_pair_day_data_id ON sgd1.pair_day_data USING btree (id);



CREATE INDEX attr_10_10_pair_day_data_daily_volume_token_1 ON sgd1.pair_day_data USING btree (daily_volume_token_1);



CREATE INDEX attr_10_11_pair_day_data_daily_volume_usd ON sgd1.pair_day_data USING btree (daily_volume_usd);



CREATE INDEX attr_10_12_pair_day_data_daily_txns ON sgd1.pair_day_data USING btree (daily_txns);



CREATE INDEX attr_10_1_pair_day_data_date ON sgd1.pair_day_data USING btree (date);



CREATE INDEX attr_10_2_pair_day_data_pair_address ON sgd1.pair_day_data USING btree (pair_address);



CREATE INDEX attr_10_3_pair_day_data_token_0 ON sgd1.pair_day_data USING gist (token_0, block_range);



CREATE INDEX attr_10_4_pair_day_data_token_1 ON sgd1.pair_day_data USING gist (token_1, block_range);



CREATE INDEX attr_10_5_pair_day_data_reserve_0 ON sgd1.pair_day_data USING btree (reserve_0);



CREATE INDEX attr_10_6_pair_day_data_reserve_1 ON sgd1.pair_day_data USING btree (reserve_1);



CREATE INDEX attr_10_7_pair_day_data_total_supply ON sgd1.pair_day_data USING btree (total_supply);



CREATE INDEX attr_10_8_pair_day_data_reserve_usd ON sgd1.pair_day_data USING btree (reserve_usd);



CREATE INDEX attr_10_9_pair_day_data_daily_volume_token_0 ON sgd1.pair_day_data USING btree (daily_volume_token_0);



CREATE INDEX attr_11_0_token_day_data_id ON sgd1.token_day_data USING btree (id);



CREATE INDEX attr_11_10_token_day_data_price_usd ON sgd1.token_day_data USING btree (price_usd);



CREATE INDEX attr_11_1_token_day_data_date ON sgd1.token_day_data USING btree (date);



CREATE INDEX attr_11_2_token_day_data_token ON sgd1.token_day_data USING gist (token, block_range);



CREATE INDEX attr_11_3_token_day_data_daily_volume_token ON sgd1.token_day_data USING btree (daily_volume_token);



CREATE INDEX attr_11_4_token_day_data_daily_volume_eth ON sgd1.token_day_data USING btree (daily_volume_eth);



CREATE INDEX attr_11_5_token_day_data_daily_volume_usd ON sgd1.token_day_data USING btree (daily_volume_usd);



CREATE INDEX attr_11_6_token_day_data_daily_txns ON sgd1.token_day_data USING btree (daily_txns);



CREATE INDEX attr_11_7_token_day_data_total_liquidity_token ON sgd1.token_day_data USING btree (total_liquidity_token);



CREATE INDEX attr_11_8_token_day_data_total_liquidity_eth ON sgd1.token_day_data USING btree (total_liquidity_eth);



CREATE INDEX attr_11_9_token_day_data_total_liquidity_usd ON sgd1.token_day_data USING btree (total_liquidity_usd);



CREATE INDEX "attr_12_0_poi2$_digest" ON sgd1."poi2$" USING btree (digest);



CREATE INDEX "attr_12_1_poi2$_id" ON sgd1."poi2$" USING btree ("left"(id, 256));



CREATE INDEX attr_1_0_token_id ON sgd1.token USING btree (id);



CREATE INDEX attr_1_10_token_derived_eth ON sgd1.token USING btree (derived_eth);



CREATE INDEX attr_1_1_token_symbol ON sgd1.token USING btree ("left"(symbol, 256));



CREATE INDEX attr_1_2_token_name ON sgd1.token USING btree ("left"(name, 256));



CREATE INDEX attr_1_3_token_decimals ON sgd1.token USING btree (decimals);



CREATE INDEX attr_1_4_token_total_supply ON sgd1.token USING btree (total_supply);



CREATE INDEX attr_1_5_token_trade_volume ON sgd1.token USING btree (trade_volume);



CREATE INDEX attr_1_6_token_trade_volume_usd ON sgd1.token USING btree (trade_volume_usd);



CREATE INDEX attr_1_7_token_untracked_volume_usd ON sgd1.token USING btree (untracked_volume_usd);



CREATE INDEX attr_1_8_token_tx_count ON sgd1.token USING btree (tx_count);



CREATE INDEX attr_1_9_token_total_liquidity ON sgd1.token USING btree (total_liquidity);



CREATE INDEX attr_2_0_pair_id ON sgd1.pair USING btree (id);



CREATE INDEX attr_2_10_pair_token_1_price ON sgd1.pair USING btree (token_1_price);



CREATE INDEX attr_2_11_pair_volume_token_0 ON sgd1.pair USING btree (volume_token_0);



CREATE INDEX attr_2_12_pair_volume_token_1 ON sgd1.pair USING btree (volume_token_1);



CREATE INDEX attr_2_13_pair_volume_usd ON sgd1.pair USING btree (volume_usd);



CREATE INDEX attr_2_14_pair_untracked_volume_usd ON sgd1.pair USING btree (untracked_volume_usd);



CREATE INDEX attr_2_15_pair_tx_count ON sgd1.pair USING btree (tx_count);



CREATE INDEX attr_2_16_pair_created_at_timestamp ON sgd1.pair USING btree (created_at_timestamp);



CREATE INDEX attr_2_17_pair_created_at_block_number ON sgd1.pair USING btree (created_at_block_number);



CREATE INDEX attr_2_1_pair_token_0 ON sgd1.pair USING gist (token_0, block_range);



CREATE INDEX attr_2_2_pair_token_1 ON sgd1.pair USING gist (token_1, block_range);



CREATE INDEX attr_2_3_pair_reserve_0 ON sgd1.pair USING btree (reserve_0);



CREATE INDEX attr_2_4_pair_reserve_1 ON sgd1.pair USING btree (reserve_1);



CREATE INDEX attr_2_5_pair_total_supply ON sgd1.pair USING btree (total_supply);



CREATE INDEX attr_2_6_pair_reserve_eth ON sgd1.pair USING btree (reserve_eth);



CREATE INDEX attr_2_7_pair_reserve_usd ON sgd1.pair USING btree (reserve_usd);



CREATE INDEX attr_2_8_pair_tracked_reserve_eth ON sgd1.pair USING btree (tracked_reserve_eth);



CREATE INDEX attr_2_9_pair_token_0_price ON sgd1.pair USING btree (token_0_price);



CREATE INDEX attr_3_0_transaction_id ON sgd1.transaction USING btree (id);



CREATE INDEX attr_3_1_transaction_block_number ON sgd1.transaction USING btree (block_number);



CREATE INDEX attr_3_2_transaction_timestamp ON sgd1.transaction USING btree ("timestamp");



CREATE INDEX attr_3_3_transaction_mints ON sgd1.transaction USING gin (mints);



CREATE INDEX attr_3_4_transaction_burns ON sgd1.transaction USING gin (burns);



CREATE INDEX attr_3_5_transaction_swaps ON sgd1.transaction USING gin (swaps);



CREATE INDEX attr_4_0_mint_id ON sgd1.mint USING btree (id);



CREATE INDEX attr_4_10_mint_amount_usd ON sgd1.mint USING btree (amount_usd);



CREATE INDEX attr_4_11_mint_fee_to ON sgd1.mint USING btree (fee_to);



CREATE INDEX attr_4_12_mint_fee_liquidity ON sgd1.mint USING btree (fee_liquidity);



CREATE INDEX attr_4_1_mint_transaction ON sgd1.mint USING gist (transaction, block_range);



CREATE INDEX attr_4_2_mint_timestamp ON sgd1.mint USING btree ("timestamp");



CREATE INDEX attr_4_3_mint_pair ON sgd1.mint USING gist (pair, block_range);



CREATE INDEX attr_4_4_mint_to ON sgd1.mint USING btree ("to");



CREATE INDEX attr_4_5_mint_liquidity ON sgd1.mint USING btree (liquidity);



CREATE INDEX attr_4_6_mint_sender ON sgd1.mint USING btree (sender);



CREATE INDEX attr_4_7_mint_amount_0 ON sgd1.mint USING btree (amount_0);



CREATE INDEX attr_4_8_mint_amount_1 ON sgd1.mint USING btree (amount_1);



CREATE INDEX attr_4_9_mint_log_index ON sgd1.mint USING btree (log_index);



CREATE INDEX attr_5_0_burn_id ON sgd1.burn USING btree (id);



CREATE INDEX attr_5_10_burn_amount_usd ON sgd1.burn USING btree (amount_usd);



CREATE INDEX attr_5_11_burn_needs_complete ON sgd1.burn USING btree (needs_complete);



CREATE INDEX attr_5_12_burn_fee_to ON sgd1.burn USING btree (fee_to);



CREATE INDEX attr_5_13_burn_fee_liquidity ON sgd1.burn USING btree (fee_liquidity);



CREATE INDEX attr_5_1_burn_transaction ON sgd1.burn USING gist (transaction, block_range);



CREATE INDEX attr_5_2_burn_timestamp ON sgd1.burn USING btree ("timestamp");



CREATE INDEX attr_5_3_burn_pair ON sgd1.burn USING gist (pair, block_range);



CREATE INDEX attr_5_4_burn_liquidity ON sgd1.burn USING btree (liquidity);



CREATE INDEX attr_5_5_burn_sender ON sgd1.burn USING btree (sender);



CREATE INDEX attr_5_6_burn_amount_0 ON sgd1.burn USING btree (amount_0);



CREATE INDEX attr_5_7_burn_amount_1 ON sgd1.burn USING btree (amount_1);



CREATE INDEX attr_5_8_burn_to ON sgd1.burn USING btree ("to");



CREATE INDEX attr_5_9_burn_log_index ON sgd1.burn USING btree (log_index);



CREATE INDEX attr_6_0_swap_id ON sgd1.swap USING btree (id);



CREATE INDEX attr_6_10_swap_to ON sgd1.swap USING btree ("to");



CREATE INDEX attr_6_11_swap_log_index ON sgd1.swap USING btree (log_index);



CREATE INDEX attr_6_12_swap_amount_usd ON sgd1.swap USING btree (amount_usd);



CREATE INDEX attr_6_1_swap_transaction ON sgd1.swap USING gist (transaction, block_range);



CREATE INDEX attr_6_2_swap_timestamp ON sgd1.swap USING btree ("timestamp");



CREATE INDEX attr_6_3_swap_pair ON sgd1.swap USING gist (pair, block_range);



CREATE INDEX attr_6_4_swap_sender ON sgd1.swap USING btree (sender);



CREATE INDEX attr_6_5_swap_from ON sgd1.swap USING btree ("from");



CREATE INDEX attr_6_6_swap_amount_0_in ON sgd1.swap USING btree (amount_0_in);



CREATE INDEX attr_6_7_swap_amount_1_in ON sgd1.swap USING btree (amount_1_in);



CREATE INDEX attr_6_8_swap_amount_0_out ON sgd1.swap USING btree (amount_0_out);



CREATE INDEX attr_6_9_swap_amount_1_out ON sgd1.swap USING btree (amount_1_out);



CREATE INDEX attr_7_0_bundle_id ON sgd1.bundle USING btree (id);



CREATE INDEX attr_7_1_bundle_eth_price ON sgd1.bundle USING btree (eth_price);



CREATE INDEX attr_8_0_uniswap_day_data_id ON sgd1.uniswap_day_data USING btree (id);



CREATE INDEX attr_8_1_uniswap_day_data_date ON sgd1.uniswap_day_data USING btree (date);



CREATE INDEX attr_8_2_uniswap_day_data_daily_volume_eth ON sgd1.uniswap_day_data USING btree (daily_volume_eth);



CREATE INDEX attr_8_3_uniswap_day_data_daily_volume_usd ON sgd1.uniswap_day_data USING btree (daily_volume_usd);



CREATE INDEX attr_8_4_uniswap_day_data_daily_volume_untracked ON sgd1.uniswap_day_data USING btree (daily_volume_untracked);



CREATE INDEX attr_8_5_uniswap_day_data_total_volume_eth ON sgd1.uniswap_day_data USING btree (total_volume_eth);



CREATE INDEX attr_8_6_uniswap_day_data_total_liquidity_eth ON sgd1.uniswap_day_data USING btree (total_liquidity_eth);



CREATE INDEX attr_8_7_uniswap_day_data_total_volume_usd ON sgd1.uniswap_day_data USING btree (total_volume_usd);



CREATE INDEX attr_8_8_uniswap_day_data_total_liquidity_usd ON sgd1.uniswap_day_data USING btree (total_liquidity_usd);



CREATE INDEX attr_8_9_uniswap_day_data_tx_count ON sgd1.uniswap_day_data USING btree (tx_count);



CREATE INDEX attr_9_0_pair_hour_data_id ON sgd1.pair_hour_data USING btree (id);



CREATE INDEX attr_9_1_pair_hour_data_hour_start_unix ON sgd1.pair_hour_data USING btree (hour_start_unix);



CREATE INDEX attr_9_2_pair_hour_data_pair ON sgd1.pair_hour_data USING gist (pair, block_range);



CREATE INDEX attr_9_3_pair_hour_data_reserve_0 ON sgd1.pair_hour_data USING btree (reserve_0);



CREATE INDEX attr_9_4_pair_hour_data_reserve_1 ON sgd1.pair_hour_data USING btree (reserve_1);



CREATE INDEX attr_9_5_pair_hour_data_reserve_usd ON sgd1.pair_hour_data USING btree (reserve_usd);



CREATE INDEX attr_9_6_pair_hour_data_hourly_volume_token_0 ON sgd1.pair_hour_data USING btree (hourly_volume_token_0);



CREATE INDEX attr_9_7_pair_hour_data_hourly_volume_token_1 ON sgd1.pair_hour_data USING btree (hourly_volume_token_1);



CREATE INDEX attr_9_8_pair_hour_data_hourly_volume_usd ON sgd1.pair_hour_data USING btree (hourly_volume_usd);



CREATE INDEX attr_9_9_pair_hour_data_hourly_txns ON sgd1.pair_hour_data USING btree (hourly_txns);



CREATE INDEX brin_bundle ON sgd1.bundle USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_burn ON sgd1.burn USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_mint ON sgd1.mint USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_pair ON sgd1.pair USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_pair_day_data ON sgd1.pair_day_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_pair_hour_data ON sgd1.pair_hour_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX "brin_poi2$" ON sgd1."poi2$" USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_swap ON sgd1.swap USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_token ON sgd1.token USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_token_day_data ON sgd1.token_day_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_transaction ON sgd1.transaction USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_uniswap_day_data ON sgd1.uniswap_day_data USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_uniswap_factory ON sgd1.uniswap_factory USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX bundle_block_range_closed ON sgd1.bundle USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX burn_block_range_closed ON sgd1.burn USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX mint_block_range_closed ON sgd1.mint USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX pair_block_range_closed ON sgd1.pair USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX pair_day_data_block_range_closed ON sgd1.pair_day_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX pair_hour_data_block_range_closed ON sgd1.pair_hour_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX "poi2$_block_range_closed" ON sgd1."poi2$" USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX swap_block_range_closed ON sgd1.swap USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX token_block_range_closed ON sgd1.token USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX token_day_data_block_range_closed ON sgd1.token_day_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX transaction_block_range_closed ON sgd1.transaction USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX uniswap_day_data_block_range_closed ON sgd1.uniswap_day_data USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX uniswap_factory_block_range_closed ON sgd1.uniswap_factory USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);