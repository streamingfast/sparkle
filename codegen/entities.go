package codegen

import (
	"fmt"
	"strconv"

	"github.com/vektah/gqlparser/ast"
)

//TODO: get rid of this global variable
var enumTypes map[string]bool

func init() {
	enumTypes = make(map[string]bool)
}

type Definition struct {
	Fields   Fields
	Name     string
	IsEntity bool

	CacheSkipDBLookup bool

	ParallelSetterSteps     map[int][]*ParallelStep
	ParallelAggregatorSteps map[int][]*ParallelStep
}

type Enum struct {
	Name   string
	Fields []string
}

type Field struct {
	Name          string
	Type          string
	GoType        string
	GoInit        string
	Init          string
	Nullable      bool
	Array         bool
	Derived       bool
	Hidden        bool
	PostgresType  string
	PostgresIndex string
	ParallelStep  *ParallelStep
}

type ParallelStep struct {
	Step int
	Type string

	Field *Field
}

type Fields []*Field
type Enums []*Enum

func (fs Fields) VisibleOnly() (out []*Field) {
	for _, f := range fs {
		if !f.Derived && !f.Hidden {
			out = append(out, f)
		}
	}
	return
}

func (fs Fields) InitOnly() (out []*Field) {
	for _, f := range fs {
		if f.GoInit != "" && !f.Derived && !f.Hidden {
			out = append(out, f)
		}
	}
	return
}

func ParseEnum(def *ast.Definition) *Enum {
	if def.Kind != "ENUM" {
		return nil
	}

	vals := make([]string, 0, len(def.EnumValues))
	for _, val := range def.EnumValues {
		vals = append(vals, val.Name)
	}

	return &Enum{
		Name:   def.Name,
		Fields: vals,
	}
}

func parseObject(def *ast.Definition) (*Definition, error) {
	if def.Kind != "OBJECT" {
		return nil, nil
	}
	fields := make([]*Field, 0, len(def.Fields))
	for _, field := range def.Fields {
		fieldDef, err := ParseFieldDefinition(field)
		if err != nil {
			return nil, fmt.Errorf("entity %q: %w", def.Name, err)
		}
		fields = append(fields, fieldDef)
	}

	out := &Definition{
		Fields: fields,
		Name:   def.Name,
	}

	for _, dir := range def.Directives {
		if dir.Name == "entity" {
			out.IsEntity = true
		}
	}

	if !out.IsEntity {
		return out, nil
	}

	// this is only applied for entities
	for _, dir := range def.Directives {
		if dir.Name == "cache" {
			for _, arg := range dir.Arguments {
				switch arg.Name {
				case "skip_db_lookup":
					if arg.Value == nil {
						return nil, fmt.Errorf("'skip_db_lookup' argument to @cache directive requires a boolean parameter")
					}
					val, err := strconv.ParseBool(arg.Value.Raw)
					if err != nil {
						return nil, fmt.Errorf("invalid bool value for 'skip_db_lookup' argument to @cache directive: %w", err)
					}
					if val {
						out.CacheSkipDBLookup = true
					}
				}
			}
		}
	}

	return out, nil
}

func ParseFieldDefinition(field *ast.FieldDefinition) (*Field, error) {
	f := &Field{
		Name:   field.Name,
		Type:   field.Type.Name(),
		Array:  bool(field.Type.Elem != nil),
		Hidden: field.Name == "id", // defined in entity.Base, so not needed in our codegen
	}
	if field.Type.Elem != nil {
		f.Nullable = !field.Type.Elem.NonNull
	} else {
		f.Nullable = !field.Type.NonNull
	}

	goType, err := f.getGoType()
	if err != nil {
		return nil, fmt.Errorf("invalid type %q for field %q: %w", field.Type.Name(), field.Name, err)
	}
	f.GoType = goType
	f.GoInit = f.getGoInit()
	f.PostgresType = f.getPostgresType()
	f.PostgresIndex = f.getPostgresIndex()

	for _, directive := range field.Directives {
		if directive.Name == "sql" {
			for _, arg := range directive.Arguments {
				switch arg.Name {
				case "index":
					if arg.Value == nil {
						return nil, fmt.Errorf("'index' argument to @sql directive requires a boolean param on field %s", f.Name)
					}
					val, err := strconv.ParseBool(arg.Value.Raw)
					if err != nil {
						return nil, fmt.Errorf("invalid bool value for 'index' argument to @sql directive for field %s: %w", f.Name, err)
					}
					if !val {
						f.PostgresIndex = "" // remove sql index
					}
				default:
					return nil, fmt.Errorf("invalid argument %q for directive @%q on field %s", arg.Name, directive.Name, f.Name)
				}
			}
		}
		if directive.Name == "derivedFrom" {
			f.Derived = true
		}
		if directive.Name == "parallel" {
			// compute any missing parallel notations, and error out if some are missing.
			step, err := toParallelStep(f, directive)
			if err != nil {
				return nil, err
			}
			f.ParallelStep = step
		}
	}

	return f, nil
}

func toParallelStep(field *Field, directive *ast.Directive) (*ParallelStep, error) {
	step := &ParallelStep{
		Field: field,
	}
	for _, arg := range directive.Arguments {
		switch arg.Name {
		case "step":
			if arg.Value == nil {
				return nil, fmt.Errorf("'step' argument to @parallel directive requires a numeric parameter in field %q", field.Name)
			}
			val, err := strconv.ParseInt(arg.Value.Raw, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer value for 'step' argument to @parallel directive: %w", err)
			}
			step.Step = int(val)

		case "type":
			if arg.Value == nil {
				return nil, fmt.Errorf("'type' argument to @parallel directive requires a parameter in field %q, supported values: SUM", field.Name)
			}
			if arg.Value.Raw != "SUM" {
				return nil, fmt.Errorf("invalid type %q for @parallel directive in field %q", arg.Value.Raw, field.Name)
			}
			step.Type = arg.Value.Raw
		default:
			return nil, fmt.Errorf("invalid @parallel directive argument %q, in field %q", arg.Name, field.Name)
		}
	}

	return step, nil
}

func (f *Field) getGoType() (string, error) {
	var dt string

	if t, ok := map[string]string{
		"ID":         "string",
		"String":     "string",
		"Int":        "int64",
		"Float":      "float64",
		"Boolean":    "entity.Bool",
		"BigInt":     "entity.Int",
		"BigDecimal": "entity.Float",
		"Bytes":      "entity.Bytes",
	}[f.Type]; ok {
		dt = t
	} else if _, ok = enumTypes[f.Type]; ok {
		dt = "entity.Enum"
	}

	res := ""
	if f.Array {
		res += "[]"
	}

	neverPointers := map[string]bool{
		"Bytes": true,
	}

	if _, ok := neverPointers[f.Type]; !ok && f.Nullable && !f.Array {
		res += "*"
	}

	if dt == "" {
		if res == "[]" {
			return "entity.LocalStringArray", nil
		}
		// FIXME: this allows us to define types like BigFloat or BigWhatever which aren't
		// defined as enties.
		// We should catch that at a later step, and fail when types are not part of the basic types nor defined in the GraphQL schema.
		res += "string"
		//return "", fmt.Errorf("%q is not a supported type in sparkle", f.Type)
	} else {
		res += dt
	}

	return res, nil
}

func (f *Field) getGoInit() string {
	initFunc := map[string]string{
		"entity.Int":   "IL(0)",
		"entity.Float": "FL(0)",
		"entity.Bytes": "[]byte{}",
	}[f.GoType]

	return initFunc
}
