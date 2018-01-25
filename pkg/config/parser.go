package config

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/flant/dapp/pkg/config/directive"
	"gopkg.in/flant/yaml.v2"
	"io/ioutil"
	"strings"
	"text/template"
)

func ParseDimgs(dappfilePath string) ([]*config.Dimg, []*config.DimgArtifact, error) {
	dappfileContent, err := parseDappfileYaml(dappfilePath)
	if err != nil {
		return nil, nil, err
	}
	return splitByDimgs(dappfileContent)
}

// FIXME: переделать на ParseFiles вместо Parse
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

func splitByDimgs(dappfileContent string) ([]*config.Dimg, []*config.DimgArtifact, error) {
	dimgsBase, err := splitByDimgsBase(dappfileContent)
	if err != nil {
		return nil, nil, err
	}

	var dimgs []*config.Dimg
	var artifacts []*config.DimgArtifact
	for _, dimgBase := range dimgsBase {
		dimg := &config.Dimg{DimgBase: *dimgBase} // TODO config.Dimg.New()
		if dimgBase.Type() == "dimg" {
			dimg.ValidateDirectives(artifacts)
			dimgs = append(dimgs, dimg)
		} else {
			dimgArtifact := &config.DimgArtifact{Dimg: *dimg} // TODO config.DimgArtifact.New()
			dimgArtifact.ValidateDirectives(artifacts)
			artifacts = append(artifacts, dimgArtifact)
		}
	}

	if len(dimgs) == 0 {
		return nil, nil, fmt.Errorf("ни одного dimgBase не объявлено!") // FIXME
	}

	return dimgs, artifacts, nil
}

func splitByDimgsBase(dappfileContent string) ([]*config.DimgBase, error) {
	scanner := bufio.NewScanner(strings.NewReader(dappfileContent))
	scanner.Split(splitYAMLDocument)

	var dimgsBase []*config.DimgBase
	for scanner.Scan() {
		dimg := &config.DimgBase{}
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
