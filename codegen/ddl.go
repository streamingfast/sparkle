package codegen

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

var PostgresSchema = "%%SCHEMA%%"

var fieldTypesNotNullable = map[string]string{
	"String":     "text not null",
	"Boolean":    "boolean not null",
	"Bytes":      "bytea not null",
	"Int":        "numeric not null",
	"Float":      "numeric not null",
	"BigInt":     "numeric not null",
	"BigDecimal": "numeric not null",
}
var fieldTypesNullable = map[string]string{
	"String":     "text",
	"Boolean":    "boolean",
	"Bytes":      "bytea",
	"Int":        "numeric",
	"Float":      "numeric",
	"BigInt":     "numeric",
	"BigDecimal": "numeric",
}

func (f *Field) getPostgresType() string {

	switch {
	case f.Nullable:
		if dt, ok := fieldTypesNullable[f.Type]; ok {
			return dt
		}
	default:
		if dt, ok := fieldTypesNotNullable[f.Type]; ok {
			return dt
		}
	}

	if f.Array {
		if f.Nullable {
			return "text[]"
		}
		return "text[] not null"
	}

	if _, ok := enumTypes[f.Type]; ok {
		return fmt.Sprintf(`%s."%s"`, PostgresSchema, strings.ToLower(strcase.ToSnake(f.Name)))
	}

	//this should be an ID
	if f.Nullable {
		return "text"
	}
	return "text not null"

}

func (f *Field) getPostgresIndex() string {
	//CREATE INDEX attr_4_5_transaction_swaps ON sgd3.transaction USING gin (swaps);

	_, ok := map[string]bool{
		"ID":         true,
		"Boolean":    true,
		"Bytes":      true,
		"Int":        true,
		"Float":      true,
		"BigInt":     true,
		"BigDecimal": true,
	}[f.Type]

	snakeName := fmt.Sprintf(`"%s"`, strings.ToLower(strcase.ToSnake(f.Name)))

	if ok {
		return fmt.Sprintf("using btree (%s)", snakeName)
	}

	if f.Type == "String" {
		return fmt.Sprintf(`("left"(%s, 256))`, snakeName)
	}

	if f.Array {
		return fmt.Sprintf("using gin (%s)", strings.ToLower(strcase.ToSnake(f.Name)))
	}

	if _, ok = enumTypes[f.Type]; ok {
		return fmt.Sprintf("using btree (%s)", snakeName)
	}

	return fmt.Sprintf("using gist (%s, block_range)", snakeName)

}
