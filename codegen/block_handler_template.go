package codegen

var templateHandleBlock = `
{{ $allHandlers := .Handlers }}
{{ $factorySource := .FactorySource }}
{{ $hasDynamicTemplate := .HasDynamicTemplate }}

func (s *Subgraph) HandleBlock(block *pbcodec.Block) error {
	idx := uint32(0)
	s.CurrentBlockDynamicDataSources = make(map[string]*DynamicDataSourceXXX)
		
	for _, trace := range block.TransactionTraces {
		logs := trace.Logs()
		for _, log := range logs {
			var ethLog interface{} = log
			eventLog := codecLogToEthLog(ethLog.(*pbcodec.Log), idx)
			idx++
{{- if $hasDynamicTemplate }}
	if bytes.Equal(FactoryAddressBytes, log.Address) || s.IsDynamicDataSource(eth.Address(log.Address).Pretty()) {
{{- else }}
	if bytes.Equal(FactoryAddressBytes, log.Address) {
{{- end }}
				ev, err := DecodeEvent(eventLog, block, trace)
				if err != nil {
					return fmt.Errorf("parsing event: %w", err)
				}
				if err := s.HandleEvent(ev); err != nil {
					return fmt.Errorf("handling event: %w", err)
				}

			}	
		}
	}

	if len(s.CurrentBlockDynamicDataSources) == 0 {
		return nil
	}

	for _, trace := range block.TransactionTraces {
		logs := trace.Logs()
		for _, log := range logs {
			var ethLog interface{} = log
			eventLog := codecLogToEthLog(ethLog.(*pbcodec.Log), idx)
			idx++
			if s.IsCurrentDynamicDataSource(eth.Address(log.Address).Pretty()) {
				ev, err := DecodeEvent(eventLog, block, trace)
				if err != nil {
					return fmt.Errorf("parsing event: %w", err)
				}
				if err := s.HandleEvent(ev); err != nil {
					return fmt.Errorf("handling event: %w", err)
				}

			}
		}
	}

	for k, v := range s.CurrentBlockDynamicDataSources {
		s.DynamicDataSources[k] = v
	}

	return nil
}
func (s *Subgraph) HandleEvent(ev interface{}) error {
	switch e := ev.(type) {
	{{ $factoryHandlers := index $allHandlers $factorySource}}
		{{- range $handler := $factoryHandlers }}
		case *{{ $factorySource }}{{ $handler.Name }}Event:
			if err := s.{{ $handler.HandlerFuncName }}(e); err != nil {
				return fmt.Errorf("handling {{ $factorySource }}{{ $handler.Name }} event: %w", err)
			}
		{{- end }}
		{{- range $source := .Templates }}
		{{ $sourcedHandlers := index $allHandlers $source }}
		{{- range $handler := $sourcedHandlers }}
		case *{{ $source }}{{ $handler.Name }}Event:
			if err := s.{{ $handler.HandlerFuncName }}(e); err != nil {
				return fmt.Errorf("handling {{ $source }}{{ $handler.Name }} event: %w", err)
			}
		{{- end }}
		{{- end }}
	}

	return nil
}

func codecLogToEthLog(l *pbcodec.Log, idx uint32) *eth.Log {
	return &eth.Log{
		Address:    l.Address,
		Topics:     l.Topics,
		Data:       l.Data,
		Index:      l.Index,
		BlockIndex: idx,
	}
}
`
