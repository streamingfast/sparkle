package deployment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/streamingfast/sparkle/subgraph"

	manifestlib "github.com/streamingfast/sparkle/manifest"

	"gopkg.in/yaml.v3"
)

type IPFSNode struct {
	address string
	client  *http.Client
}

func NewIPFSNode(address string) *IPFSNode {
	return &IPFSNode{
		address: address,
		client:  http.DefaultClient,
	}
}

type IPFSAddResponse struct {
	Hash string
	File string
}

func (ipfs *IPFSNode) add(fileContent []byte) (*IPFSAddResponse, error) {
	endpoint := ipfs.address + "/api/v0/add?quiet=false"

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormField("file")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fw, bytes.NewReader(fileContent))
	if err != nil {
		return nil, err
	}
	defer w.Close()

	req, _ := http.NewRequest(http.MethodPost, endpoint, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := ipfs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling add api endpoint: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http error: %d (%s)", resp.StatusCode, resp.Status)
	}

	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decoding json response: %w", err)
	}

	return &IPFSAddResponse{
		File: fmt.Sprintf("/ipfs/%s", data["Name"].(string)),
		Hash: data["Hash"].(string),
	}, nil
}

func (ipfs *IPFSNode) UploadManifest(subgraphDef *subgraph.Definition) (string, []string, error) {
	var uploadedHashes []string
	manifest, err := manifestlib.DecodeYamlManifest(subgraphDef.Manifest)
	if err != nil {
		return "", uploadedHashes, fmt.Errorf("unable to decode manfiest: %w", err)
	}
	//upload schema
	if manifest.Schema.File.IsLocalFile {
		addResponse, err := ipfs.add([]byte(subgraphDef.GraphQLSchema))
		if err != nil {
			return "", uploadedHashes, err
		}
		uploadedHashes = append(uploadedHashes, addResponse.Hash)

		manifest.Schema.File = &manifestlib.Link{
			IsLocalFile: false,
			Path:        addResponse.File,
		}
	}

	//datasources
	for i, datasource := range manifest.DataSources {
		for j, abi := range datasource.Mapping.Abis {
			abiCnt, found := subgraphDef.Abis[abi.Name]
			if !found {
				return "", uploadedHashes, fmt.Errorf("manifest specfied abi %q is not loaded in the generate code", abi.Name)
			}

			addResponse, err := ipfs.add([]byte(abiCnt))
			if err != nil {
				return "", uploadedHashes, err
			}
			uploadedHashes = append(uploadedHashes, addResponse.Hash)

			manifest.DataSources[i].Mapping.Abis[j].File = &manifestlib.Link{
				IsLocalFile: false,
				Path:        addResponse.File,
			}
		}
	}

	//templates
	for i, template := range manifest.Templates {
		for j, abi := range template.Mapping.Abis {
			abiCnt, found := subgraphDef.Abis[abi.Name]
			if !found {
				return "", uploadedHashes, fmt.Errorf("manifest specfied abi %q is not loaded in the generate code", abi.Name)
			}

			addResponse, err := ipfs.add([]byte(abiCnt))
			if err != nil {
				return "", uploadedHashes, err
			}
			uploadedHashes = append(uploadedHashes, addResponse.Hash)

			manifest.Templates[i].Mapping.Abis[j].File = &manifestlib.Link{
				IsLocalFile: false,
				Path:        addResponse.File,
			}
		}
	}

	compiledManifest := bytes.NewBuffer(nil)
	err = yaml.NewEncoder(compiledManifest).Encode(manifest)
	if err != nil {
		return "", uploadedHashes, fmt.Errorf("compiling manifest: %w", err)
	}

	compiledManifestResponse, err := ipfs.add(compiledManifest.Bytes())
	if err != nil {
		return "", uploadedHashes, fmt.Errorf("uploading compiled manifest: %w", err)
	}
	uploadedHashes = append(uploadedHashes, compiledManifestResponse.Hash)

	return compiledManifestResponse.Hash, uploadedHashes, nil
}

func mustGetAbsoluteFileFromRelative(referenceFilePath, relativeFilePath string) string {
	absolutePath, err := filepath.Abs(filepath.Join(filepath.Dir(referenceFilePath), relativeFilePath))
	if err != nil {
		panic(err)
	}
	return absolutePath
}
