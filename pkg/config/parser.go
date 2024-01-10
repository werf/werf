package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"gopkg.in/yaml.v2"

	"github.com/werf/logboek"
	"github.com/werf/werf/pkg/giterminism_manager"
	"github.com/werf/werf/pkg/slug"
	"github.com/werf/werf/pkg/tmp_manager"
	"github.com/werf/werf/pkg/util"
)

type WerfConfigOptions struct {
	LogRenderedFilePath bool
	Env                 string
}

func RenderWerfConfig(ctx context.Context, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath string, imagesToProcess []string, giterminismManager giterminism_manager.Interface, opts WerfConfigOptions) error {
	_, werfConfig, err := GetWerfConfig(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, giterminismManager, opts)
	if err != nil {
		return err
	}

	if len(imagesToProcess) == 0 {
		_, werfConfigRenderContent, err := renderWerfConfigYaml(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, giterminismManager, opts.Env)
		if err != nil {
			return err
		}

		fmt.Print(werfConfigRenderContent)
		return nil
	}

	if err := werfConfig.CheckThatImagesExist(imagesToProcess); err != nil {
		return err
	}

	var imageDocs []string
	for _, imageToProcess := range imagesToProcess {
		if i := werfConfig.GetArtifact(imageToProcess); i != nil {
			imageDocs = append(imageDocs, string(i.raw.doc.Content))
		} else if i := werfConfig.GetStapelImage(imageToProcess); i != nil {
			imageDocs = append(imageDocs, string(i.raw.doc.Content))
		} else if i := werfConfig.GetDockerfileImage(imageToProcess); i != nil {
			imageDocs = append(imageDocs, string(i.raw.doc.Content))
		}
	}

	fmt.Print(strings.Join(imageDocs, "---\n"))
	return nil
}

func GetWerfConfig(ctx context.Context, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath string, giterminismManager giterminism_manager.Interface, opts WerfConfigOptions) (string, *WerfConfig, error) {
	werfConfigPath, werfConfigRenderContent, err := renderWerfConfigYaml(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, giterminismManager, opts.Env)
	if err != nil {
		return "", nil, err
	}

	werfConfigRenderPath, err := tmp_manager.CreateWerfConfigRender(ctx)
	if err != nil {
		return "", nil, err
	}

	if opts.LogRenderedFilePath {
		logboek.Context(ctx).LogF("Using werf config render file: %s\n", werfConfigRenderPath)
	}

	err = writeWerfConfigRender(werfConfigRenderContent, werfConfigRenderPath)
	if err != nil {
		return "", nil, fmt.Errorf("unable to write rendered config to %s: %w", werfConfigRenderPath, err)
	}

	docs, err := splitByDocs(werfConfigRenderContent, werfConfigRenderPath)
	if err != nil {
		return "", nil, err
	}

	meta, rawStapelImages, rawImagesFromDockerfile, err := splitByMetaAndRawImages(docs)
	if err != nil {
		return "", nil, err
	}

	if meta == nil {
		defaultProjectName, err := GetDefaultProjectName(ctx, giterminismManager)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get default project name: %w", err)
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
			"###              Read more about meta config section: https://werf.io/documentation/reference/werf_yaml.html               ###\n" +
			"##############################################################################################################################"

		return "", nil, fmt.Errorf(format, defaultProjectName)
	}

	werfConfig, err := prepareWerfConfig(giterminismManager, rawStapelImages, rawImagesFromDockerfile, meta)
	if err != nil {
		return "", nil, err
	}

	return werfConfigPath, werfConfig, nil
}

func GetDefaultProjectName(ctx context.Context, giterminismManager giterminism_manager.Interface) (string, error) {
	if remoteOriginUrl, err := giterminismManager.LocalGitRepo().RemoteOriginUrl(ctx); err != nil {
		return "", nil
	} else if remoteOriginUrl != "" {
		ep, err := transport.NewEndpoint(remoteOriginUrl)
		if err != nil {
			return "", fmt.Errorf("bad url %q: %w", remoteOriginUrl, err)
		}

		gitName := strings.TrimSuffix(ep.Path, ".git")

		return slug.Project(gitName), nil
	}

	name := filepath.Base(giterminismManager.ProjectDir())
	return slug.Project(name), nil
}

