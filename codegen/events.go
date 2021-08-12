package codegen

type EventFields []*EventField

func (a EventFields) NamedOnly() (out []*EventField) {
	for _, el := range a {
		if el.Name != "" {
			out = append(out, el)
		}
	}
	return
}

type EventDef struct {
	Anonymous bool
	Inputs    EventFields
	Name      string
	Type      string
}

type EventField struct {
	Indexed      bool
	InternalType string
	Name         string
	Type         string
}
