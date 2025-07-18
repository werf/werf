package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/adler32"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"gopkg.in/yaml.v2"

	"github.com/werf/3p-helm/pkg/engine"
	"github.com/werf/common-go/pkg/util"
	"github.com/werf/logboek"
	"github.com/werf/werf/v2/pkg/giterminism_manager"
	"github.com/werf/werf/v2/pkg/slug"
	"github.com/werf/werf/v2/pkg/tmp_manager"
)

type WerfConfigOptions struct {
	LogRenderedFilePath bool
	Env                 string
	DebugTemplates      bool
}

func RenderWerfConfig(ctx context.Context, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath string, imageNameList []string, giterminismManager giterminism_manager.Interface, opts WerfConfigOptions) error {
	_, werfConfig, err := GetWerfConfig(ctx, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath, giterminismManager, opts)
	if err != nil {
		return err
	}

	if len(imageNameList) == 0 {
		return renderWerfConfig(ctx, renderWerfConfigYamlOpts{
			customWerfConfigRelPath:             customWerfConfigRelPath,
			customWerfConfigTemplatesDirRelPath: customWerfConfigTemplatesDirRelPath,
			giterminismManager:                  giterminismManager,
			env:                                 opts.Env,
			includesConfigRelPath:               "", // default
			debugTemplates:                      opts.DebugTemplates,
		})
	}
	return renderSpecificImages(werfConfig, imageNameList)
}

