package config

import (
	"bytes"
	"github.com/Masterminds/sprig"
	"text/template"
	"io/ioutil"
	"bufio"
	"strings"
	"gopkg.in/flant/yaml.v2"
	"fmt"
)

func LoadDappfile(dappfilePath string) (interface{}, error) {
	dappfileYaml, err := parseYaml(dappfilePath)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(dappfileYaml))
	scanner.Split(splitYAMLDocument)

	dimgs := []*Dimg{}
	for scanner.Scan() {
		config := &Dimg{}
		err = yaml.Unmarshal(scanner.Bytes(), &config)
		if err != nil {
			return nil, err
		}
		dimgs = append(dimgs, config)
	}

	for _, dimg := range dimgs {
		out, _ := yaml.Marshal(dimg)
		fmt.Printf("%+v\n\n", string(out))
	}

	return dappfileYaml, err
}

func parseYaml(dappfilePath string) (string, error) {
	data, err := ioutil.ReadFile(dappfilePath)
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
