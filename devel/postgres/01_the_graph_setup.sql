\c "graph-node";

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


CREATE SCHEMA chain1;



CREATE SCHEMA chain2;



CREATE SCHEMA info;



CREATE SCHEMA primary_public;



CREATE SCHEMA subgraphs;



CREATE EXTENSION IF NOT EXISTS btree_gist WITH SCHEMA public;



COMMENT ON EXTENSION btree_gist IS 'support for indexing common datatypes in GiST';



CREATE EXTENSION IF NOT EXISTS pg_stat_statements WITH SCHEMA public;



COMMENT ON EXTENSION pg_stat_statements IS 'track planning and execution statistics of all SQL statements executed';



CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;



COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';



CREATE EXTENSION IF NOT EXISTS postgres_fdw WITH SCHEMA public;



COMMENT ON EXTENSION postgres_fdw IS 'foreign-data wrapper for remote PostgreSQL servers';



CREATE TYPE public.deployment_schema_version AS ENUM (
    'split',
    'relational'
);



CREATE TYPE subgraphs.health AS ENUM (
    'failed',
    'healthy',
    'unhealthy'
);



CREATE FUNCTION public.reduce_dim(anyarray) RETURNS SETOF anyarray
    LANGUAGE plpgsql IMMUTABLE
    AS $_$
DECLARE
    s $1%TYPE;
BEGIN
    FOREACH s SLICE 1  IN ARRAY $1 LOOP
        RETURN NEXT s;
    END LOOP;
    RETURN;
END;
$_$;



CREATE FUNCTION public.subgraph_log_entity_event() RETURNS trigger
    LANGUAGE plpgsql
    AS $_$
DECLARE
    event_id INTEGER;
    new_event_id INTEGER;
    is_reversion BOOLEAN := FALSE;
    operation_type INTEGER := 10;
    event_source  VARCHAR;
    entity VARCHAR;
    entity_id VARCHAR;
    data_before JSONB;