func renderWerfConfig(ctx context.Context, opts renderWerfConfigYamlOpts) error {
	_, werfConfigRenderContent, err := renderWerfConfigYaml(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Print(werfConfigRenderContent)
	return nil
}

func renderSpecificImages(werfConfig *WerfConfig, imageNameList []string) error {
	imagesToProcess, err := NewImagesToProcess(werfConfig, imageNameList, false, false)
	if err != nil {
		return err
	}

	var docs []string
	for _, imageConfig := range werfConfig.getSpecificImages(imagesToProcess) {
		docs = append(docs, string(imageConfig.rawDoc().Content))
	}

	fmt.Print(strings.Join(docs, "---\n"))
	return nil
}

func GetWerfConfig(ctx context.Context, customWerfConfigRelPath, customWerfConfigTemplatesDirRelPath string, giterminismManager giterminism_manager.Interface, opts WerfConfigOptions) (string, *WerfConfig, error) {
	var path string
	var config *WerfConfig
	err := logboek.Context(ctx).Info().LogProcess("Render werf config").DoError(func() error {
		werfConfigPath, werfConfigRenderContent, err := renderWerfConfigYaml(ctx, renderWerfConfigYamlOpts{
			customWerfConfigRelPath:             customWerfConfigRelPath,
			customWerfConfigTemplatesDirRelPath: customWerfConfigTemplatesDirRelPath,
			giterminismManager:                  giterminismManager,
			env:                                 opts.Env,
			debugTemplates:                      opts.DebugTemplates,
		})
		if err != nil {
			return fmt.Errorf("unable to render werf config: %w", err)
		}

		werfConfigRenderPath, err := tmp_manager.CreateWerfConfigRender(ctx)
		if err != nil {
			return err
		}

		if opts.LogRenderedFilePath {
			logboek.Context(ctx).LogF("Using werf config render file: %s\n", werfConfigRenderPath)
		}

		err = writeWerfConfigRender(werfConfigRenderContent, werfConfigRenderPath)
		if err != nil {
			return fmt.Errorf("unable to write rendered config to %s: %w", werfConfigRenderPath, err)
		}

		docs, err := splitByDocs(werfConfigRenderContent, werfConfigRenderPath)
		if err != nil {
			return err
		}

		meta, rawStapelImages, rawImagesFromDockerfile, err := splitByMetaAndRawImages(docs)
		if err != nil {
			return err
		}

		imgPlatformValidator := newImagePlatformValidator()
		if err = imgPlatformValidator.Validate(rawStapelImages, rawImagesFromDockerfile); err != nil {
			return fmt.Errorf("invalid image platform cross-references: %w", err)
		}

		if meta == nil {
			defaultProjectName, err := GetDefaultProjectName(ctx, giterminismManager)
			if err != nil {
				return fmt.Errorf("failed to get default project name: %w", err)
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
				"##############################################################################################################################"

			return fmt.Errorf(format, defaultProjectName)
		}

		werfConfig, err := prepareWerfConfig(giterminismManager, rawStapelImages, rawImagesFromDockerfile, meta)
		if err != nil {
			return err
		}

		path = werfConfigPath
		config = werfConfig

		return nil
	})
	if err != nil {
		return "", nil, err
	}

	return path, config, nil
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

type renderWerfConfigYamlOpts struct {
	customWerfConfigRelPath             string
	customWerfConfigTemplatesDirRelPath string
	giterminismManager                  giterminism_manager.Interface
	env                                 string
	includesConfigRelPath               string
	debugTemplates                      bool
}

func renderWerfConfigYaml(ctx context.Context, opts renderWerfConfigYamlOpts) (string, string, error) {
	tmpl := template.New("werfConfig")
	tmpl.Funcs(funcMap(ctx, tmpl, opts.giterminismManager, opts.debugTemplates))

	err := parseWerfConfigTemplatesDir(ctx, parseWerfConfigTemplatesDirOpts{
		tmpl:                                tmpl,
		customWerfConfigTemplatesDirRelPath: opts.customWerfConfigTemplatesDirRelPath,
		giterminismManager:                  opts.giterminismManager.(*giterminism_manager.Manager),
	})
	if err != nil {
		return "", "", err
	}

	configPath, err := parseWerfConfig(ctx, parseWerfConfigOpts{
		tmpl:               tmpl,
		giterminismManager: opts.giterminismManager.(*giterminism_manager.Manager),
		relWerfConfigPath:  opts.customWerfConfigRelPath,
	})
	if err != nil {
		return "", "", err
	}

	templateData := make(map[string]interface{})
	templateData["Files"] = files{
		ctx:                ctx,
		giterminismManager: opts.giterminismManager.(*giterminism_manager.Manager),
	}
	templateData["Env"] = opts.env

	headHash, err := opts.giterminismManager.LocalGitRepo().HeadCommitHash(ctx)
	if err != nil {
		return "", "", fmt.Errorf("unable to get HEAD commit hash: %w", err)
	}

	headTime, err := opts.giterminismManager.LocalGitRepo().HeadCommitTime(ctx)
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
	if err != nil {
		return "", "", detailedTemplateError(tmpl, detailedTemplateErrorData{
			templateName: "werfConfig",
		}, opts.debugTemplates, err)
	}

	return configPath, config, nil
}

type parseWerfConfigOpts struct {
	tmpl               *template.Template
	giterminismManager *giterminism_manager.Manager
	relWerfConfigPath  string
}

func parseWerfConfig(ctx context.Context, opts parseWerfConfigOpts) (string, error) {
	configPath, configData, err := opts.giterminismManager.FileManager.ReadConfig(ctx, opts.relWerfConfigPath)
	if err != nil {
		return "", err
	}

	if _, err := opts.tmpl.Parse(string(configData)); err != nil {
		return "", err
	}

	return configPath, nil
}

type parseWerfConfigTemplatesDirOpts struct {
	tmpl                                *template.Template
	customWerfConfigTemplatesDirRelPath string
	giterminismManager                  *giterminism_manager.Manager
}

func parseWerfConfigTemplatesDir(ctx context.Context, opts parseWerfConfigTemplatesDirOpts) error {
	err := opts.giterminismManager.FileManager.ReadConfigTemplateFiles(ctx, opts.customWerfConfigTemplatesDirRelPath, func(name, content string) error {
		return addTemplate(opts.tmpl, name, content)
	})
	if err != nil {
		return fmt.Errorf("unable to read werf config templates: %w", err)
	}

	return nil
}

func addTemplate(tmpl *template.Template, templateName, templateContent string) error {
	extraTemplate := tmpl.New(templateName)
	_, err := extraTemplate.Parse(templateContent)
	return err
}

func funcMap(ctx context.Context, tmpl *template.Template, giterminismManager giterminism_manager.Interface, debug bool) template.FuncMap {
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
		result, err := executeTemplate(tmpl, name, data)
		if err != nil {
			return "", detailedTemplateError(tmpl, detailedTemplateErrorData{
				funcName:     "include",
				templateName: name,
			}, debug, err)
		}

		return result, nil
	}
	funcMap["tpl"] = func(templateContent string, data interface{}) (string, error) {
		templateName := buildTplTemplateName(templateContent)

		if err := addTemplate(tmpl, templateName, templateContent); err != nil {
			return "", detailedTemplateError(tmpl, detailedTemplateErrorData{
				funcName:        "tpl",
				templateName:    templateName,
				templateContent: templateContent,
			}, debug, err)
		}

		result, err := executeTemplate(tmpl, templateName, data)
		if err != nil {
			return "", detailedTemplateError(tmpl, detailedTemplateErrorData{
				funcName:        "tpl",
				templateName:    templateName,
				templateContent: templateContent,
			}, debug, err)
		}

		return result, nil
	}

	funcMap["env"] = func(value interface{}, args ...string) (string, error) {
		if len(args) > 1 {
			return "", fmt.Errorf("more than 1 optional argument prohibited")
		}

		envVarName := fmt.Sprint(value)
		if err := giterminismManager.Inspector().InspectConfigGoTemplateRenderingEnv(ctx, envVarName); err != nil {
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

	// debug functions
	funcMap["tpl_debug"] = func(templateContent string, data interface{}) (string, error) {
		templateName := buildTplTemplateName(templateContent)

		if err := addTemplate(tmpl, templateName, templateContent); err != nil {
			return "", detailedTemplateError(tmpl, detailedTemplateErrorData{
				funcName:        "tpl_debug",
				templateName:    templateName,
				templateContent: templateContent,
			}, debug, err)
		}

		if debug {
			logboek.Context(ctx).Default().LogF("-- tpl_debug %q content:\n%s\n\n", templateName, templateContent)
		}

		result, err := executeTemplate(tmpl, templateName, data)
		if err != nil {
			return "", detailedTemplateError(tmpl, detailedTemplateErrorData{
				funcName:        "tpl_debug",
				templateName:    templateName,
				templateContent: templateContent,
			}, debug, err)
		}

		if debug {
			logboek.Context(ctx).Default().LogF("-- tpl_debug %q result:\n%s\n\n", templateName, result)
		}

		return result, nil
	}

	funcMap["include_debug"] = func(name string, data interface{}) (string, error) {
		result, execErr := executeTemplate(tmpl, name, data)

		var templateContent string
		if execErr != nil || debug {
			templateContent, _ = templateContentFromTree(tmpl, name)
		}

		if debug {
			logboek.Context(ctx).Default().LogF("-- include_debug template %q content:\n%s\n\n", name, templateContent)
		}

		if execErr != nil {
			return "", detailedTemplateError(tmpl, detailedTemplateErrorData{
				funcName:        "include_debug",
				templateName:    name,
				templateContent: templateContent,
			}, debug, execErr)
		}

		if debug {
			logboek.Context(ctx).Default().LogF("-- include_debug template %q result:\n%s\n\n", name, result)
		}

		return result, nil
	}

	funcMap["printf_debug"] = func(format string, args ...interface{}) string {
		if debug {
			logboek.Context(ctx).Default().LogF("-- printf_debug format %q result:\n%s\n\n", format, fmt.Sprintf(format, args...))
		}

		return ""
	}

	funcMap["dump_debug"] = func(obj interface{}) string {
		if debug {
			logboek.Context(ctx).Default().LogF("-- dump_debug result:\n%s\n\n", spew.Sdump(obj))
		}

		return ""
	}

	return funcMap
}

func buildTplTemplateName(templateContent string) string {
	return fmt.Sprintf("tpl-%d", adler32.Checksum([]byte(templateContent)))
}

func executeTemplate(tmpl *template.Template, name string, data interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	if err := tmpl.ExecuteTemplate(buf, name, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func templateContentFromTree(tmpl *template.Template, name string) (string, error) {
	t := tmpl.Lookup(name)
	if t == nil || t.Tree == nil || t.Tree.Root == nil {
		return "", fmt.Errorf("template %q not found", name)
	}

	return strings.TrimSpace(t.Tree.Root.String()), nil
}

type detailedTemplateErrorData struct {
	funcName        string
	templateName    string
	templateContent string
}

func detailedTemplateError(tmpl *template.Template, d detailedTemplateErrorData, debug bool, err error) error {
	if debug {
		if d.templateContent == "" {
			d.templateContent, _ = templateContentFromTree(tmpl, d.templateName)
		}

		return fmt.Errorf(
			"%w\n\nDetails:\n  Function name: %q\n  Template name: %q\n  Template content:\n%s",
			err,
			d.funcName,
			d.templateName,
			strings.TrimRightFunc(util.NumerateLines(d.templateContent, 1), unicode.IsSpace),
		)
	}
	if strings.Contains(err.Error(), engine.TemplateErrHint) {
		return err
	}

	return fmt.Errorf("%w\n%s", err, engine.TemplateErrHint)
}

type files struct {
	ctx                context.Context
	giterminismManager *giterminism_manager.Manager
}

func (f files) Get(relPath string) string {
	if res, err := f.giterminismManager.FileManager.ConfigGoTemplateFilesGet(f.ctx, relPath); err != nil {
		err := fmt.Errorf("{{ .Files.Get %q }}: %w", relPath, err)
		panic(err.Error())
	} else {
		return string(res)
	}
}

func (f files) doGlob(ctx context.Context, pattern string) (map[string]interface{}, error) {
	res, err := f.giterminismManager.FileManager.ConfigGoTemplateFilesGlob(ctx, pattern)
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

func (f files) Exists(relPath string) bool {
	exist, err := f.giterminismManager.FileManager.ConfigGoTemplateFilesExists(f.ctx, relPath)
	if err != nil {
		panic(err.Error())
	}
	return exist
}

func (f files) IsDir(relPath string) bool {
	exist, err := f.giterminismManager.FileManager.ConfigGoTemplateFilesIsDir(f.ctx, relPath)
	if err != nil {
		panic(err.Error())
	}
	return exist
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
	var images []ImageInterface

	for _, rawImage := range rawImagesFromDockerfile {
		imageList, err := rawImage.toImageFromDockerfileDirectives(giterminismManager)
		if err != nil {
			return nil, err
		}

		for _, image := range imageList {
			if meta.Build.ImageSpec != nil && image.final {
				merged := mergeImageSpec(meta.Build.ImageSpec, image.ImageSpec)
				image.ImageSpec = &merged
			}

			if util.GetBoolEnvironmentDefaultFalse("WERF_FORCE_STAGED_DOCKERFILE") {
				image.Staged = true
			} else if !rawImage.isFillStaged {
				image.Staged = meta.Build.Staged
			}
			images = append(images, image)
		}
	}

	for _, rawImage := range rawImages {
		if rawImage.stapelImageType() == "images" {
			imageList, err := rawImage.toStapelImageDirectives(giterminismManager)
			if err != nil {
				return nil, err
			}

			for _, image := range imageList {
				if meta.Build.ImageSpec != nil && image.final {
					merged := mergeImageSpec(meta.Build.ImageSpec, image.ImageSpec)
					image.ImageSpec = &merged
				}

				if image.Sbom() == nil && image.final {
					image.sbom = meta.Build.Sbom
				}

				images = append(images, image)
			}
		} else {
			if image, err := rawImage.toStapelImageArtifactDirectives(giterminismManager); err != nil {
				return nil, err
			} else {
				if meta.Build.ImageSpec != nil && image.final {
					merged := mergeImageSpec(meta.Build.ImageSpec, image.ImageSpec)
					image.ImageSpec = &merged
				}

				if image.Sbom() == nil && image.final {
					image.sbom = meta.Build.Sbom
				}

				images = append(images, image)
			}
		}
	}

	werfConfig := NewWerfConfig(meta, images)

	if err := werfConfig.validateConflictBetweenImagesNames(); err != nil {
		return nil, err
	}

	if err := werfConfig.validateRelatedImages(); err != nil {
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
			if isStagedDoc(raw) {
				imageFromDockerfile.isFillStaged = true
			}
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

func isStagedDoc(h map[string]interface{}) bool {
	if _, ok := h["staged"]; ok {
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
