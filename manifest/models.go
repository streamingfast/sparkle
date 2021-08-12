package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
)

type SubgraphManifest struct {
	SpecVersion string `yaml:"specVersion"`
	Description string `yaml:"description"`
	Repository  string `yaml:"repository"`
	Schema      struct {
		File *Link `yaml:"file"`
	} `yaml:"schema"`
	DataSources []ContractDataSource         `yaml:"dataSources"`
	Templates   []ContractDataSourceTemplate `yaml:"templates"`
}

func (m *SubgraphManifest) Network() string {
	var network string
	for _, datasource := range m.DataSources {
		network = datasource.Network
	}
	return network
}

func (m *SubgraphManifest) ReadSchema(yamlFilePath string) ([]byte, error) {
	schemaFilePath, err := filepath.Abs(filepath.Join(filepath.Dir(yamlFilePath), m.Schema.File.Path))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(schemaFilePath)
}

type ContractDataSource struct {
	// Kind    string          `yaml:"kind"`
	Network string          `yaml:"network"`
	Name    string          `yaml:"name"`
	Source  ContractSource  `yaml:"source"`
	Mapping ContractMapping `yaml:"mapping"`
}

type ContractDataSourceTemplate struct {
	// Kind    string                           `yaml:"kind"`
	Network string                           `yaml:"network"`
	Name    string                           `yaml:"name"`
	Source  ContractDataSourceTemplateSource `yaml:"source"`
	Mapping ContractMapping                  `yaml:"mapping"`
}

type ContractSource struct {
	Address    string `yaml:"address"`
	Abi        string `yaml:"abi"`
	StartBlock int    `yaml:"startBlock"`
}

type ContractDataSourceTemplateSource struct {
	Abi string `yaml:"abi"`
}

type AbiRef struct {
	Name string `yaml:"name"`
	File *Link  `yaml:"file"`
}
type ContractMapping struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	// Language      string   `yaml:"language"`
	// File *Link `yaml:"file"`
	// Entities      []string `yaml:"entities"`
	Abis          []AbiRef `yaml:"abis"`
	EventHandlers []struct {
		Event   string `yaml:"event"`
		Handler string `yaml:"handler"`
	} `yaml:"eventHandlers"`
}

type Link struct {
	IsLocalFile bool
	Path        string
}

func (l *Link) MarshalYAML() (interface{}, error) {
	if !l.IsLocalFile {
		return map[string]string{"/": l.Path}, nil
	}

	return l.Path, nil
}

func (l *Link) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var err error

	// check if value is a plain file path (in the case of user-defined subgraph yaml file)
	var file string
	if err = unmarshal(&file); err == nil {
		l.Path = file
		l.IsLocalFile = true
		return nil
	}

	// check if value is an ipfs file link (in the case of a compiled manifest)
	var ipfsMap map[string]string
	fileKey := "/"
	if err = unmarshal(&ipfsMap); err == nil {
		path, ok := ipfsMap[fileKey]
		if !ok {
			return fmt.Errorf("key `/` not found for link")
		}
		l.Path = path
		l.IsLocalFile = false
		return nil
	}

	return err
}
