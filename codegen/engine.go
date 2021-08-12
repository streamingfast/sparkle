package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/streamingfast/sparkle/manifest"

	"github.com/iancoleman/strcase"
	"github.com/streamingfast/sparkle/zapbox"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser"
	"gopkg.in/yaml.v3"
)

type handler struct {
	Name            string
	Signature       string
	HandlerName     string
	HandlerFuncName string
}

type TemplateEngine struct {
	PackageName  string
	GoModulePath string
	Schema       string

	StartBlock     int
	FactorySource  string
	FactoryAddress string

	Manifest      string
	GraphQLSchema string
	Abis          map[string]string

	Entities []*Definition
	Enums    Enums

	HighestParallelStep int

	HasDynamicTemplate bool

	EventDefs map[string][]*EventDef
	Handlers  map[string][]*handler
	Templates []string

	CurrentSource      string
	CurrentEvent       string //super sorry!
	CurrentHandlerName string //again, really sorry.

	basePath         string
	subgraphManifest *manifest.SubgraphManifest
	graphqlSchemaDoc *ast.SchemaDocument
	userLog          *zapbox.CLILogger
}

func NewEngine(yamlFilePath string, GoModulePath string, userLog *zapbox.CLILogger) (*TemplateEngine, error) {
	e := &TemplateEngine{
		GoModulePath: GoModulePath,
		userLog:      userLog,
		Schema:       PostgresSchema,
		Abis:         map[string]string{},

		EventDefs: map[string][]*EventDef{},
		Handlers:  map[string][]*handler{},
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("can't get cwd: %w", err)
	}
	e.basePath = currentDir

	userLog.Printf("Code will be generated to: %q", currentDir)

	manifestContent, subgraphManifest, err := manifest.DecodeYamlManifestFromFile(yamlFilePath)
	if err != nil {
		return nil, err
	}
	e.Manifest = manifestContent
	e.subgraphManifest = subgraphManifest

	yamlFile := filepath.Base(yamlFilePath)
	yamlFileName := strings.TrimSuffix(yamlFile, filepath.Ext(yamlFile))
	e.PackageName = yamlFileName

	e.FactorySource = e.subgraphManifest.DataSources[0].Source.Abi
	e.StartBlock = e.subgraphManifest.DataSources[0].Source.StartBlock
	e.FactoryAddress = e.subgraphManifest.DataSources[0].Source.Address

	e.HasDynamicTemplate = len(e.subgraphManifest.Templates) > 0
	for _, tpl := range e.subgraphManifest.Templates {
		e.Templates = append(e.Templates, tpl.Name)
	}

	schemaFilePath, err := filepath.Abs(filepath.Join(filepath.Dir(yamlFilePath), e.subgraphManifest.Schema.File.Path))
	if err != nil {
		return nil, fmt.Errorf("getting abs path: %w", err)
	}

	for _, datasource := range subgraphManifest.DataSources {
		for _, abi := range datasource.Mapping.Abis {
			if err := e.readABI(yamlFilePath, abi); err != nil {
				return nil, fmt.Errorf("unable to process datasources abi: %w", err)
			}
		}
	}

	for _, template := range subgraphManifest.Templates {
		for _, abi := range template.Mapping.Abis {
			if err := e.readABI(yamlFilePath, abi); err != nil {
				return nil, fmt.Errorf("unable to process templates abi: %w", err)
			}
		}
	}

	graphqlSchemaContent, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	e.GraphQLSchema = string(graphqlSchemaContent)

	graphqlSchemaDoc, gqlerr := parser.ParseSchema(&ast.Source{
		Input: string(graphqlSchemaContent),
	})
	if gqlerr != nil {
		return nil, fmt.Errorf("parsing schema: %w", err)
	}
	e.graphqlSchemaDoc = graphqlSchemaDoc

	//todo: replace yamlFilePath abis folder
	err = e.decodeEventDefs(yamlFilePath)
	if err != nil {
		return nil, err
	}

	e.decodeEnums()
	if err := e.decodeEntities(); err != nil {
		return nil, fmt.Errorf("entities: %w", err)
	}
	e.decodeHandlers()

	if err := e.computeSteps(); err != nil {
		return nil, fmt.Errorf("computing total number of steps: %w", err)
	}

	userLog.Printf("Subgraph decoding completed")

	return e, nil
}

