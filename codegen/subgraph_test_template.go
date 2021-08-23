package codegen

var templateSubgraphTest = `
var defaultTestTokens = map[string]*eth.Token{
	"0x00": {Address: []byte{0}, Name: "token.0.name", Symbol: "token.0.symbol", Decimals: 0, TotalSupply: big.NewInt(1000)},
	"0x01": {Address: []byte{1}, Name: "token.1.name", Symbol: "token.1.symbol", Decimals: 10, TotalSupply: big.NewInt(10000)},
	"0x02": {Address: []byte{2}, Name: "token.2.name", Symbol: "token.2.symbol", Decimals: 18, TotalSupply: big.NewInt(100000)},
}

func NewTestSubgraph(int subgraph.Intrinsics) *Subgraph {
	return &Subgraph{
		Base: subgraph.Base{
			Intrinsics: int,
			Definition: Definition,
			ID:         "testSubgraph",
			Log:        zlog,
		},
{{- if .HasDynamicTemplate }}
		CurrentBlockDynamicDataSources: map[string]*DynamicDataSourceXXX{},
		DynamicDataSources: map[string]*DynamicDataSourceXXX{},
{{- end }}
	}
}

type TestIntrinsics struct {
	tokens map[string]*eth.Token
	store  map[string]map[string]entity.Interface
	step   int
}

func NewTestIntrinsics(testCase *TestCase) *TestIntrinsics {
	i := &TestIntrinsics{
		tokens: make(map[string]*eth.Token),
		store:  make(map[string]map[string]entity.Interface),
		step:   99999,
	}

	if testCase != nil {
		i.initialize(testCase)
	}

	return i
}

func (i *TestIntrinsics) initialize(testCase *TestCase) {
	i.setStoreData(testCase.StoreData)
	i.setTokens(testCase.Tokens)
}

func (i *TestIntrinsics) setStoreData(ents []*TypedEntity) {
	for _, value := range ents {
		err := i.Save(value.Entity)
		if err != nil {
			panic(err)
		}
	}
}

func (i *TestIntrinsics) setTokens(tokens []*TokenInfo) {
	if len(tokens) == 0 {
		i.tokens = defaultTestTokens
		return
	}

	for _, tokenInfo := range tokens {
		token := &eth.Token{
			Name:        tokenInfo.Name,
			Symbol:      tokenInfo.Symbol,
			Address:     []byte{byte(tokenInfo.Address)},
			Decimals:    uint(tokenInfo.Decimals),
			TotalSupply: big.NewInt(int64(tokenInfo.TotalSupply)),
		}
		i.tokens[token.Address.Pretty()] = token
	}
}

func (i *TestIntrinsics) Save(e entity.Interface) error {
	tableName := entity.GetTableName(e)
	tbl, found := i.store[tableName]
	if !found {
		tbl = make(map[string]entity.Interface)
		i.store[tableName] = tbl
	}

	e.SetExists(true)
	e.SetMutated(i.step)

	tbl[e.GetID()] = e
	return nil
}

func (i *TestIntrinsics) Load(e entity.Interface) error {
	tableName := entity.GetTableName(e)
	tbl, found := i.store[tableName]
	if !found {
		return nil
	}

	id := e.GetID()
	cachedEntity, found := tbl[id]
	if found {
		if cachedEntity == nil {
			return nil
		}
		ve := reflect.ValueOf(e).Elem()
		ve.Set(reflect.ValueOf(cachedEntity).Elem())
		return nil
	}

	return nil
}

func (i *TestIntrinsics) LoadAllDistinct(e entity.Interface, blockNum uint64) ([]entity.Interface, error) {
	result := make([]entity.Interface, 0)

	tableName := entity.GetTableName(e)
	tbl, found := i.store[tableName]
	if !found {
		return result, nil
	}

	for _, v := range tbl {
		result = append(result, v)
	}
	return result, nil
}

func (i *TestIntrinsics) Remove(e entity.Interface) error {
	tableName := entity.GetTableName(e)
	tbl, found := i.store[tableName]
	if !found {
		return nil
	}

	id := e.GetID()
	delete(tbl, id)
	return nil
}

func (i *TestIntrinsics) Block() subgraph.BlockRef {
	return &blockRef{
		id:        "0x1",
		num:       1,
		timestamp: time.Time{},
	}
}

func (i *TestIntrinsics) StepBelow(step int) bool {
	return i.step < step
}

func (i *TestIntrinsics) StepAbove(step int) bool {
	return i.step > step
}

func (i *TestIntrinsics) GetTokenInfo(address eth.Address) (*eth.Token) {
    tok := i.tokens[address.Pretty()]
	return tok
}

type TestCase struct {
	StoreData []*TypedEntity         ` + "`" + `yaml:"storeData" json:"storeData"` + "`" + `
	Tokens    []*TokenInfo           ` + "`" + `yaml:"tokens" json:"tokens"` + "`" + `
	Events    []*TypedEvent          ` + "`" + `yaml:"events" json:"events"` + "`" + `
}

type TokenInfo struct {
	Address     int    ` + "`" + `yaml:"address" json:"address"` + "`" + `
	Name        string ` + "`" + `yaml:"name" json:"name"` + "`" + `
	Symbol      string ` + "`" + `yaml:"symbol" json:"symbol"` + "`" + `
	Decimals    int    ` + "`" + `yaml:"decimals" json:"decimals"` + "`" + `
	TotalSupply int    ` + "`" + `yaml:"total_supply" json:"total_supply"` + "`" + `
}

type TypedEntity struct {
	Type   string
	Entity entity.Interface
}

func (t *TypedEntity) UnmarshalJSON(data []byte) error {
	s := &struct {
		Type   string          ` + "`" + `json:"type" yaml:"type"` + "`" + `
		Entity json.RawMessage ` + "`" + `json:"entity" yaml:"entity"` + "`" + `
	}{}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ent entity.Interface
	switch s.Type {
{{- range $ent := .Entities }}
	case "{{ $engine.ToLowerSnakeCase $ent.Name }}":
		tempEnt := &{{ $ent.Name }}{}
		err := json.Unmarshal(s.Entity, &tempEnt)
		if err != nil {
			return err
		}
		ent = tempEnt
{{- end }}
	}

	t.Entity = ent
	t.Type = s.Type

	return nil
}

type TypedEvent struct {
	Type  string
	Event interface{}
}

func (t *TypedEvent) UnmarshalJSON(data []byte) error {
	s := &struct {
		Type  string
		Event json.RawMessage
	}{}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var event interface{}
	switch s.Type {
{{- range $source, $defs := .EventDefs }}
{{- range $defs }}
	case "{{ $source }}{{ .Name }}Event":
		ev := &{{ $source }}{{ .Name }}Event{}
		err := json.Unmarshal(s.Event, &ev)
		if err != nil {
			return err
		}
		event = ev
{{- end }}
{{- end }}
	}

	t.Event = event
	t.Type = s.Type

	return nil
}
`
