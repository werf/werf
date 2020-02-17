package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/yaml.v2"

	"github.com/flant/logboek"

	"github.com/flant/werf/pkg/git_repo"
	"github.com/flant/werf/pkg/logging"
	"github.com/flant/werf/pkg/slug"
	"github.com/flant/werf/pkg/tmp_manager"
	"github.com/flant/werf/pkg/util"
)

func RenderWerfConfig(werfConfigPath string, imagesToProcess []string) error {
	werfConfig, err := GetWerfConfig(werfConfigPath, false)
	if err != nil {
		return err
	}

	if len(imagesToProcess) == 0 {
		werfConfigRenderContent, err := parseWerfConfigYaml(werfConfigPath)
		if err != nil {
			return fmt.Errorf("cannot parse config: %s", err)
		}

		fmt.Print(werfConfigRenderContent)
	} else {
		var imageDocs []string

		for _, imageToProcess := range imagesToProcess {
			if !werfConfig.HasImageOrArtifact(imageToProcess) {
				return fmt.Errorf("specified image %s is not defined in werf.yaml", logging.ImageLogName(imageToProcess, false))
			} else {
				if i := werfConfig.GetArtifact(imageToProcess); i != nil {
					imageDocs = append(imageDocs, string(i.raw.doc.Content))
				} else if i := werfConfig.GetStapelImage(imageToProcess); i != nil {
					imageDocs = append(imageDocs, string(i.raw.doc.Content))
				} else if i := werfConfig.GetDockerfileImage(imageToProcess); i != nil {
					imageDocs = append(imageDocs, string(i.raw.doc.Content))
				}
			}
		}

		fmt.Print(strings.Join(imageDocs, "---\n"))
	}

	return nil
}

func GetWerfConfig(werfConfigPath string, logRenderedFilePath bool) (*WerfConfig, error) {
	werfConfigRenderContent, err := parseWerfConfigYaml(werfConfigPath)
	if err != nil {
		return nil, fmt.Errorf("cannot parse config: %s", err)
	}

	werfConfigRenderPath, err := tmp_manager.CreateWerfConfigRender()
	if err != nil {
		return nil, err
	}

	if logRenderedFilePath {
		logboek.LogF("Using werf config render file: %s\n", werfConfigRenderPath)
	}

	err = writeWerfConfigRender(werfConfigRenderContent, werfConfigRenderPath)
	if err != nil {
		return nil, fmt.Errorf("unable to write rendered config to %s: %s", werfConfigRenderPath, err)
	}

	docs, err := splitByDocs(werfConfigRenderContent, werfConfigRenderPath)
	if err != nil {
		return nil, err
	}

	meta, rawStapelImages, rawImagesFromDockerfile, err := splitByMetaAndRawImages(docs)
	if err != nil {
		return nil, err
	}

	if meta == nil {
		defaultProjectName, err := GetProjectName(filepath.Dir(werfConfigPath))
		if err != nil {
			return nil, fmt.Errorf("failed to get default project name: %s", err)
		}

		format := "meta config section (part of YAML stream separated by three hyphens, https://yaml.org/spec/1.2/spec.html#id2800132) is not defined: add following example config section with required fields, e.g:\n\n" +
			"```\n" +
			"configVersion: 1\n" +
			"project: %s\n" +
			"---\n" +
			"```\n\n" +
			"##############################################################################################################################\n" +
			"###           WARNING! Project name cannot be changed later without rebuilding and redeploying your application!           ###\n" +
			"###       Project name should be unique within group of projects that shares build hosts and deployed into the same        ###\n" +
			"###                    Kubernetes clusters (i.e. unique across all groups within the same gitlab).                         ###\n" +
			"### Read more about meta config section: https://werf.io/documentation/configuration/introduction.html#meta-config-section ###\n" +
			"##############################################################################################################################"

		return nil, fmt.Errorf(format, defaultProjectName)
	}

	werfConfig, err := prepareWerfConfig(rawStapelImages, rawImagesFromDockerfile, meta)
	if err != nil {
		return nil, err
	}

	return werfConfig, nil
}