func (t *TemplateEngine) decodeYamlManifest(yamlFilePath string) error {
	yamlContent, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return fmt.Errorf("reading subgraph file %q: %w", yamlFilePath, err)
	}

	var subgraphManifest *manifest.SubgraphManifest
	if err := yaml.NewDecoder(bytes.NewReader(yamlContent)).Decode(&subgraphManifest); err != nil {
		return fmt.Errorf("decofing subgraph file %q: %w", yamlFilePath, err)
	}

	t.subgraphManifest = subgraphManifest
	return nil
}

func (t *TemplateEngine) decodeEnums() {
	enums := make([]*Enum, 0, len(t.graphqlSchemaDoc.Definitions))
	for _, d := range t.graphqlSchemaDoc.Definitions {
		if def := ParseEnum(d); def != nil {
			enums = append(enums, def)
			enumTypes[d.Name] = true
		}
	}

	t.Enums = enums
}

func (t *TemplateEngine) decodeEntities() error {
	defs := make([]*Definition, 0, len(t.graphqlSchemaDoc.Definitions))
	for _, d := range t.graphqlSchemaDoc.Definitions {
		def, err := parseObject(d)
		if err != nil {
			return err
		}
		if def != nil && def.IsEntity {
			defs = append(defs, def)
		}
	}

	t.Entities = defs

	return nil
}

func (t *TemplateEngine) computeSteps() error {
	var highestStep int
	//var totalFields int
	for _, ent := range t.Entities {
		setterSteps := map[int][]*ParallelStep{}
		aggSteps := map[int][]*ParallelStep{}
		for _, field := range ent.Fields {
			//totalFields++
			if field.ParallelStep == nil {
				continue
			}
			ppStep := field.ParallelStep
			if ppStep.Step > highestStep {
				highestStep = ppStep.Step
			}

			if ppStep.Type == "" {
				setterSteps[ppStep.Step] = append(setterSteps[ppStep.Step], ppStep)
			} else {
				aggSteps[ppStep.Step] = append(aggSteps[ppStep.Step], ppStep)
			}
		}
		ent.ParallelSetterSteps = setterSteps
		ent.ParallelAggregatorSteps = aggSteps
	}

	// TODO: make sure if `highestStep` is non-zero, that we have a @parallel directive
	// defined on ALL of the fields where it makes sense (skip the @derivedFrom, the ID, etc..)

	t.HighestParallelStep = highestStep
	return nil
}

func (t *TemplateEngine) decodeHandlers() {
	eventsMap := map[string]map[string]string{}
	/// gather all event handlers
	eventsMap[t.subgraphManifest.DataSources[0].Source.Abi] = map[string]string{}
	for _, eventHandlerDef := range t.subgraphManifest.DataSources[0].Mapping.EventHandlers {
		eventsMap[t.subgraphManifest.DataSources[0].Source.Abi][eventHandlerDef.Handler] = eventHandlerDef.Event
	}

	for _, tpl := range t.subgraphManifest.Templates {
		eventsMap[tpl.Source.Abi] = map[string]string{}
		for _, eventHandlerDef := range tpl.Mapping.EventHandlers {
			eventsMap[tpl.Source.Abi][eventHandlerDef.Handler] = eventHandlerDef.Event
		}
	}

	for source, handlerMap := range eventsMap {
		for handlerName, signature := range handlerMap {
			eventName := strings.Split(signature, "(")[0]
			handlerFuncName := fmt.Sprintf("Handle%s%sEvent", source, eventName)
			t.Handlers[source] = append(t.Handlers[source], &handler{
				Name:            eventName,
				Signature:       signature,
				HandlerName:     handlerName,
				HandlerFuncName: handlerFuncName,
			})
		}

		sort.Slice(t.Handlers[source], func(i, j int) bool {
			return t.Handlers[source][i].Name < t.Handlers[source][j].Name
		})
	}
}

