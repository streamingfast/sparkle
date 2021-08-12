package codegen

var templateDDS = `
{{- if .HasDynamicTemplate }}
type DynamicDataSourceXXX struct {
	entity.Base

    Context string ` + "`" + `db:"context" csv:"context"` + "`" + `
	ABI string ` + "`" + `db:"abi" csv:"abi"` + "`" + `
}

func NewDynamicDataSource(id string, abi string, context string) *DynamicDataSourceXXX {
	return &DynamicDataSourceXXX{
		Base:     entity.NewBase(id),
		Context: context,
		ABI: abi,
	}
}

{{- range $tpl := .Templates }}
func (s *Subgraph) Create{{ $tpl }}Template(address eth.Address, obj interface{}) error {
	cnt, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	ds := NewDynamicDataSource(address.Pretty(), "{{ $tpl }}", string(cnt))

	if err := s.Save(ds); err != nil {
		return fmt.Errorf("saving {{ $tpl }} datasource: %w", err)
	}

	s.CurrentBlockDynamicDataSources[address.Pretty()] = ds

	return nil
}

{{- end }}
{{- end }}

func (s *Subgraph) IsDynamicDataSource(address string) bool {
	_, ok := s.DynamicDataSources[address]
	return ok
}

func (s *Subgraph) IsCurrentDynamicDataSource(address string) bool {
	_, ok := s.CurrentBlockDynamicDataSources[address]
	return ok
}


func (s *Subgraph) LoadDynamicDataSources(blockNum uint64) error {
{{- if .HasDynamicTemplate }}
	res, err := s.LoadAllDistinct(&DynamicDataSourceXXX{}, blockNum)
	if err != nil {
		return fmt.Errorf("loading dynamic data sources: %w", err)
	}

{{- range $tpl := .Templates }}
	for _, dsi := range res {
		ds := dsi.(*DynamicDataSourceXXX)
		if ds.ABI == "{{ $tpl }}" {
			s.DynamicDataSources[ds.GetID()] = ds
		}
	}
{{- end }}
{{- end }}
	return nil
}

`