BEGIN
    -- Get operation type and source
    IF (TG_OP = 'INSERT') THEN
        operation_type := 0;
        event_source := NEW.event_source;
        entity := NEW.entity;
        entity_id := NEW.id;
        data_before := NULL;
    ELSIF (TG_OP = 'UPDATE') THEN
        operation_type := 1;
        event_source := NEW.event_source;
        entity := OLD.entity;
        entity_id := OLD.id;
        data_before := OLD.data;
    ELSIF (TG_OP = 'DELETE') THEN
        operation_type := 2;
        event_source := current_setting('vars.current_event_source', TRUE);
        entity := OLD.entity;
        entity_id := OLD.id;
        data_before := OLD.data;
    ELSE
        RAISE EXCEPTION 'unexpected entity row operation type, %', TG_OP;
    END IF;

    IF event_source = 'REVERSION' THEN
        is_reversion := TRUE;
    END IF;

    SELECT id INTO event_id
    FROM event_meta_data
    WHERE db_transaction_id = txid_current();

    new_event_id := null;

    IF event_id IS NULL THEN
        -- Log information on the postgres transaction for later use in
        -- revert operations
        INSERT INTO event_meta_data
            (db_transaction_id, db_transaction_time, source)
        VALUES
            (txid_current(), statement_timestamp(), event_source)
        RETURNING event_meta_data.id INTO new_event_id;
    END IF;

    -- Log row metadata and changes, specify whether event was an original
    -- ethereum event or a reversion
    EXECUTE format('INSERT INTO %I.entity_history
        (event_id, entity_id, entity,
         data_before, reversion, op_id)
      VALUES
        ($1, $2, $3, $4, $5, $6)', TG_TABLE_SCHEMA)
    USING COALESCE(new_event_id, event_id), entity_id, entity,
          data_before, is_reversion, operation_type;
    RETURN NULL;
END;
$_$;


SET default_tablespace = '';

SET default_table_access_method = heap;


CREATE TABLE chain1.blocks (
    hash bytea NOT NULL,
    number bigint NOT NULL,
    parent_hash bytea NOT NULL,
    data jsonb NOT NULL
);



CREATE TABLE chain1.call_cache (
    id bytea NOT NULL,
    return_value bytea NOT NULL,
    contract_address bytea NOT NULL,
    block_number integer NOT NULL
);



CREATE TABLE chain1.call_meta (
    contract_address bytea NOT NULL,
    accessed_at date NOT NULL
);



CREATE TABLE chain2.blocks (
    hash bytea NOT NULL,
    number bigint NOT NULL,
    parent_hash bytea NOT NULL,
    data jsonb NOT NULL
);



CREATE TABLE chain2.call_cache (
    id bytea NOT NULL,
    return_value bytea NOT NULL,
    contract_address bytea NOT NULL,
    block_number integer NOT NULL
);



CREATE TABLE chain2.call_meta (
    contract_address bytea NOT NULL,
    accessed_at date NOT NULL
);



CREATE VIEW info.activity AS
 SELECT COALESCE(NULLIF(pg_stat_activity.application_name, ''::text), 'unknown'::text) AS application_name,
    pg_stat_activity.pid,
    date_part('epoch'::text, age(now(), pg_stat_activity.query_start)) AS query_age,
    date_part('epoch'::text, age(now(), pg_stat_activity.xact_start)) AS txn_age,
    pg_stat_activity.query
   FROM pg_stat_activity
  WHERE (pg_stat_activity.state = 'active'::text)
  ORDER BY pg_stat_activity.query_start DESC;



CREATE TABLE public.deployment_schemas (
    id integer NOT NULL,
    subgraph character varying NOT NULL,
    name character varying NOT NULL,
    version public.deployment_schema_version NOT NULL,
    shard text NOT NULL,
    network text NOT NULL,
    active boolean NOT NULL
);



CREATE MATERIALIZED VIEW info.subgraph_sizes AS
 SELECT a.name,
    a.subgraph,
    a.version,
    a.row_estimate,
    a.total_bytes,
    a.index_bytes,
    a.toast_bytes,
    a.table_bytes,
    pg_size_pretty(a.total_bytes) AS total,
    pg_size_pretty(a.index_bytes) AS index,
    pg_size_pretty(a.toast_bytes) AS toast,
    pg_size_pretty(a.table_bytes) AS "table"
   FROM ( SELECT a_1.name,
            a_1.subgraph,
            a_1.version,
            a_1.row_estimate,
            a_1.total_bytes,
            a_1.index_bytes,
            a_1.toast_bytes,
            ((a_1.total_bytes - a_1.index_bytes) - COALESCE(a_1.toast_bytes, (0)::numeric)) AS table_bytes
           FROM ( SELECT n.nspname AS name,
                    ds.subgraph,
                    (ds.version)::text AS version,
                    sum(c.reltuples) AS row_estimate,
                    sum(pg_total_relation_size((c.oid)::regclass)) AS total_bytes,
                    sum(pg_indexes_size((c.oid)::regclass)) AS index_bytes,
                    sum(pg_total_relation_size((c.reltoastrelid)::regclass)) AS toast_bytes
                   FROM ((pg_class c
                     JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
                     JOIN public.deployment_schemas ds ON (((ds.name)::text = n.nspname)))
                  WHERE ((c.relkind = 'r'::"char") AND (n.nspname ~~ 'sgd%'::text))
                  GROUP BY n.nspname, ds.subgraph, ds.version) a_1) a
  WITH NO DATA;



CREATE MATERIALIZED VIEW info.table_sizes AS
 SELECT a.table_schema,
    a.table_name,
    a.version,
    a.row_estimate,
    a.total_bytes,
    a.index_bytes,
    a.toast_bytes,
    a.table_bytes,
    pg_size_pretty(a.total_bytes) AS total,
    pg_size_pretty(a.index_bytes) AS index,
    pg_size_pretty(a.toast_bytes) AS toast,
    pg_size_pretty(a.table_bytes) AS "table"
   FROM ( SELECT a_1.table_schema,
            a_1.table_name,
            a_1.version,
            a_1.row_estimate,
            a_1.total_bytes,
            a_1.index_bytes,
            a_1.toast_bytes,
            ((a_1.total_bytes - a_1.index_bytes) - COALESCE(a_1.toast_bytes, (0)::bigint)) AS table_bytes
           FROM ( SELECT n.nspname AS table_schema,
                    c.relname AS table_name,
                    'shared'::text AS version,
                    c.reltuples AS row_estimate,
                    pg_total_relation_size((c.oid)::regclass) AS total_bytes,
                    pg_indexes_size((c.oid)::regclass) AS index_bytes,
                    pg_total_relation_size((c.reltoastrelid)::regclass) AS toast_bytes
                   FROM (pg_class c
                     JOIN pg_namespace n ON ((n.oid = c.relnamespace)))
                  WHERE ((c.relkind = 'r'::"char") AND (n.nspname = ANY (ARRAY['public'::name, 'subgraphs'::name])))) a_1) a
  WITH NO DATA;



CREATE VIEW info.all_sizes AS
 SELECT subgraph_sizes.name,
    subgraph_sizes.subgraph,
    subgraph_sizes.version,
    subgraph_sizes.row_estimate,
    subgraph_sizes.total_bytes,
    subgraph_sizes.index_bytes,
    subgraph_sizes.toast_bytes,
    subgraph_sizes.table_bytes,
    subgraph_sizes.total,
    subgraph_sizes.index,
    subgraph_sizes.toast,
    subgraph_sizes."table"
   FROM info.subgraph_sizes
UNION ALL
 SELECT table_sizes.table_schema AS name,
    table_sizes.table_name AS subgraph,
    table_sizes.version,
    table_sizes.row_estimate,
    table_sizes.total_bytes,
    table_sizes.index_bytes,
    table_sizes.toast_bytes,
    table_sizes.table_bytes,
    table_sizes.total,
    table_sizes.index,
    table_sizes.toast,
    table_sizes."table"
   FROM info.table_sizes;



CREATE TABLE subgraphs.subgraph (
    id text NOT NULL,
    name text NOT NULL,
    current_version text,
    pending_version text,
    created_at numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE TABLE subgraphs.subgraph_deployment (
    deployment text NOT NULL,
    failed boolean NOT NULL,
    synced boolean NOT NULL,
    earliest_ethereum_block_hash bytea,
    earliest_ethereum_block_number numeric,
    latest_ethereum_block_hash bytea,
    latest_ethereum_block_number numeric,
    entity_count numeric NOT NULL,
    graft_base text,
    graft_block_hash bytea,
    graft_block_number numeric,
    fatal_error text,
    non_fatal_errors text[] DEFAULT '{}'::text[],
    health subgraphs.health NOT NULL,
    reorg_count integer DEFAULT 0 NOT NULL,
    current_reorg_depth integer DEFAULT 0 NOT NULL,
    max_reorg_depth integer DEFAULT 0 NOT NULL,
    last_healthy_ethereum_block_hash bytea,
    last_healthy_ethereum_block_number numeric,
    id integer NOT NULL
);



CREATE TABLE subgraphs.subgraph_version (
    id text NOT NULL,
    subgraph text NOT NULL,
    deployment text NOT NULL,
    created_at numeric NOT NULL,
    vid bigint NOT NULL,
    block_range int4range NOT NULL
);



CREATE VIEW info.subgraph_info AS
 SELECT ds.id AS schema_id,
    ds.name AS schema_name,
    ds.subgraph,
    ds.version,
    s.name,
        CASE
            WHEN (s.pending_version = v.id) THEN 'pending'::text
            WHEN (s.current_version = v.id) THEN 'current'::text
            ELSE 'unused'::text
        END AS status,
    d.failed,
    d.synced
   FROM public.deployment_schemas ds,
    subgraphs.subgraph_deployment d,
    subgraphs.subgraph_version v,
    subgraphs.subgraph s
  WHERE ((d.deployment = (ds.subgraph)::text) AND (v.deployment = d.deployment) AND (v.subgraph = s.id));



CREATE VIEW info.wraparound AS
 SELECT ((pg_class.oid)::regclass)::text AS "table",
    LEAST((( SELECT (pg_settings.setting)::integer AS setting
           FROM pg_settings
          WHERE (pg_settings.name = 'autovacuum_freeze_max_age'::text)) - age(pg_class.relfrozenxid)), (( SELECT (pg_settings.setting)::integer AS setting
           FROM pg_settings
          WHERE (pg_settings.name = 'autovacuum_multixact_freeze_max_age'::text)) - mxid_age(pg_class.relminmxid))) AS tx_before_wraparound_vacuum,
    pg_size_pretty(pg_total_relation_size((pg_class.oid)::regclass)) AS size,
    pg_stat_get_last_autovacuum_time(pg_class.oid) AS last_autovacuum,
    age(pg_class.relfrozenxid) AS xid_age,
    mxid_age(pg_class.relminmxid) AS mxid_age
   FROM pg_class
  WHERE ((pg_class.relfrozenxid <> 0) AND (pg_class.oid > (16384)::oid) AND (pg_class.relkind = 'r'::"char"))
  ORDER BY LEAST((( SELECT (pg_settings.setting)::integer AS setting
           FROM pg_settings
          WHERE (pg_settings.name = 'autovacuum_freeze_max_age'::text)) - age(pg_class.relfrozenxid)), (( SELECT (pg_settings.setting)::integer AS setting
           FROM pg_settings
          WHERE (pg_settings.name = 'autovacuum_multixact_freeze_max_age'::text)) - mxid_age(pg_class.relminmxid)));



CREATE TABLE public.active_copies (
    src integer NOT NULL,
    dst integer NOT NULL,
    queued_at timestamp with time zone NOT NULL,
    cancelled_at timestamp with time zone
);



CREATE VIEW primary_public.active_copies AS
 SELECT active_copies.src,
    active_copies.dst,
    active_copies.queued_at,
    active_copies.cancelled_at
   FROM public.active_copies;



CREATE TABLE public.chains (
    id integer NOT NULL,
    name text NOT NULL,
    net_version text NOT NULL,
    genesis_block_hash text NOT NULL,
    shard text NOT NULL,
    namespace text NOT NULL,
    CONSTRAINT chains_genesis_version_check CHECK (((net_version IS NULL) = (genesis_block_hash IS NULL)))
);



CREATE VIEW primary_public.chains AS
 SELECT chains.id,
    chains.name,
    chains.net_version,
    chains.genesis_block_hash,
    chains.shard,
    chains.namespace
   FROM public.chains;



CREATE VIEW primary_public.deployment_schemas AS
 SELECT deployment_schemas.id,
    deployment_schemas.subgraph,
    deployment_schemas.name,
    deployment_schemas.version,
    deployment_schemas.shard,
    deployment_schemas.network,
    deployment_schemas.active
   FROM public.deployment_schemas;



CREATE TABLE public.__diesel_schema_migrations (
    version character varying(50) NOT NULL,
    run_on timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);



CREATE SEQUENCE public.chains_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE public.chains_id_seq OWNED BY public.chains.id;



CREATE TABLE public.db_version (
    db_version bigint NOT NULL
);



CREATE SEQUENCE public.deployment_schemas_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE public.deployment_schemas_id_seq OWNED BY public.deployment_schemas.id;



CREATE TABLE public.ens_names (
    hash character varying NOT NULL,
    name character varying NOT NULL
);



CREATE TABLE public.eth_call_cache (
    id bytea NOT NULL,
    return_value bytea NOT NULL,
    contract_address bytea NOT NULL,
    block_number integer NOT NULL
);



CREATE TABLE public.eth_call_meta (
    contract_address bytea NOT NULL,
    accessed_at date NOT NULL
);



CREATE TABLE public.ethereum_blocks (
    hash character varying NOT NULL,
    number bigint NOT NULL,
    parent_hash character varying NOT NULL,
    network_name character varying NOT NULL,
    data jsonb NOT NULL
);



CREATE TABLE public.ethereum_networks (
    name character varying NOT NULL,
    head_block_hash character varying,
    head_block_number bigint,
    net_version character varying NOT NULL,
    genesis_block_hash character varying NOT NULL,
    namespace text NOT NULL,
    CONSTRAINT ethereum_networks_check CHECK (((head_block_hash IS NULL) = (head_block_number IS NULL))),
    CONSTRAINT ethereum_networks_check1 CHECK (((net_version IS NULL) = (genesis_block_hash IS NULL)))
);



CREATE TABLE public.event_meta_data (
    id integer NOT NULL,
    db_transaction_id bigint NOT NULL,
    db_transaction_time timestamp without time zone NOT NULL,
    source character varying
);



CREATE SEQUENCE public.event_meta_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE public.event_meta_data_id_seq OWNED BY public.event_meta_data.id;



CREATE UNLOGGED TABLE public.large_notifications (
    id integer NOT NULL,
    payload character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);



COMMENT ON TABLE public.large_notifications IS 'Table for notifications whose payload is too big to send directly';



CREATE SEQUENCE public.large_notifications_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE public.large_notifications_id_seq OWNED BY public.large_notifications.id;



CREATE TABLE public.unneeded_event_ids (
    event_id bigint NOT NULL
);



CREATE TABLE public.unused_deployments (
    deployment text NOT NULL,
    unused_at timestamp with time zone DEFAULT now() NOT NULL,
    removed_at timestamp with time zone,
    subgraphs text[],
    namespace text NOT NULL,
    shard text NOT NULL,
    entity_count integer DEFAULT 0 NOT NULL,
    latest_ethereum_block_hash bytea,
    latest_ethereum_block_number integer,
    failed boolean DEFAULT false NOT NULL,
    synced boolean DEFAULT false NOT NULL,
    id integer NOT NULL
);



CREATE TABLE subgraphs.copy_state (
    src integer NOT NULL,
    dst integer NOT NULL,
    target_block_hash bytea NOT NULL,
    target_block_number integer NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    cancelled_at timestamp with time zone
);



CREATE TABLE subgraphs.copy_table_state (
    id integer NOT NULL,
    entity_type text NOT NULL,
    dst integer NOT NULL,
    next_vid bigint NOT NULL,
    target_vid bigint NOT NULL,
    batch_size bigint NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    finished_at timestamp with time zone,
    duration_ms bigint DEFAULT 0 NOT NULL
);



CREATE SEQUENCE subgraphs.copy_table_state_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE subgraphs.copy_table_state_id_seq OWNED BY subgraphs.copy_table_state.id;



CREATE TABLE subgraphs.dynamic_ethereum_contract_data_source (
    name text NOT NULL,
    ethereum_block_hash bytea NOT NULL,
    ethereum_block_number numeric NOT NULL,
    deployment text NOT NULL,
    vid bigint NOT NULL,
    context text,
    address bytea NOT NULL,
    abi text NOT NULL,
    start_block integer NOT NULL
);



CREATE SEQUENCE subgraphs.dynamic_ethereum_contract_data_source_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE subgraphs.dynamic_ethereum_contract_data_source_vid_seq OWNED BY subgraphs.dynamic_ethereum_contract_data_source.vid;



CREATE TABLE subgraphs.subgraph_deployment_assignment (
    node_id text NOT NULL,
    id integer NOT NULL
);



CREATE TABLE subgraphs.subgraph_error (
    id text NOT NULL,
    subgraph_id text NOT NULL,
    message text NOT NULL,
    block_hash bytea,
    handler text,
    vid bigint NOT NULL,
    block_range int4range NOT NULL,
    deterministic boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);



CREATE SEQUENCE subgraphs.subgraph_error_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE subgraphs.subgraph_error_vid_seq OWNED BY subgraphs.subgraph_error.vid;



CREATE TABLE subgraphs.subgraph_manifest (
    spec_version text NOT NULL,
    description text,
    repository text,
    schema text NOT NULL,
    features text[] DEFAULT '{}'::text[] NOT NULL,
    id integer NOT NULL
);



CREATE SEQUENCE subgraphs.subgraph_version_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE subgraphs.subgraph_version_vid_seq OWNED BY subgraphs.subgraph_version.vid;



CREATE SEQUENCE subgraphs.subgraph_vid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE subgraphs.subgraph_vid_seq OWNED BY subgraphs.subgraph.vid;



ALTER TABLE ONLY public.chains ALTER COLUMN id SET DEFAULT nextval('public.chains_id_seq'::regclass);



ALTER TABLE ONLY public.chains ALTER COLUMN namespace SET DEFAULT ('chain'::text || currval('public.chains_id_seq'::regclass));



ALTER TABLE ONLY public.deployment_schemas ALTER COLUMN id SET DEFAULT nextval('public.deployment_schemas_id_seq'::regclass);



ALTER TABLE ONLY public.deployment_schemas ALTER COLUMN name SET DEFAULT ('sgd'::text || currval('public.deployment_schemas_id_seq'::regclass));



ALTER TABLE ONLY public.event_meta_data ALTER COLUMN id SET DEFAULT nextval('public.event_meta_data_id_seq'::regclass);



ALTER TABLE ONLY public.large_notifications ALTER COLUMN id SET DEFAULT nextval('public.large_notifications_id_seq'::regclass);



ALTER TABLE ONLY subgraphs.copy_table_state ALTER COLUMN id SET DEFAULT nextval('subgraphs.copy_table_state_id_seq'::regclass);



ALTER TABLE ONLY subgraphs.dynamic_ethereum_contract_data_source ALTER COLUMN vid SET DEFAULT nextval('subgraphs.dynamic_ethereum_contract_data_source_vid_seq'::regclass);



ALTER TABLE ONLY subgraphs.subgraph ALTER COLUMN vid SET DEFAULT nextval('subgraphs.subgraph_vid_seq'::regclass);



ALTER TABLE ONLY subgraphs.subgraph_error ALTER COLUMN vid SET DEFAULT nextval('subgraphs.subgraph_error_vid_seq'::regclass);



ALTER TABLE ONLY subgraphs.subgraph_version ALTER COLUMN vid SET DEFAULT nextval('subgraphs.subgraph_version_vid_seq'::regclass);



ALTER TABLE ONLY chain1.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (hash);



ALTER TABLE ONLY chain1.call_cache
    ADD CONSTRAINT call_cache_pkey PRIMARY KEY (id);



ALTER TABLE ONLY chain1.call_meta
    ADD CONSTRAINT call_meta_pkey PRIMARY KEY (contract_address);



ALTER TABLE ONLY chain2.blocks
    ADD CONSTRAINT blocks_pkey PRIMARY KEY (hash);



ALTER TABLE ONLY chain2.call_cache
    ADD CONSTRAINT call_cache_pkey PRIMARY KEY (id);



ALTER TABLE ONLY chain2.call_meta
    ADD CONSTRAINT call_meta_pkey PRIMARY KEY (contract_address);



ALTER TABLE ONLY public.__diesel_schema_migrations
    ADD CONSTRAINT __diesel_schema_migrations_pkey PRIMARY KEY (version);



ALTER TABLE ONLY public.active_copies
    ADD CONSTRAINT active_copies_pkey PRIMARY KEY (dst);



ALTER TABLE ONLY public.active_copies
    ADD CONSTRAINT active_copies_src_dst_key UNIQUE (src, dst);



ALTER TABLE ONLY public.chains
    ADD CONSTRAINT chains_name_key UNIQUE (name);



ALTER TABLE ONLY public.chains
    ADD CONSTRAINT chains_pkey PRIMARY KEY (id);



ALTER TABLE ONLY public.db_version
    ADD CONSTRAINT db_version_pkey PRIMARY KEY (db_version);



ALTER TABLE ONLY public.deployment_schemas
    ADD CONSTRAINT deployment_schemas_pkey PRIMARY KEY (id);



ALTER TABLE ONLY public.ens_names
    ADD CONSTRAINT ens_names_pkey PRIMARY KEY (hash);



ALTER TABLE ONLY public.eth_call_cache
    ADD CONSTRAINT eth_call_cache_pkey PRIMARY KEY (id);



ALTER TABLE ONLY public.eth_call_meta
    ADD CONSTRAINT eth_call_meta_pkey PRIMARY KEY (contract_address);



ALTER TABLE ONLY public.ethereum_blocks
    ADD CONSTRAINT ethereum_blocks_pkey PRIMARY KEY (hash);



ALTER TABLE ONLY public.ethereum_networks
    ADD CONSTRAINT ethereum_networks_pkey PRIMARY KEY (name);



ALTER TABLE ONLY public.event_meta_data
    ADD CONSTRAINT event_meta_data_db_transaction_id_key UNIQUE (db_transaction_id);



ALTER TABLE ONLY public.event_meta_data
    ADD CONSTRAINT event_meta_data_pkey PRIMARY KEY (id);



ALTER TABLE ONLY public.large_notifications
    ADD CONSTRAINT large_notifications_pkey PRIMARY KEY (id);



ALTER TABLE ONLY public.unneeded_event_ids
    ADD CONSTRAINT unneeded_event_ids_pkey PRIMARY KEY (event_id);



ALTER TABLE ONLY public.unused_deployments
    ADD CONSTRAINT unused_deployments_pkey PRIMARY KEY (id);



ALTER TABLE ONLY subgraphs.copy_state
    ADD CONSTRAINT copy_state_pkey PRIMARY KEY (dst);



ALTER TABLE ONLY subgraphs.copy_table_state
    ADD CONSTRAINT copy_table_state_dst_entity_type_key UNIQUE (dst, entity_type);



ALTER TABLE ONLY subgraphs.copy_table_state
    ADD CONSTRAINT copy_table_state_pkey PRIMARY KEY (id);



ALTER TABLE ONLY subgraphs.dynamic_ethereum_contract_data_source
    ADD CONSTRAINT dynamic_ethereum_contract_data_source_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY subgraphs.subgraph_deployment_assignment
    ADD CONSTRAINT subgraph_deployment_assignment_pkey PRIMARY KEY (id);



ALTER TABLE ONLY subgraphs.subgraph_deployment
    ADD CONSTRAINT subgraph_deployment_id_key UNIQUE (deployment);



ALTER TABLE ONLY subgraphs.subgraph_deployment
    ADD CONSTRAINT subgraph_deployment_pkey PRIMARY KEY (id);



ALTER TABLE ONLY subgraphs.subgraph_error
    ADD CONSTRAINT subgraph_error_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY subgraphs.subgraph_error
    ADD CONSTRAINT subgraph_error_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY subgraphs.subgraph
    ADD CONSTRAINT subgraph_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY subgraphs.subgraph_manifest
    ADD CONSTRAINT subgraph_manifest_pkey PRIMARY KEY (id);



ALTER TABLE ONLY subgraphs.subgraph
    ADD CONSTRAINT subgraph_name_uq UNIQUE (name);



ALTER TABLE ONLY subgraphs.subgraph
    ADD CONSTRAINT subgraph_pkey PRIMARY KEY (vid);



ALTER TABLE ONLY subgraphs.subgraph_version
    ADD CONSTRAINT subgraph_version_id_block_range_excl EXCLUDE USING gist (id WITH =, block_range WITH &&);



ALTER TABLE ONLY subgraphs.subgraph_version
    ADD CONSTRAINT subgraph_version_pkey PRIMARY KEY (vid);



CREATE INDEX blocks_number ON chain1.blocks USING btree (number);



CREATE INDEX blocks_number ON chain2.blocks USING btree (number);



CREATE UNIQUE INDEX deployment_schemas_deployment_active ON public.deployment_schemas USING btree (subgraph) WHERE active;



CREATE UNIQUE INDEX deployment_schemas_subgraph_shard_uq ON public.deployment_schemas USING btree (subgraph, shard);



CREATE INDEX ethereum_blocks_name_number ON public.ethereum_blocks USING btree (network_name, number);



CREATE INDEX event_meta_data_source ON public.event_meta_data USING btree (source);



CREATE INDEX attr_0_0_subgraph_id ON subgraphs.subgraph USING btree (id);



CREATE INDEX attr_0_1_subgraph_name ON subgraphs.subgraph USING btree ("left"(name, 256));



CREATE INDEX attr_0_2_subgraph_current_version ON subgraphs.subgraph USING btree (current_version);



CREATE INDEX attr_0_3_subgraph_pending_version ON subgraphs.subgraph USING btree (pending_version);



CREATE INDEX attr_0_4_subgraph_created_at ON subgraphs.subgraph USING btree (created_at);



CREATE INDEX attr_16_0_subgraph_error_id ON subgraphs.subgraph_error USING btree (id);



CREATE INDEX attr_16_1_subgraph_error_subgraph_id ON subgraphs.subgraph_error USING btree ("left"(subgraph_id, 256));



CREATE INDEX attr_1_0_subgraph_version_id ON subgraphs.subgraph_version USING btree (id);



CREATE INDEX attr_1_1_subgraph_version_subgraph ON subgraphs.subgraph_version USING btree (subgraph);



CREATE INDEX attr_1_2_subgraph_version_deployment ON subgraphs.subgraph_version USING btree (deployment);



CREATE INDEX attr_1_3_subgraph_version_created_at ON subgraphs.subgraph_version USING btree (created_at);



CREATE INDEX attr_2_0_subgraph_deployment_id ON subgraphs.subgraph_deployment USING btree (deployment);



CREATE INDEX attr_2_11_subgraph_deployment_entity_count ON subgraphs.subgraph_deployment USING btree (entity_count);



CREATE INDEX attr_2_2_subgraph_deployment_failed ON subgraphs.subgraph_deployment USING btree (failed);



CREATE INDEX attr_2_3_subgraph_deployment_synced ON subgraphs.subgraph_deployment USING btree (synced);



CREATE INDEX attr_2_4_subgraph_deployment_earliest_ethereum_block_hash ON subgraphs.subgraph_deployment USING btree (earliest_ethereum_block_hash);



CREATE INDEX attr_2_5_subgraph_deployment_earliest_ethereum_block_number ON subgraphs.subgraph_deployment USING btree (earliest_ethereum_block_number);



CREATE INDEX attr_2_6_subgraph_deployment_latest_ethereum_block_hash ON subgraphs.subgraph_deployment USING btree (latest_ethereum_block_hash);



CREATE INDEX attr_2_7_subgraph_deployment_latest_ethereum_block_number ON subgraphs.subgraph_deployment USING btree (latest_ethereum_block_number);



CREATE INDEX attr_3_1_subgraph_deployment_assignment_node_id ON subgraphs.subgraph_deployment_assignment USING btree ("left"(node_id, 256));



CREATE INDEX attr_6_9_dynamic_ethereum_contract_data_source_deployment ON subgraphs.dynamic_ethereum_contract_data_source USING btree (deployment);



CREATE INDEX attr_subgraph_deployment_health ON subgraphs.subgraph_deployment USING btree (health);



ALTER TABLE ONLY public.active_copies
    ADD CONSTRAINT active_copies_dst_fkey FOREIGN KEY (dst) REFERENCES public.deployment_schemas(id) ON DELETE CASCADE;



ALTER TABLE ONLY public.active_copies
    ADD CONSTRAINT active_copies_src_fkey FOREIGN KEY (src) REFERENCES public.deployment_schemas(id);



ALTER TABLE ONLY public.deployment_schemas
    ADD CONSTRAINT deployment_schemas_network_fkey FOREIGN KEY (network) REFERENCES public.chains(name);



ALTER TABLE ONLY public.ethereum_blocks
    ADD CONSTRAINT ethereum_blocks_network_name_fkey FOREIGN KEY (network_name) REFERENCES public.ethereum_networks(name);



ALTER TABLE ONLY subgraphs.copy_state
    ADD CONSTRAINT copy_state_dst_fkey FOREIGN KEY (dst) REFERENCES subgraphs.subgraph_deployment(id) ON DELETE CASCADE;



ALTER TABLE ONLY subgraphs.copy_table_state
    ADD CONSTRAINT copy_table_state_dst_fkey FOREIGN KEY (dst) REFERENCES subgraphs.copy_state(dst) ON DELETE CASCADE;



ALTER TABLE ONLY subgraphs.subgraph_error
    ADD CONSTRAINT subgraph_error_subgraph_id_fkey FOREIGN KEY (subgraph_id) REFERENCES subgraphs.subgraph_deployment(deployment) ON DELETE CASCADE;



ALTER TABLE ONLY subgraphs.subgraph_manifest
    ADD CONSTRAINT subgraph_manifest_new_id_fkey FOREIGN KEY (id) REFERENCES subgraphs.subgraph_deployment(id) ON DELETE CASCADE;




