package codegen

var templatesEntities = `
{{ range $enum := $.Enums }}
var (
	{{- range $field := $enum.Fields }}
	{{ $enum.Name }}{{ $.ToCamelCase $field }} = entity.Enum("{{ $.ToCamelCase $field }}")
	{{- end }}
)
{{ end }}

{{ range $ent := $.Entities }}
// {{ $ent.Name }}
type {{ $ent.Name }} struct {
	entity.Base

{{- range $field := $ent.Fields.VisibleOnly }}
	{{ $.ToCamelCase $field.Name }} {{ $field.GoType }} ` + "`" + `db:"{{ $.ToLowerSnakeCase $field.Name }}{{ if $field.Nullable }},nullable{{ end }}" csv:"{{ $engine.ToLowerSnakeCase $field.Name }}"` + "`" + `
{{- end }}
}

func New{{ $ent.Name }}(id string) *{{ $ent.Name }} {
	return &{{ $ent.Name }}{
		Base: entity.NewBase(id),
{{- range $field := $ent.Fields.InitOnly }}
		{{ $.ToCamelCase $field.Name }}: {{ $field.GoInit }},
{{- end }}
	}
}

func (_ *{{ $ent.Name }}) SkipDBLookup() bool {
  return {{$ent.CacheSkipDBLookup}}
}

{{- if ne $.HighestParallelStep 0 }}
func (next *{{ $ent.Name }}) Merge(step int, cached *{{ $ent.Name }}) {
{{- range $step := $.IterateSteps }}
{{- if or (ne (len (index $ent.ParallelSetterSteps $step)) 0) (ne (len (index $ent.ParallelAggregatorSteps $step)) 0) }}
	if step == {{ $.PlusOne $step }}  {
{{- range $pp := index $ent.ParallelAggregatorSteps $step }}
{{- $fieldName := $.ToCamelCase $pp.Field.Name }}
		{{- if and (eq $pp.Field.GoType "entity.Int") (eq $pp.Type "SUM") }}
		next.{{ $fieldName }} = entity.IntAdd(next.{{ $fieldName }}, cached.{{ $fieldName }})
		{{- else if and (eq $pp.Field.GoType "entity.Float") (eq $pp.Type "SUM") }}
		next.{{ $fieldName }} = entity.FloatAdd(next.{{ $fieldName }}, cached.{{ $fieldName }})
		{{- else if and (eq $pp.Field.GoType "int64") (eq $pp.Type "SUM") }}
		next.{{ $fieldName }} += cached.{{ $fieldName }}
		{{- else if and (eq $pp.Field.GoType "float64") (eq $pp.Type "SUM") }}
		next.{{ $fieldName }} += cached.{{ $fieldName }}
		{{- else if and (eq $pp.Field.GoType "entity.LocalStringArray") (eq $pp.Type "SUM") }}
		next.{{ $fieldName }} = entity.ArrayAdd(next.{{ $fieldName }}, cached.{{ $fieldName }})
		{{- else }}
		next.{{ $fieldName }} = cached.{{ $fieldName }}
		{{- end }}
{{- end }}
		if next.MutatedOnStep != {{ $step }} {
{{- range $pp := index $ent.ParallelSetterSteps $step }}
{{- $fieldName := $.ToCamelCase $pp.Field.Name }}
				next.{{ $fieldName }} = cached.{{ $fieldName }}
{{- end }}
		}
	}
{{- end }}
{{- end }}
}
{{- end }}

{{ end }}
`