func writeWerfConfigRender(werfConfigRenderContent, werfConfigRenderPath string) error {
	werfConfigRenderFile, err := os.OpenFile(werfConfigRenderPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
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

func splitByDocs(werfConfigRenderContent, werfConfigRenderPath string) ([]*doc, error) {
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

func renderWerfConfigYaml(ctx context.Context, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath string, giterminismManager giterminism_manager.Interface, env string) (string, string, error) {
	tmpl := template.New("werfConfig")
	tmpl.Funcs(funcMap(tmpl, giterminismManager))

	if err := parseWerfConfigTemplatesDir(ctx, tmpl, giterminismManager, customWerfConfigTemplatesDirRelPath); err != nil {
		return "", "", err
	}

	configPath, err := parseWerfConfig(ctx, tmpl, giterminismManager, customWerfConfigRelPath)
	if err != nil {
		return "", "", err
	}

	templateData := make(map[string]interface{})
	templateData["Files"] = files{
		ctx:                ctx,
		giterminismManager: giterminismManager,
	}
	templateData["Env"] = env

	headHash, err := giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return "", "", fmt.Errorf("unable to get HEAD commit hash: %w", err)
	}

	headTime, err := giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
	if err != nil {
		return "", "", fmt.Errorf("unable to get HEAD commit time: %w", err)
	}

	templateData["Commit"] = map[string]interface{}{
		"Hash": headHash,
		"Date": map[string]string{
			"Human": headTime.String(),
			"Unix":  strconv.FormatInt(headTime.Unix(), 10),
		},
	}

	config, err := executeTemplate(tmpl, "werfConfig", templateData)

	return configPath, config, err
}

func parseWerfConfig(ctx context.Context, tmpl *template.Template, giterminismManager giterminism_manager.Interface, relWerfConfigPath string) (string, error) {
	configPath, configData, err := giterminismManager.FileReader().ReadConfig(ctx, relWerfConfigPath)
	if err != nil {
		return "", err
	}

	if _, err := tmpl.Parse(string(configData)); err != nil {
		return "", err
	}

	return configPath, nil
}

func parseWerfConfigTemplatesDir(ctx context.Context, tmpl *template.Template, giterminismManager giterminism_manager.Interface, customWerfConfigTemplatesDirRelPath string) error {
	return giterminismManager.FileReader().ReadConfigTemplateFiles(ctx, customWerfConfigTemplatesDirRelPath, func(templatePathInsideDir string, data []byte, err error) error {
		if err != nil {
			return err
		}

		templateName := filepath.ToSlash(templatePathInsideDir)
		if err := addTemplate(tmpl, templateName, string(data)); err != nil {
			return err
		}

		return nil
	})
}

func addTemplate(tmpl *template.Template, templateName, templateContent string) error {
	extraTemplate := tmpl.New(templateName)
	_, err := extraTemplate.Parse(templateContent)
	return err
}

func funcMap(tmpl *template.Template, giterminismManager giterminism_manager.Interface) template.FuncMap {
	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "expandenv")

	funcMap["fromYaml"] = func(str string) (map[string]interface{}, error) {
		m := map[string]interface{}{}

		if err := yaml.Unmarshal([]byte(str), &m); err != nil {
			return nil, err
		}

		return m, nil
	}
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		return executeTemplate(tmpl, name, data)
	}
	funcMap["tpl"] = func(templateContent string, data interface{}) (string, error) {
		templateName := util.GenerateConsistentRandomString(10)
		if err := addTemplate(tmpl, templateName, templateContent); err != nil {
			return "", err
		}

		return executeTemplate(tmpl, templateName, data)
	}

	funcMap["env"] = func(value interface{}, args ...string) (string, error) {
		if len(args) > 1 {
			return "", fmt.Errorf("more than 1 optional argument prohibited")
		}

		envVarName := fmt.Sprint(value)
		if err := giterminismManager.Inspector().InspectConfigGoTemplateRenderingEnv(context.Background(), envVarName); err != nil {
			return "", err
		}

		var fallbackValue *string
		if len(args) == 1 {
			fallbackValue = &args[0]
		}

		giterministic := !giterminismManager.LooseGiterminism()
		envVarValue, envVarFound := os.LookupEnv(envVarName)
		if !envVarFound {
			if fallbackValue != nil {
				return *fallbackValue, nil
			} else if giterministic {
				return "", fmt.Errorf("the environment variable %q must be set or default must be provided", envVarName)
			} else {
				return "", nil
			}
		}

		if envVarValue == "" && fallbackValue != nil {
			return *fallbackValue, nil
		}

		return envVarValue, nil
	}

	funcMap["required"] = func(msg string, val interface{}) (interface{}, error) {
		if val == nil {
			return val, errors.New(msg)
		} else if _, ok := val.(string); ok {
			if val == "" {
				return val, errors.New(msg)
			}
		}
		return val, nil
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
	ctx                context.Context
	giterminismManager giterminism_manager.Interface
}

func (f files) Get(relPath string) string {
	if res, err := f.giterminismManager.FileReader().ConfigGoTemplateFilesGet(f.ctx, relPath); err != nil {
		panic(err.Error())
	} else {
		return string(res)
	}
}

func (f files) doGlob(ctx context.Context, pattern string) (map[string]interface{}, error) {
	res, err := f.giterminismManager.FileReader().ConfigGoTemplateFilesGlob(ctx, pattern)
	if err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, fmt.Errorf("{{ .Files.Glob %q }}: no matches found", pattern)
	}

	return res, nil
}