func (t *TemplateEngine) decodeEventDefs(yamlFilePath string) error {
	dataSource := t.subgraphManifest.DataSources[0]

	requiredABI := map[string]bool{}
	abiFileMap := map[string]string{}

	abiFileMapx := map[string]map[string]string{}
	abiFileMapx[dataSource.Source.Abi] = map[string]string{}

	// TODO: determine required ABI files to read
	requiredABI[dataSource.Source.Abi] = true
	for _, tpl := range t.subgraphManifest.Templates {
		requiredABI[tpl.Source.Abi] = true
		abiFileMapx[tpl.Source.Abi] = map[string]string{}
	}

	/// gather all abi files
	for _, abiLoc := range dataSource.Mapping.Abis {
		abiFileMap[abiLoc.Name] = abiLoc.File.Path
		abiFileMapx[dataSource.Source.Abi][abiLoc.Name] = abiLoc.File.Path
	}

	for _, abiLoc := range dataSource.Mapping.Abis {
		abiFileMap[abiLoc.Name] = abiLoc.File.Path
	}

	for _, tpl := range t.subgraphManifest.Templates {
		for _, abiLoc := range tpl.Mapping.Abis {
			abiFileMap[abiLoc.Name] = abiLoc.File.Path
			abiFileMapx[tpl.Source.Abi][abiLoc.Name] = abiLoc.File.Path
		}
	}

	for source, abiMap := range abiFileMapx {
		for abiName, relativeAbiFilePath := range abiMap {
			if abiName != source {
				continue
			}

			abiFilePath, err := filepath.Abs(filepath.Join(filepath.Dir(yamlFilePath), relativeAbiFilePath))
			if err != nil {
				return fmt.Errorf("locating abi file: %w", err)
			}

			abiContent, err := ioutil.ReadFile(abiFilePath)
			if err != nil {
				return fmt.Errorf("reading abi: %w", err)
			}

			var types []*EventDef
			if err := json.Unmarshal(abiContent, &types); err != nil {
				return fmt.Errorf("unmarshal: %w", err)
			}

			for _, e := range types {
				if e.Type == "event" {
					t.EventDefs[source] = append(t.EventDefs[source], e)
				}
			}

			sort.Slice(t.EventDefs[source], func(i, j int) bool {
				return t.EventDefs[source][i].Name < t.EventDefs[source][j].Name
			})
		}
	}

	return nil
}

func (t *TemplateEngine) GenerateCode(noGoMod bool) error {
	packagePath := path.Join(t.basePath, t.PackageName)
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		err = os.MkdirAll(packagePath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating package directory %q: %w", packagePath, err)
		}
	}

	cmdPath := path.Join(t.basePath, "cmd", t.PackageName)
	if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
		err = os.MkdirAll(cmdPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating cmd directory %q: %w", packagePath, err)
		}
	}

	mainGoFile := filepath.Join("cmd", t.PackageName, "/main.go")
	if err := t.executeTemplate(templateSubgraphMain, mainGoFile, true, false); err != nil {
		return fmt.Errorf("executing template %q: %w", mainGoFile, err)
	}

	if noGoMod {
		t.userLog.Printf("Skipping go.mod generation")
	} else {
		goModFile := "/go.mod"
		if err := t.executeTemplate(templateSubgraphMainGoMod, goModFile, false, false); err != nil {
			return fmt.Errorf("executing template %q: %w", goModFile, err)
		}
	}

	generatedTemplate := templateBase +
		templateSubgraph +
		templatesEntities +
		templateHandleBlock +
		templateEvents +
		templateDDS +
		templateDDL +
		templateSubgraphTest

	generatedGoFile := filepath.Join(t.PackageName, "generated.go")
	if err := t.executeTemplate(generatedTemplate, generatedGoFile, true, true); err != nil {
		return fmt.Errorf("executing template %q: %w", generatedGoFile, err)
	}

	entitesGoFile := filepath.Join(t.PackageName, "entities.go")
	if err := t.executeTemplate(templateMerge, entitesGoFile, true, false); err != nil {
		return fmt.Errorf("executing template %q: %w", entitesGoFile, err)
	}

	subgraphExtGoFile := filepath.Join(t.PackageName, "subgraph.go")
	if err := t.executeTemplate(templateSubgraphExt, subgraphExtGoFile, true, false); err != nil {
		return fmt.Errorf("executing template %q: %w", subgraphExtGoFile, err)
	}

	for source, handlers := range t.Handlers {
		for _, h := range handlers {
			t.CurrentSource = source
			t.CurrentHandlerName = h.HandlerFuncName
			t.CurrentEvent = fmt.Sprintf("%s%s", source, h.Name)

			handlerGoFile := filepath.Join(t.PackageName, t.ToLowerSnakeCase(h.HandlerFuncName)+".go")
			if err := t.executeTemplate(templateHandler, handlerGoFile, true, false); err != nil {
				return fmt.Errorf("executing template %q: %w", handlerGoFile, err)
			}
		}
	}

	return nil
}

