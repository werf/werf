package config

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/flant/dapp/pkg/config/directive"
	raw "github.com/flant/dapp/pkg/config/raw"
	"gopkg.in/flant/yaml.v2"
	"io/ioutil"
	"strings"
	"text/template"
)

func ParseDimgs(dappfilePath string) ([]*config.Dimg, error) {
	dappfileContent, err := parseDappfileYaml(dappfilePath)
	if err != nil {
		return nil, err
	}
	return splitByDimgs(dappfileContent)
}

// TODO: переделать на ParseFiles вместо Parse
func parseDappfileYaml(dappfileContent string) (string, error) {
	data, err := ioutil.ReadFile(dappfileContent)
	if err != nil {
		return "", err
	}

	tmpl := template.New("dappfile")
	tmpl.Funcs(funcMap(tmpl))
	if _, err := tmpl.Parse(string(data)); err != nil {
		return "", err
	}
	return executeTemplate(tmpl, "dappfile", nil)
}

func funcMap(tmpl *template.Template) template.FuncMap {
	funcMap := sprig.TxtFuncMap()
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		return executeTemplate(tmpl, name, data)
	}
	return funcMap
}

func executeTemplate(tmpl *template.Template, name string, data interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func splitByDimgs(dappfileContent string) ([]*config.Dimg, error) {
	rawDimgs, err := splitByRawDimgs(dappfileContent)
	if err != nil {
		return nil, err
	}

	var dimgs []*config.Dimg
	var artifacts []*config.DimgArtifact

	for _, rawDimg := range rawDimgs {
		if rawDimg.Type() == "dimg" {
			if dimg, err := rawDimg.ToDimg(); err != nil {
				return nil, err
			} else {
				dimgs = append(dimgs, dimg)
			}
		} else {
			if dimgArtifact, err := rawDimg.ToDimgArtifact(); err != nil {
				return nil, err
			} else {
				artifacts = append(artifacts, dimgArtifact)
			}
		}
	}

	if len(dimgs) == 0 {
		return nil, fmt.Errorf("не описано ни одного dimg-а!") // FIXME
	}

	if err = associateArtifacts(dimgs, artifacts); err != nil {
		return nil, err
	}

	return dimgs, nil
}

func associateArtifacts(dimgs []*config.Dimg, artifacts []*config.DimgArtifact) error {
	for _, dimg := range dimgs {
		for _, importArtifact := range dimg.Import {
			if err := importArtifact.AssociateArtifact(artifacts); err != nil {
				return err
			}
		}
	}
	for _, dimg := range artifacts {
		for _, importArtifact := range dimg.Import {
			if err := importArtifact.AssociateArtifact(artifacts); err != nil {
				return err
			}
		}
	}
	return nil
}

func splitByRawDimgs(dappfileContent string) ([]*raw.Dimg, error) {
	scanner := bufio.NewScanner(strings.NewReader(dappfileContent))
	scanner.Split(splitYAMLDocument)

	var dimgsBase []*raw.Dimg
	for scanner.Scan() {
		dimg := &raw.Dimg{}
		err := yaml.Unmarshal(scanner.Bytes(), &dimg)
		if err != nil {
			return nil, err
		}
		dimgsBase = append(dimgsBase, dimg)
	}

	return dimgsBase, nil
}

func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	yamlSeparator := "\n---"

	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	sep := len([]byte(yamlSeparator))
	if i := bytes.Index(data, []byte(yamlSeparator)); i >= 0 {
		i += sep
		after := data[i:]
		if len(after) == 0 {
			if atEOF {
				return len(data), data[:len(data)-sep], nil
			}
			return 0, nil, nil
		}
		if j := bytes.IndexByte(after, '\n'); j >= 0 {
			return i + j + 1, data[0 : i-sep], nil
		}
		return 0, nil, nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
