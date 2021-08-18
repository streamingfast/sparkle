package codegen

var templateDDL = `
{{$schema := .Schema}}

type DDL struct {
	createTables map[string]string
	indexes      map[string][]*index
	schemaSetup     string

}

var ddl *DDL

type index struct {
	createStatement string
	dropStatement   string
}

var createTables = map[string]string{}
var indexes = map[string][]*index{}

func init() {
	ddl = &DDL{
		createTables: map[string]string{},
		indexes:      map[string][]*index{},
	}

	Definition.DDL = ddl

{{ range .Entities }}
{{$tableName := $engine.ToLowerSnakeCase .Name }}

	ddl.createTables["{{ $tableName }}"] = {{ $engine.BackTick }}
create table if not exists {{ $schema }}.{{ $tableName }}
(
	id text not null,
{{ range $field := .Fields.VisibleOnly }}
	"{{$engine.ToLowerSnakeCase $field.Name }}" {{ $field.PostgresType }},
{{ end }}
	vid bigserial not null constraint {{ $tableName }}_pkey primary key,
	block_range int4range not null,
	_updated_block_number numeric not null
);

alter table {{ $schema }}.{{ $tableName }} owner to graph;
alter sequence {{ $schema }}.{{ $tableName }}_vid_seq owned by {{ $schema }}.{{ $tableName }}.vid;
alter table only {{ $schema }}.{{ $tableName }} alter column vid SET DEFAULT nextval('{{ $schema }}.{{ $tableName }}_vid_seq'::regclass);
{{ $engine.BackTick }}
{{- end }}

{{ range .Entities }}
{{$tableName :=$engine.ToLowerSnakeCase .Name }}
ddl.indexes["{{ $tableName }}"] = func () []*index {
		var indexes []*index
		indexes = append(indexes, &index{
			createStatement: {{ $engine.BackTick }}create index if not exists {{ $tableName }}_block_range_closed on {{ $schema }}.{{ $tableName }} (COALESCE(upper(block_range), 2147483647)) where (COALESCE(upper(block_range), 2147483647) < 2147483647);{{ $engine.BackTick }},
			dropStatement:   {{ $engine.BackTick }}drop index if exists {{ $schema }}.{{ $tableName }}_block_range_closed;{{ $engine.BackTick }},
		})
		indexes = append(indexes, &index{
			createStatement: {{ $engine.BackTick }}create index if not exists {{ $tableName }}_id on {{ $schema }}.{{ $tableName }} (id);{{ $engine.BackTick }},
			dropStatement:   {{ $engine.BackTick }}drop index if exists {{ $schema }}.{{ $tableName }}_id;{{ $engine.BackTick }},
		})
		indexes = append(indexes, &index{
			createStatement: {{ $engine.BackTick }}create index if not exists {{ $tableName }}_updated_block_number on {{ $schema }}.{{ $tableName }} (_updated_block_number);{{ $engine.BackTick }},
			dropStatement:   {{ $engine.BackTick }}drop index if exists {{ $schema }}.{{ $tableName }}_updated_block_number;{{ $engine.BackTick }},
		})

		indexes = append(indexes, &index{
			createStatement: {{ $engine.BackTick }}create index if not exists {{ $tableName }}_id_block_range_fake_excl on {{ $schema }}.{{ $tableName }} using gist (block_range, id);{{ $engine.BackTick }},
			dropStatement:   {{ $engine.BackTick }}drop index if exists {{ $schema }}.{{ $tableName }}_id_block_range_fake_excl;{{ $engine.BackTick }},
		})

		{{- range $field := .Fields.VisibleOnly }}
		{{ if ne $field.PostgresIndex "" }}
			{{$fieldName :=$engine.ToLowerSnakeCase $field.Name}}

		indexes = append(indexes, &index{
			createStatement: {{ $engine.BackTick }}create index if not exists {{ $tableName }}_{{ $fieldName }} on {{ $schema }}.{{ $tableName }} {{ $field.PostgresIndex }};{{ $engine.BackTick }},
			dropStatement:   {{ $engine.BackTick }}drop index if exists {{ $schema }}.{{ $tableName }}_{{ $fieldName }};{{ $engine.BackTick }},
		})
		{{- end }}

		{{- end }}


		return indexes
}()

{{- end }}
ddl.schemaSetup = {{ $engine.BackTick }}
CREATE SCHEMA if not exists {{ $schema }};
DO
$do$
    BEGIN
        IF NOT EXISTS (
                SELECT FROM pg_catalog.pg_roles  -- SELECT list can be empty for this
                WHERE  rolname = 'graph') THEN
            CREATE ROLE graph;
        END IF;
    END
$do$;

set statement_timeout = 0;
set idle_in_transaction_session_timeout = 0;
set client_encoding = 'UTF8';
set standard_conforming_strings = on;
select pg_catalog.set_config('search_path', '', false);
set check_function_bodies = false;
set xmloption = content;
set client_min_messages = warning;
set row_security = off;

create extension if not exists btree_gist with schema {{ $schema }};
{{ range .Enums }}
{{$enumName :=$engine.ToLowerSnakeCase .Name }}
create type {{ $schema }}.{{ $enumName }} as enum ({{ $engine.StringJoin .Fields }});
alter type {{ $schema }}.{{ $enumName }} owner to graph;
{{ end }}

create table if not exists {{ $schema }}.cursor
(
	id integer not null
		constraint cursor_pkey
			primary key,
	cursor text
);
alter table {{ $schema }}.cursor owner to graph;

create table {{ $schema }}.poi2$
(
    digest      bytea     not null,
    id          text      not null,
    vid         bigserial not null
        constraint poi2$_pkey
            primary key,
    block_range int4range not null,
	_updated_block_number  numeric not null,
    constraint poi2$_id_block_range_excl
        exclude using gist (id with =, block_range with &&)
);

alter table {{ $schema }}.poi2$
    owner to graph;

create index brin_poi2$
    on {{ $schema }}.poi2$ using brin (lower(block_range), COALESCE(upper(block_range), 2147483647), vid);

CREATE INDEX poi2$_updated_block_number
    ON {{ $schema }}.poi2$ USING btree
	(_updated_block_number ASC NULLS LAST)
	TABLESPACE pg_default;

create index poi2$_block_range_closed
    on {{ $schema }}.poi2$ (COALESCE(upper(block_range), 2147483647))
    where (COALESCE(upper(block_range), 2147483647) < 2147483647);

create index attr_12_0_poi2$_digest
    on {{ $schema }}.poi2$ (digest);

create index attr_12_1_poi2$_id
    on {{ $schema }}.poi2$ ("left"(id, 256));

create table if not exists {{ $schema }}.dynamic_data_source_xxx
(
	id text not null,
	context text not null,
	abi text not null,
	vid bigserial not null
		constraint dynamic_data_source_xxx_pkey
			primary key,
	block_range int4range not null,
	_updated_block_number numeric not null
);

alter table {{ $schema }}.dynamic_data_source_xxx owner to graph;

create index if not exists dynamic_data_source_xxx_block_range_closed
	on {{ $schema }}.dynamic_data_source_xxx (COALESCE(upper(block_range), 2147483647))
	where (COALESCE(upper(block_range), 2147483647) < 2147483647);

create index if not exists dynamic_data_source_xxx_id
	on {{ $schema }}.dynamic_data_source_xxx (id);

create index if not exists dynamic_data_source_xxx_abi
	on {{ $schema }}.dynamic_data_source_xxx (abi);

{{ $engine.BackTick }}

}


func(d *DDL) InitiateSchema(handleStatement func(statement string) error) error {
	err := handleStatement(d.schemaSetup)
	if err != nil {
		return fmt.Errorf("handle statement: %w", err)
	}
	return nil
}

func(d *DDL) CreateTables(handleStatement func(table string, statement string) error) error {
	for table, statement := range d.createTables {
		err := handleStatement(table, statement)
		if err != nil {
			return fmt.Errorf("handle statement: %w", err)
		}
	}
	return nil
}

func(d *DDL) CreateIndexes(handleStatement func(table string, statement string) error) error {
	for table, idxs := range d.indexes {
		for _, idx := range idxs {
			err := handleStatement(table, idx.createStatement)
			if err != nil {
				return fmt.Errorf("handle statement: %w", err)
			}
		}
	}
	return nil
}

func(d *DDL) DropIndexes(handleStatement func(table string, statement string) error) error {
	for table, idxs := range d.indexes {
		for _, idx := range idxs {
			err := handleStatement(table, idx.dropStatement)
			if err != nil {
				return fmt.Errorf("handle statement: %w", err)
			}
		}
	}
	return nil
}
`