func GetProjectName(projectDir string) (string, error) {
	name := filepath.Base(projectDir)

	if exist, err := util.DirExists(filepath.Join(projectDir, ".git")); err != nil {
		return "", err
	} else if exist {
		remoteOriginUrl, err := gitOwnRepoOriginUrl(projectDir)
		if err != nil {
			return "", err
		}

		if remoteOriginUrl != "" {
			ep, err := transport.NewEndpoint(remoteOriginUrl)
			if err != nil {
				return "", fmt.Errorf("bad url '%s': %s", remoteOriginUrl, err)
			}

			gitName := strings.TrimSuffix(ep.Path, ".git")

			return slug.Project(gitName), nil
		}
	}

	return slug.Project(name), nil
}

func gitOwnRepoOriginUrl(projectDir string) (string, error) {
	localGitRepo := &git_repo.Local{
		Path:   projectDir,
		GitDir: filepath.Join(projectDir, ".git"),
	}

	remoteOriginUrl, err := localGitRepo.RemoteOriginUrl()
	if err != nil {
		return "", nil
	}

	return remoteOriginUrl, nil
}

func writeWerfConfigRender(werfConfigRenderContent string, werfConfigRenderPath string) error {
	werfConfigRenderFile, err := os.OpenFile(werfConfigRenderPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = werfConfigRenderFile.Write([]byte(werfConfigRenderContent))
	if err != nil {
		return err
	}

	err = werfConfigRenderFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func splitByDocs(werfConfigRenderContent string, werfConfigRenderPath string) ([]*doc, error) {
	var docs []*doc
	var line int
	for _, docContent := range splitContent([]byte(werfConfigRenderContent)) {
		if !emptyDocContent(docContent) {
			docs = append(docs, &doc{
				Line:           line,
				Content:        docContent,
				RenderFilePath: werfConfigRenderPath,
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

func parseWerfConfigYaml(werfConfigPath string) (string, error) {
	data, err := ioutil.ReadFile(werfConfigPath)
	if err != nil {
		return "", err
	}

	tmpl := template.New("werfConfig")
	tmpl.Funcs(funcMap(tmpl))

	projectDir := filepath.Dir(werfConfigPath)
	werfConfigsDir := filepath.Join(projectDir, ".werf")
	werfConfigsTemplates, err := getWerfConfigsTemplates(werfConfigsDir)
	if err != nil {
		return "", err
	}

	if len(werfConfigsTemplates) != 0 {
		for _, templatePath := range werfConfigsTemplates {
			templateName, err := filepath.Rel(werfConfigsDir, templatePath)
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

	files := files{filepath.Dir(werfConfigPath)}
	config, err := executeTemplate(tmpl, "werfConfig", map[string]interface{}{"Files": files})

	return config, err
}

func getWerfConfigsTemplates(path string) ([]string, error) {
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
		logboek.LogErrorF("WARNING: Config: {{ .Files.Get '%s' }}: file '%s' not exist!\n", path, filePath)
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

func prepareWerfConfig(rawImages []*rawStapelImage, rawImagesFromDockerfile []*rawImageFromDockerfile, meta *Meta) (*WerfConfig, error) {
	var stapelImages []*StapelImage
	var imagesFromDockerfile []*ImageFromDockerfile
	var artifacts []*StapelImageArtifact

	for _, rawImageFromDockerfile := range rawImagesFromDockerfile {
		if sameImages, err := rawImageFromDockerfile.toImageFromDockerfileDirectives(); err != nil {
			return nil, err
		} else {
			imagesFromDockerfile = append(imagesFromDockerfile, sameImages...)
		}
	}

	for _, rawImage := range rawImages {
		if rawImage.stapelImageType() == "images" {
			if sameImages, err := rawImage.toStapelImageDirectives(); err != nil {
				return nil, err
			} else {
				stapelImages = append(stapelImages, sameImages...)
			}
		} else {
			if imageArtifacts, err := rawImage.toStapelImageArtifactDirectives(); err != nil {
				return nil, err
			} else {
				artifacts = append(artifacts, imageArtifacts...)
			}
		}
	}

	werfConfig := &WerfConfig{
		Meta:                 meta,
		StapelImages:         stapelImages,
		ImagesFromDockerfile: imagesFromDockerfile,
		Artifacts:            artifacts,
	}

	if err := werfConfig.validateImagesNames(); err != nil {
		return nil, err
	}

	if err := werfConfig.validateImagesFrom(); err != nil {
		return nil, err
	}

	if err := werfConfig.associateImportsArtifacts(); err != nil {
		return nil, err
	}

	if err := werfConfig.exportsAutoExcluding(); err != nil {
		return nil, err
	}

	if err := werfConfig.validateInfiniteLoopBetweenRelatedImages(); err != nil {
		return nil, err
	}

	return werfConfig, nil
}

func splitByMetaAndRawImages(docs []*doc) (*Meta, []*rawStapelImage, []*rawImageFromDockerfile, error) {
	var rawStapelImages []*rawStapelImage
	var rawImagesFromDockerfile []*rawImageFromDockerfile
	var resultMeta *Meta

	parentStack = util.NewStack()
	for _, doc := range docs {
		var raw map[string]interface{}
		err := yaml.UnmarshalStrict(doc.Content, &raw)
		if err != nil {
			return nil, nil, nil, newYamlUnmarshalError(err, doc)
		}

		if isMetaDoc(raw) {
			if resultMeta != nil {
				return nil, nil, nil, newYamlUnmarshalError(errors.New("duplicate meta config section definition"), doc)
			}

			rawMeta := &rawMeta{doc: doc}
			err := yaml.UnmarshalStrict(doc.Content, &rawMeta)
			if err != nil {
				return nil, nil, nil, newYamlUnmarshalError(err, doc)
			}

			resultMeta = rawMeta.toMeta()
		} else if isImageFromDockerfileDoc(raw) {
			imageFromDockerfile := &rawImageFromDockerfile{doc: doc}
			err := yaml.UnmarshalStrict(doc.Content, &imageFromDockerfile)
			if err != nil {
				return nil, nil, nil, newYamlUnmarshalError(err, doc)
			}

			rawImagesFromDockerfile = append(rawImagesFromDockerfile, imageFromDockerfile)
		} else if isImageDoc(raw) {
			image := &rawStapelImage{doc: doc}
			err := yaml.UnmarshalStrict(doc.Content, &image)
			if err != nil {
				return nil, nil, nil, newYamlUnmarshalError(err, doc)
			}

			rawStapelImages = append(rawStapelImages, image)
		} else {
			return nil, nil, nil, newYamlUnmarshalError(errors.New("cannot recognize type of config section (part of YAML stream separated by three hyphens, https://yaml.org/spec/1.2/spec.html#id2800132):\n * 'configVersion' required for meta config section;\n * 'image' required for the image config sections;\n * 'artifact' required for the artifact config sections;"), doc)
		}
	}

	return resultMeta, rawStapelImages, rawImagesFromDockerfile, nil
}

func isMetaDoc(h map[string]interface{}) bool {
	if _, ok := h["configVersion"]; ok {
		return true
	}

	return false
}

func isImageDoc(h map[string]interface{}) bool {
	if _, ok := h["image"]; ok {
		return true
	} else if _, ok := h["artifact"]; ok {
		return true
	}

	return false
}

func isImageFromDockerfileDoc(h map[string]interface{}) bool {
	if _, ok := h["dockerfile"]; ok {
		return true
	}

	return false
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
