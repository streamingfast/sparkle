package codegen

var templateEvents = `
{{- range $source, $defs := .EventDefs }}
	// {{ $source }}

{{- range $defs }}
// {{ $source }}{{ .Name }} event

type {{ $source }}{{ .Name }}Event struct {
	*entity.BaseEvent
	LogAddress eth.Address
	LogIndex int

	// Fields
{{- range $input := .Inputs.NamedOnly }}
	{{ $engine.ToCamelCase $input.Name }} {{ $engine.MapEventType $input.Type }} ` + "`" + `eth:"{{ if $input.Indexed }},indexed{{ end }}"` + "`" + `
{{- end }}
}

var hash{{ $source }}{{ .Name }}Event = eth.Keccak256([]byte("
  {{- .Name }}(
    {{- range $idx, $input := .Inputs }}{{ if ne $idx 0 }},{{end}}{{ $input.Type }}{{ end -}}
  )"))

func Is{{ $source }}{{ .Name }}Event(log *eth.Log) bool {
	return bytes.Equal(log.Topics[0], hash{{ $source }}{{ .Name }}Event)
}

func New{{ $source }}{{ .Name }}Event(log *eth.Log, block *pbcodec.Block, trace *pbcodec.TransactionTrace) (*{{ $source }}{{ .Name }}Event, error) {
	var err error
	ev := &{{ $source }}{{ .Name }}Event{
		BaseEvent: &entity.BaseEvent{},
		LogAddress: log.Address,
		LogIndex: int(log.BlockIndex),
	}

	ev.SetBlockAndTransaction(block, trace)

	dec := eth.NewLogDecoder(log)
	if _, err := dec.ReadTopic(); err != nil {
		return nil, fmt.Errorf("reading topic 0: %w", err)
	}

{{- range $idx, $input := .Inputs.NamedOnly }}

{{- if .Indexed }}
	f{{ $idx }}, err := dec.ReadTypedTopic("{{ $input.Type }}")
	if err != nil {
		return nil, fmt.Errorf("reading {{ $input.Name }}: %w", err)
	}
	ev.{{ $engine.ToCamelCase $input.Name }} = f{{ $idx }}.({{ $engine.MapEventType $input.Type }})
{{- else }}
	ev.{{ $engine.ToCamelCase $input.Name }}, err = dec.DataDecoder.{{ $engine.MapEventDecoder $input.Type }}()
	if err != nil {
		return nil, fmt.Errorf("reading {{ $input.Name }}:  %w", err)
	}
{{- end }}

{{- end }}
	return ev, nil
}

{{- end}}
{{- end }}


func DecodeEvent(log *eth.Log, block *pbcodec.Block, trace *pbcodec.TransactionTrace) (interface{}, error) {
{{ range $source, $defs := .EventDefs }}
{{- range $defs }}
	if Is{{ $source }}{{ .Name }}Event(log) {
		ev, err := New{{ $source }}{{ .Name }}Event(log, block, trace)
		if err != nil {
			return nil, fmt.Errorf("decoding {{ $source }}{{ .Name }} event: %w", err)
		}
		return ev, nil
	}
{{- end }}
{{ end }}
	return nil, nil
}
`
