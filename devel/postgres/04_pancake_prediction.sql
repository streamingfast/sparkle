\c "graph-node";

CREATE SCHEMA sgd3;

CREATE TYPE sgd3."position" AS ENUM (
    'Bear',
    'Bull',
    'House'
);

CREATE TABLE sgd3.bet (
    id text NOT NULL,
    round text NOT NULL,
    "user" text NOT NULL,
    hash bytea NOT NULL,
    amount numeric NOT NULL,
    "position" sgd3."position" NOT NULL,
    claimed boolean NOT NULL,
    claimed_hash bytea,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE SEQUENCE sgd3.bet_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd3.bet_vid_seq OWNED BY sgd3.bet.vid;



CREATE TABLE sgd3.market (
    id text NOT NULL,
    epoch text,
    paused boolean NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE SEQUENCE sgd3.market_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd3.market_vid_seq OWNED BY sgd3.market.vid;



CREATE TABLE sgd3."poi2$" (
    digest bytea NOT NULL,
    id text NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE SEQUENCE sgd3."poi2$_vid_seq"
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd3."poi2$_vid_seq" OWNED BY sgd3."poi2$".vid;



CREATE TABLE sgd3.round (
    id text NOT NULL,
    epoch numeric NOT NULL,
    "position" sgd3."position",
    failed boolean,
    previous text,
    start_at numeric NOT NULL,
    start_block numeric NOT NULL,
    start_hash bytea NOT NULL,
    lock_at numeric,
    lock_block numeric,
    lock_hash bytea,
    lock_price numeric,
    end_at numeric,
    end_block numeric,
    end_hash bytea,
    close_price numeric,
    total_bets numeric NOT NULL,
    total_amount numeric NOT NULL,
    bull_bets numeric NOT NULL,
    bull_amount numeric NOT NULL,
    bear_bets numeric NOT NULL,
    bear_amount numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE SEQUENCE sgd3.round_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd3.round_vid_seq OWNED BY sgd3.round.vid;



CREATE TABLE sgd3."user" (
    id text NOT NULL,
    address bytea NOT NULL,
    created_at numeric NOT NULL,
    updated_at numeric NOT NULL,
    block numeric NOT NULL,
    total_bets numeric NOT NULL,
    total_bnb numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE SEQUENCE sgd3.user_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE sgd3.user_vid_seq OWNED BY sgd3."user".vid;



ALTER TABLE ONLY sgd3.bet ALTER COLUMN vid SET DEFAULT nextval('sgd3.bet_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.market ALTER COLUMN vid SET DEFAULT nextval('sgd3.market_vid_seq'::regclass);



ALTER TABLE ONLY sgd3."poi2$" ALTER COLUMN vid SET DEFAULT nextval('sgd3."poi2$_vid_seq"'::regclass);



ALTER TABLE ONLY sgd3.round ALTER COLUMN vid SET DEFAULT nextval('sgd3.round_vid_seq'::regclass);



ALTER TABLE ONLY sgd3."user" ALTER COLUMN vid SET DEFAULT nextval('sgd3.user_vid_seq'::regclass);



ALTER TABLE ONLY sgd3.bet
    ADD CONSTRAINT bet_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.bet
    ADD CONSTRAINT bet_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.market
    ADD CONSTRAINT market_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.market
    ADD CONSTRAINT market_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3."poi2$"
    ADD CONSTRAINT "poi2$_id_block_range_excl" EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3."poi2$"
    ADD CONSTRAINT "poi2$_pkey" PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3.round
    ADD CONSTRAINT round_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3.round
    ADD CONSTRAINT round_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY sgd3."user"
    ADD CONSTRAINT user_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY sgd3."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (vid);



CREATE INDEX attr_0_0_market_id ON sgd3.market USING btree (id);



CREATE INDEX attr_0_1_market_epoch ON sgd3.market USING gist (epoch, block_range);



CREATE INDEX attr_0_2_market_paused ON sgd3.market USING btree (paused);



CREATE INDEX attr_1_0_round_id ON sgd3.round USING btree (id);



CREATE INDEX attr_1_10_round_lock_hash ON sgd3.round USING btree (lock_hash);



CREATE INDEX attr_1_11_round_lock_price ON sgd3.round USING btree (lock_price);



CREATE INDEX attr_1_12_round_end_at ON sgd3.round USING btree (end_at);



CREATE INDEX attr_1_13_round_end_block ON sgd3.round USING btree (end_block);



CREATE INDEX attr_1_14_round_end_hash ON sgd3.round USING btree (end_hash);



CREATE INDEX attr_1_15_round_close_price ON sgd3.round USING btree (close_price);



CREATE INDEX attr_1_16_round_total_bets ON sgd3.round USING btree (total_bets);



CREATE INDEX attr_1_17_round_total_amount ON sgd3.round USING btree (total_amount);



CREATE INDEX attr_1_18_round_bull_bets ON sgd3.round USING btree (bull_bets);



CREATE INDEX attr_1_19_round_bull_amount ON sgd3.round USING btree (bull_amount);



CREATE INDEX attr_1_1_round_epoch ON sgd3.round USING btree (epoch);



CREATE INDEX attr_1_20_round_bear_bets ON sgd3.round USING btree (bear_bets);



CREATE INDEX attr_1_21_round_bear_amount ON sgd3.round USING btree (bear_amount);



CREATE INDEX attr_1_2_round_position ON sgd3.round USING btree ("position");



CREATE INDEX attr_1_3_round_failed ON sgd3.round USING btree (failed);



CREATE INDEX attr_1_4_round_previous ON sgd3.round USING gist (previous, block_range);



CREATE INDEX attr_1_5_round_start_at ON sgd3.round USING btree (start_at);



CREATE INDEX attr_1_6_round_start_block ON sgd3.round USING btree (start_block);



CREATE INDEX attr_1_7_round_start_hash ON sgd3.round USING btree (start_hash);



CREATE INDEX attr_1_8_round_lock_at ON sgd3.round USING btree (lock_at);



CREATE INDEX attr_1_9_round_lock_block ON sgd3.round USING btree (lock_block);



CREATE INDEX attr_2_0_user_id ON sgd3."user" USING btree (id);



CREATE INDEX attr_2_1_user_address ON sgd3."user" USING btree (address);



CREATE INDEX attr_2_2_user_created_at ON sgd3."user" USING btree (created_at);



CREATE INDEX attr_2_3_user_updated_at ON sgd3."user" USING btree (updated_at);



CREATE INDEX attr_2_4_user_block ON sgd3."user" USING btree (block);



CREATE INDEX attr_2_5_user_total_bets ON sgd3."user" USING btree (total_bets);



CREATE INDEX attr_2_6_user_total_bnb ON sgd3."user" USING btree (total_bnb);



CREATE INDEX attr_3_0_bet_id ON sgd3.bet USING btree (id);



CREATE INDEX attr_3_1_bet_round ON sgd3.bet USING gist (round, block_range);



CREATE INDEX attr_3_2_bet_user ON sgd3.bet USING gist ("user", block_range);



CREATE INDEX attr_3_3_bet_hash ON sgd3.bet USING btree (hash);



CREATE INDEX attr_3_4_bet_amount ON sgd3.bet USING btree (amount);



CREATE INDEX attr_3_5_bet_position ON sgd3.bet USING btree ("position");



CREATE INDEX attr_3_6_bet_claimed ON sgd3.bet USING btree (claimed);



CREATE INDEX attr_3_7_bet_claimed_hash ON sgd3.bet USING btree (claimed_hash);



CREATE INDEX "attr_4_0_poi2$_digest" ON sgd3."poi2$" USING btree (digest);



CREATE INDEX "attr_4_1_poi2$_id" ON sgd3."poi2$" USING btree ("left"(id, 256));



CREATE INDEX bet_block_range_closed ON sgd3.bet USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX brin_bet ON sgd3.bet USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_market ON sgd3.market USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX "brin_poi2$" ON sgd3."poi2$" USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_round ON sgd3.round USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX brin_user ON sgd3."user" USING brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);



CREATE INDEX market_block_range_closed ON sgd3.market USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX "poi2$_block_range_closed" ON sgd3."poi2$" USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX round_block_range_closed ON sgd3.round USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);



CREATE INDEX user_block_range_closed ON sgd3."user" USING btree (COALESCE(upper(block_range), 2147483647)) WHERE (COALESCE(upper(block_range), 2147483647) < 2147483647);




