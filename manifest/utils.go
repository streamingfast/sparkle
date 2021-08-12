package manifest

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func DecodeYamlManifestFromFile(yamlFilePath string) (string, *SubgraphManifest, error) {
	yamlContent, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		return "", nil, fmt.Errorf("reading subgraph file %q: %w", yamlFilePath, err)
	}

	subgraphManifest, err := DecodeYamlManifest(string(yamlContent))
	if err != nil {
		return "", nil, fmt.Errorf("decoding subgraph file %q: %w", yamlFilePath, err)
	}

	return string(yamlContent), subgraphManifest, nil
}

func DecodeYamlManifest(manifestContent string) (*SubgraphManifest, error) {
	var subgraphManifest *SubgraphManifest
	if err := yaml.NewDecoder(bytes.NewReader([]byte(manifestContent))).Decode(&subgraphManifest); err != nil {
		return nil, fmt.Errorf("decoding manifest content %q: %w", manifestContent, err)
	}

	return subgraphManifest, nil
}