func (t *TemplateEngine) executeTemplate(tpl string, fileName string, goFormat bool, overwrite bool) error {
	_, err := os.Stat(filepath.Join(t.basePath, fileName))
	if !os.IsNotExist(err) && !overwrite {
		t.userLog.Printf("Skipping %q as it already exists", fileName)
		return nil
	}

	f, err := os.Create(filepath.Join(t.basePath, fileName))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	tp, err := template.New(tpl).Parse(tpl)
	if err != nil {
		ioutil.WriteFile("/tmp/parsed.txt", []byte(tpl), 0644)
		return fmt.Errorf("parsing template: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	err = tp.Execute(buf, t)
	if err != nil {
		return fmt.Errorf("template execution: %w", err)
	}

	code := buf.Bytes()
	if goFormat {
		ioutil.WriteFile("/tmp/generated.go", []byte(code), 0644)
		code, err = format.Source(code)
		if err != nil {
			return fmt.Errorf("formating code: %w", err)
		}

	}
	_, err = f.Write(code)
	if err != nil {
		return fmt.Errorf("writing code to file: %w", err)
	}

	t.userLog.Printf("%q generated", fileName)

	return nil
}

func (t *TemplateEngine) ToCamelCase(input string) string {
	return strcase.ToCamel(input)
}

func (t *TemplateEngine) PlusOne(input int) int {
	return input + 1
}

func (t *TemplateEngine) ToLowerCamelCase(input string) string {
	return strcase.ToLowerCamel(input)
}

func (t *TemplateEngine) ToLowerSnakeCase(input string) string {
	return strings.ToLower(strcase.ToSnake(input))
}

func (t *TemplateEngine) FirstLetter(input string) string {
	return strings.ToLower(string([]byte(input)[0]))
}

func (t *TemplateEngine) MapEventType(input string) string {
	return map[string]string{
		"uint256": "*big.Int",
		"int256":  "*big.Int",
		"uint112": "*big.Int",
		"address": "eth.Address",
		"bool":    "bool",
	}[input]
}

func (t *TemplateEngine) MapEventDecoder(input string) string {
	return map[string]string{
		"uint256": "ReadBigInt",
		"int256":  "ReadBigInt",
		"uint112": "ReadBigInt",
		"address": "ReadAddress",
		"bool":    "ReadBool",
	}[input]
}

func (t *TemplateEngine) BackTick() string {
	return "`"
}

func (t *TemplateEngine) IterateSteps() (out []int) {
	for i := 0; i < t.HighestParallelStep; i++ {
		out = append(out, i+1)
	}
	return out
}

func (t *TemplateEngine) StringJoin(in []string) string {
	var values []string
	for _, s := range in {
		values = append(values, "'"+s+"'")
	}
	return strings.Join(values, ",")
}

func (t *TemplateEngine) readABI(yamlFilePath string, abi manifest.AbiRef) error {
	if _, found := t.Abis[abi.Name]; found {
		// skipping known abi
		return nil
	}
	abiFilePath, err := filepath.Abs(filepath.Join(filepath.Dir(yamlFilePath), abi.File.Path))
	if err != nil {
		return fmt.Errorf("unable to get ABI absolute filepath: %w", err)
	}

	abiData, err := ioutil.ReadFile(abiFilePath)
	if err != nil {
		return fmt.Errorf("unable to read ABI file: %w", err)
	}
	t.Abis[abi.Name] = string(abiData)
	return nil
}
