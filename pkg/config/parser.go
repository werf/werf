package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/flant/yaml.v2"

	"github.com/flant/dapp/pkg/util"
)

func ParseDimgs(dappfilePath string) ([]*Dimg, error) {
	dappfileRenderContent, err := parseDappfileYaml(dappfilePath)
	if err != nil {
		return nil, err
	}

	dappfileRenderPath, err := dumpDappfileRender(dappfilePath, dappfileRenderContent)
	if err != nil {
		return nil, err
	}

	docs, err := splitByDocs(dappfileRenderContent, dappfileRenderPath)
	if err != nil {
		return nil, err
	}

	dimgs, err := splitByDimgs(docs, dappfileRenderContent, dappfileRenderPath)
	if err != nil {
		return nil, err
	}

	return dimgs, nil
}

func dumpDappfileRender(dappfilePath string, dappfileRenderContent string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dappfileNameParts := strings.Split(path.Base(dappfilePath), ".")
	dappfileRenderNameParts := []string{}
	dappfileRenderNameParts = append(dappfileRenderNameParts, dappfileNameParts[0:len(dappfileNameParts)-1]...)
	dappfileRenderNameParts = append(dappfileRenderNameParts, "render", dappfileNameParts[len(dappfileNameParts)-1])
	dappfileRenderPath := path.Join(wd, fmt.Sprintf(".%s", strings.Join(dappfileRenderNameParts, ".")))

	dappfileRenderFile, err := os.OpenFile(dappfileRenderPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", err
	}
	dappfileRenderFile.Write([]byte(dappfileRenderContent))
	dappfileRenderFile.Close()

	return dappfileRenderPath, nil
}

func splitByDocs(dappfileRenderContent string, dappfileRenderPath string) ([]*Doc, error) {
	scanner := bufio.NewScanner(strings.NewReader(dappfileRenderContent))
	scanner.Split(splitYAMLDocument)

	var docs []*Doc
	var line int
	for scanner.Scan() {
		content := make([]byte, len(scanner.Bytes()))
		copy(content, scanner.Bytes())

		if strings.TrimSpace(string(content)) != "" {
			docs = append(docs, &Doc{
				Line:           line,
				Content:        content,
				RenderFilePath: dappfileRenderPath,
			})
		}

		contentLines := bytes.Split(content, []byte("\n"))
		if string(contentLines[len(contentLines)-1]) == "" {
			contentLines = contentLines[0 : len(contentLines)-1]
		}
		line += len(contentLines) + 1
	}

	return docs, nil
}

// TODO: переделать на ParseFiles вместо Parse
func parseDappfileYaml(dappfilePath string) (string, error) {
	data, err := ioutil.ReadFile(dappfilePath)
	if err != nil {
		return "", err
	}

	tmpl := template.New("dappfile")
	tmpl.Funcs(funcMap(tmpl))
	if _, err := tmpl.Parse(string(data)); err != nil {
		return "", err
	}
	return executeTemplate(tmpl, "dappfile", map[string]interface{}{"Files": Files{filepath.Dir(dappfilePath)}})
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

type Files struct {
	HomePath string
}

func (f Files) Get(path string) string {
	b, err := ioutil.ReadFile(filepath.Join(f.HomePath, path))
	if err != nil {
		return ""
	}
	return string(b)
}

func splitYAMLDocument(data []byte, atEOF bool) (advance int, token []byte, err error) {
	yamlSeparator := "---"

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

func splitByDimgs(docs []*Doc, dappfileRenderContent string, dappfileRenderPath string) ([]*Dimg, error) {
	rawDimgs, err := splitByRawDimgs(docs)
	if err != nil {
		return nil, err
	}

	var dimgs []*Dimg
	var artifacts []*DimgArtifact

	for _, rawDimg := range rawDimgs {
		if rawDimg.Type() == "dimg" {
			if dimg, err := rawDimg.ToDirective(); err != nil {
				return nil, err
			} else {
				dimgs = append(dimgs, dimg)
			}
		} else {
			if dimgArtifact, err := rawDimg.ToArtifactDirective(); err != nil {
				return nil, err
			} else {
				artifacts = append(artifacts, dimgArtifact)
			}
		}
	}

	if len(dimgs) == 0 {
		return nil, NewConfigError(fmt.Sprintf("No dimgs defined, at least one dimg required!\n\n%s:\n\n```\n%s```\n", dappfileRenderPath, dappfileRenderContent))
	}

	if err = associateArtifacts(dimgs, artifacts); err != nil {
		return nil, err
	}

	return dimgs, nil
}

func associateArtifacts(dimgs []*Dimg, artifacts []*DimgArtifact) error {
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

func splitByRawDimgs(docs []*Doc) ([]*RawDimg, error) {
	var rawDimgs []*RawDimg
	ParentStack = util.NewStack()
	for _, doc := range docs {
		dimg := &RawDimg{Doc: doc}
		err := yaml.Unmarshal(doc.Content, &dimg)
		if err != nil {
			return nil, newYamlUnmarshalError(err, doc)
		}
		rawDimgs = append(rawDimgs, dimg)
	}

	return rawDimgs, nil
}

func newYamlUnmarshalError(err error, doc *Doc) error {
	switch err.(type) {
	case *ConfigError:
		return err
	default:
		message := err.Error()
		reg, err := regexp.Compile("line ([0-9]+)")
		if err != nil {
			return err
		}

		res := reg.FindStringSubmatch(message)

		if len(res) == 2 {
			line, err := strconv.Atoi(res[1])
			if err != nil {
				return err
			}

			message = reg.ReplaceAllString(message, fmt.Sprintf("line %d", line+doc.Line))
		}
		return NewDetailedConfigError(message, nil, doc)
	}
}