func (f files) Glob(pattern string) map[string]interface{} {
	if res, err := f.doGlob(f.ctx, pattern); err != nil {
		panic(err.Error())
	} else {
		return res
	}
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

func prepareWerfConfig(giterminismManager giterminism_manager.Interface, rawImages []*rawStapelImage, rawImagesFromDockerfile []*rawImageFromDockerfile, meta *Meta) (*WerfConfig, error) {
	var stapelImages []*StapelImage
	var imagesFromDockerfile []*ImageFromDockerfile
	var artifacts []*StapelImageArtifact

	for _, rawImageFromDockerfile := range rawImagesFromDockerfile {
		if sameImages, err := rawImageFromDockerfile.toImageFromDockerfileDirectives(giterminismManager); err != nil {
			return nil, err
		} else {
			imagesFromDockerfile = append(imagesFromDockerfile, sameImages...)
		}
	}

	for _, rawImage := range rawImages {
		if rawImage.stapelImageType() == "images" {
			if sameImages, err := rawImage.toStapelImageDirectives(giterminismManager); err != nil {
				return nil, err
			} else {
				stapelImages = append(stapelImages, sameImages...)
			}
		} else {
			if imageArtifact, err := rawImage.toStapelImageArtifactDirectives(giterminismManager); err != nil {
				return nil, err
			} else {
				artifacts = append(artifacts, imageArtifact)
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

	if err := werfConfig.validateDependencies(); err != nil {
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

		switch {
		case isMetaDoc(raw):
			if resultMeta != nil {
				return nil, nil, nil, newYamlUnmarshalError(errors.New("duplicate meta config section definition"), doc)
			}
			rawMeta := &rawMeta{doc: doc}
			err := yaml.UnmarshalStrict(doc.Content, &rawMeta)
			if err != nil {
				return nil, nil, nil, newYamlUnmarshalError(err, doc)
			}
			resultMeta = rawMeta.toMeta()
		case isImageFromDockerfileDoc(raw):
			imageFromDockerfile := &rawImageFromDockerfile{doc: doc}
			err := yaml.UnmarshalStrict(doc.Content, &imageFromDockerfile)
			if err != nil {
				return nil, nil, nil, newYamlUnmarshalError(err, doc)
			}
			rawImagesFromDockerfile = append(rawImagesFromDockerfile, imageFromDockerfile)
		case isImageDoc(raw):
			image := &rawStapelImage{doc: doc}
			err := yaml.UnmarshalStrict(doc.Content, &image)
			if err != nil {
				return nil, nil, nil, newYamlUnmarshalError(err, doc)
			}
			rawStapelImages = append(rawStapelImages, image)
		default:
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
		reg := regexp.MustCompile("line ([0-9]+)")
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
