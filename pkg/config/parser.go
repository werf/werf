package config

import (
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
	"github.com/ghodss/yaml"

	"github.com/flant/dapp/pkg/logger"
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
	var dappfileRenderNameParts []string
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

func splitByDocs(dappfileRenderContent string, dappfileRenderPath string) ([]*doc, error) {
	var docs []*doc
	var line int
	for _, docContent := range splitContent([]byte(dappfileRenderContent)) {
		if !emptyDocContent(docContent) {
			docs = append(docs, &doc{
				Line:           line,
				Content:        docContent,
				RenderFilePath: dappfileRenderPath,
			})
		}

		contentLines := bytes.Split(docContent, []byte("\n"))
		if string(contentLines[len(contentLines)-1]) == "" {
			contentLines = contentLines[0 : len(contentLines)-1]
		}
		line += len(contentLines) + 1
	}

	return docs, nil
}

func parseDappfileYaml(dappfilePath string) (string, error) {
	data, err := ioutil.ReadFile(dappfilePath)
	if err != nil {
		return "", err
	}

	tmpl := template.New("dappfile")
	tmpl.Funcs(funcMap(tmpl))

	projectDir := filepath.Dir(dappfilePath)
	dappfilesDir := filepath.Join(projectDir, ".dappfiles")
	dappfilesTemplates, err := getDappfilesTemplates(dappfilesDir)
	if err != nil {
		return "", err
	}

	if len(dappfilesTemplates) != 0 {
		for _, templatePath := range dappfilesTemplates {
			templateName, err := filepath.Rel(dappfilesDir, templatePath)
			if err != nil {
				return "", err
			}

			extraTemplate := tmpl.New(templateName)

			var filePathData []byte
			if filePathData, err = ioutil.ReadFile(templatePath); err != nil {
				return "", err
			}

			if _, err := extraTemplate.Parse(string(filePathData)); err != nil {
				return "", err
			}
		}
	}

	if _, err := tmpl.Parse(string(data)); err != nil {
		return "", err
	}

	files := files{filepath.Dir(dappfilePath)}
	config, err := executeTemplate(tmpl, "dappfile", map[string]interface{}{"Files": files})

	return config, err
}

func getDappfilesTemplates(path string) ([]string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	var templates []string
	err := filepath.Walk(path, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		matched, err := filepath.Match("*.tmpl", fi.Name())
		if err != nil {
			return err
		}

		if matched {
			templates = append(templates, fp)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
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

type files struct {
	HomePath string
}

func (f files) Get(path string) string {
	filePath := filepath.Join(f.HomePath, path)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.LogWarningF("WARNING: Config: {{ .Files.Get '%s' }}: file '%s' not exist!\n", path, filePath)
		return ""
	}

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(b)
}

func splitContent(content []byte) (docsContents [][]byte) {
	const (
		stateLineBegin   = "stateLineBegin"
		stateRegularLine = "stateRegularLine"
		stateDocDash1    = "stateDocDash1"
		stateDocDash2    = "stateDocDash2"
		stateDocDash3    = "stateDocDash3"
		stateDocSpaces   = "stateDocSpaces"
		stateDocComment  = "stateDocComment"
	)

	state := stateLineBegin
	var docStartIndex, separatorLength int
	var docContent []byte
	var index int
	var ch byte
	for index, ch = range content {
		switch ch {
		case '-':
			switch state {
			case stateLineBegin:
				separatorLength = 1
				state = stateDocDash1
			case stateDocDash1, stateDocDash2:
				separatorLength += 1

				switch state {
				case stateDocDash1:
					state = stateDocDash2
				case stateDocDash2:
					state = stateDocDash3
				}
			default:
				state = stateRegularLine
			}
		case '\n':
			switch state {
			case stateDocDash3, stateDocSpaces, stateDocComment:
				if docStartIndex == index-separatorLength {
					docContent = []byte{}
				} else {
					docContent = content[docStartIndex : index-separatorLength]
				}
				docsContents = append(docsContents, docContent)
				docStartIndex = index + 1
			}
			separatorLength = 0
			state = stateLineBegin
		case ' ', '\r', '\t':
			switch state {
			case stateDocDash3, stateDocSpaces:
				separatorLength += 1
				state = stateDocSpaces
			case stateDocComment:
				separatorLength += 1
			default:
				state = stateRegularLine
			}
		case '#':
			switch state {
			case stateDocDash3, stateDocSpaces, stateDocComment:
				separatorLength += 1
				state = stateDocComment
			default:
				state = stateRegularLine
			}
		default:
			switch state {
			case stateDocComment:
				separatorLength += 1
			default:
				state = stateRegularLine
			}
		}
	}

	if docStartIndex != index+1 {
		switch state {
		case stateDocDash3, stateDocSpaces, stateDocComment:
			separatorLengthWithoutCursor := separatorLength - 1
			if docStartIndex == index-separatorLengthWithoutCursor {
				docContent = []byte{}
			} else {
				docContent = content[docStartIndex : index-separatorLengthWithoutCursor]
			}
		default:
			docContent = content[docStartIndex:]
		}
		docsContents = append(docsContents, docContent)
	}

	return docsContents
}

func emptyDocContent(content []byte) bool {
	const (
		stateRegular = 0
		stateComment = 1
	)

	state := stateRegular
	for _, ch := range content {
		switch ch {
		case '#':
			state = stateComment
		case '\n':
			state = stateRegular
		case ' ', '\r', '\t':
		default:
			if state == stateRegular {
				return false
			}
		}
	}
	return true
}

func splitByDimgs(docs []*doc, dappfileRenderContent string, dappfileRenderPath string) ([]*Dimg, error) {
	rawDimgs, err := splitByRawDimgs(docs)
	if err != nil {
		return nil, err
	}

	var dimgs []*Dimg
	var artifacts []*DimgArtifact

	for _, rawDimg := range rawDimgs {
		if rawDimg.dimgType() == "dimgs" {
			if sameDimgs, err := rawDimg.toDimgDirectives(); err != nil {
				return nil, err
			} else {
				dimgs = append(dimgs, sameDimgs...)
			}
		} else {
			if dimgArtifact, err := rawDimg.toDimgArtifactDirective(); err != nil {
				return nil, err
			} else {
				artifacts = append(artifacts, dimgArtifact)
			}
		}
	}

	if len(dimgs) == 0 {
		return nil, newConfigError(fmt.Sprintf("no dimgs defined, at least one dimg required!\n\n%s:\n\n```\n%s```\n", dappfileRenderPath, dappfileRenderContent))
	}

	if err = exportsAutoExcluding(dimgs, artifacts); err != nil {
		return nil, err
	}

	if err = validateDimgsNames(dimgs, artifacts); err != nil {
		return nil, err
	}

	if err = associateImportsArtifacts(dimgs, artifacts); err != nil {
		return nil, err
	}

	if err = associateDimgsAndArtifactsFrom(dimgs, artifacts); err != nil {
		return nil, err
	}

	return dimgs, nil
}

func exportsAutoExcluding(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	for _, dimg := range dimgs {
		if err := dimg.exportsAutoExcluding(); err != nil {
			return err
		}
	}

	for _, artifact := range artifacts {
		if err := artifact.exportsAutoExcluding(); err != nil {
			return err
		}
	}

	return nil
}

func validateDimgsNames(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	existByDimgName := map[string]bool{}

	dimgByName := map[string]*Dimg{}
	for _, dimg := range dimgs {
		name := dimg.Name

		if d, ok := dimgByName[name]; ok {
			return newConfigError(fmt.Sprintf("conflict between dimgs names!\n\n%s%s\n", dumpConfigDoc(d.raw.doc), dumpConfigDoc(dimg.raw.doc)))
		} else {
			dimgByName[name] = dimg
			existByDimgName[name] = true
		}
	}

	dimgArtifactByName := map[string]*DimgArtifact{}
	for _, artifact := range artifacts {
		name := artifact.Name

		if a, ok := dimgArtifactByName[name]; ok {
			return newConfigError(fmt.Sprintf("conflict between artifacts names!\n\n%s%s\n", dumpConfigDoc(a.raw.doc), dumpConfigDoc(artifact.raw.doc)))
		} else {
			dimgArtifactByName[name] = artifact
		}

		if exist, ok := existByDimgName[name]; ok && exist {
			d := dimgByName[name]

			return newConfigError(fmt.Sprintf("conflict between dimg and artifact names!\n\n%s%s\n", dumpConfigDoc(d.raw.doc), dumpConfigDoc(artifact.raw.doc)))
		} else {
			dimgArtifactByName[name] = artifact
		}
	}

	return nil
}

func associateImportsArtifacts(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	var artifactImports []*ArtifactImport

	for _, dimg := range dimgs {
		for _, relatedDimgInterface := range dimg.relatedDimgs() {
			switch relatedDimgInterface.(type) {
			case *Dimg:
				artifactImports = append(artifactImports, relatedDimgInterface.(*Dimg).Import...)
			case *DimgArtifact:
				artifactImports = append(artifactImports, relatedDimgInterface.(*DimgArtifact).Import...)
			}
		}
	}

	for _, artifactDimg := range artifacts {
		for _, relatedDimgInterface := range artifactDimg.relatedDimgs() {
			switch relatedDimgInterface.(type) {
			case *Dimg:
				artifactImports = append(artifactImports, relatedDimgInterface.(*Dimg).Import...)
			case *DimgArtifact:
				artifactImports = append(artifactImports, relatedDimgInterface.(*DimgArtifact).Import...)
			}
		}
	}

	for _, artifactImport := range artifactImports {
		if err := artifactImport.associateArtifact(artifacts); err != nil {
			return err
		}
	}

	return nil
}

func associateDimgsAndArtifactsFrom(dimgs []*Dimg, artifacts []*DimgArtifact) error {
	for _, dimg := range dimgs {
		if err := associateDimgFrom(dimg.lastLayerOrSelf(), dimgs, artifacts); err != nil {
			return err
		}
	}

	for _, dimg := range artifacts {
		if err := associateDimgFrom(dimg.lastLayerOrSelf(), dimgs, artifacts); err != nil {
			return err
		}
	}

	return nil
}

func associateDimgFrom(dimg DimgInterface, dimgs []*Dimg, artifacts []*DimgArtifact) error {
	switch dimg.(type) {
	case *Dimg:
		return dimg.(*Dimg).associateFrom(dimgs, artifacts)
	case *DimgArtifact:
		return dimg.(*DimgArtifact).associateFrom(dimgs, artifacts)
	default:
		panic("runtime error")
	}
}

func splitByRawDimgs(docs []*doc) ([]*rawDimg, error) {
	var rawDimgs []*rawDimg
	parentStack = util.NewStack()
	for _, doc := range docs {
		dimg := &rawDimg{doc: doc}
		err := yaml.Unmarshal(doc.Content, &dimg)
		if err != nil {
			return nil, newYamlUnmarshalError(err, doc)
		}
		rawDimgs = append(rawDimgs, dimg)
	}

	return rawDimgs, nil
}

func newYamlUnmarshalError(err error, doc *doc) error {
	switch err.(type) {
	case *configError:
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
		return newDetailedConfigError(message, nil, doc)
	}
}
