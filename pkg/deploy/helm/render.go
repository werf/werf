package helm

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/strvals"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/tiller/environment"
	"k8s.io/helm/pkg/timeconv"
	"k8s.io/helm/pkg/version"
)

var (
	defaultKubeVersion = fmt.Sprintf("%s.%s", chartutil.DefaultKubeVersion.Major, chartutil.DefaultKubeVersion.Minor)
)

type valueFiles []string

func (v *valueFiles) String() string {
	return fmt.Sprint(*v)
}

func (v *valueFiles) Type() string {
	return "valueFiles"
}

func (v *valueFiles) Set(value string) error {
	for _, filePath := range strings.Split(value, ",") {
		*v = append(*v, filePath)
	}
	return nil
}

type RenderOptions struct {
	ShowNotes bool
}

func Render(out io.Writer, chartPath, releaseName, namespace string, values []string, secretValues []map[string]interface{}, set, setString []string, opts RenderOptions) error {
	// get combined values and create config
	rawVals, err := vals(values, secretValues, set, setString, []string{}, "", "", "")
	if err != nil {
		return err
	}
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := loadChartfile(chartPath)
	if err != nil {
		return err
	}

	renderOpts := renderOptions{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      releaseName,
			IsInstall: false,
			IsUpgrade: true,
			Time:      timeconv.Now(),
			Namespace: namespace,
		},
		KubeVersion: defaultKubeVersion,
	}

	renderedTemplates, err := render(c, config, renderOpts)
	if err != nil {
		return err
	}

	listManifests := manifest.SplitManifests(renderedTemplates)
	var manifestsToRender []manifest.Manifest

	manifestsToRender = listManifests
	for _, m := range tiller.SortByKind(manifestsToRender) {
		data := m.Content
		b := filepath.Base(m.Name)

		if !opts.ShowNotes && b == "NOTES.txt" {
			continue
		}

		if strings.HasPrefix(b, "_") {
			continue
		}

		fmt.Fprintf(out, "---\n# Source: %s\n", m.Name)
		fmt.Fprintln(out, data)
	}

	return nil
}

func vals(values valueFiles, secretValues []map[string]interface{}, set []string, setString []string, setFile []string, CertFile, KeyFile, CAFile string) ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range values {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			bytes, err = readFile(filePath, CertFile, KeyFile, CAFile)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.UnmarshalStrict(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	for i := range secretValues {
		base = mergeValues(base, secretValues[i])
	}

	// User specified a value via --set
	for _, value := range set {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	// User specified a value via --set-string
	for _, value := range setString {
		if err := strvals.ParseIntoString(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-string data: %s", err)
		}
	}

	// User specified a value via --set-file
	for _, value := range setFile {
		reader := func(rs []rune) (interface{}, error) {
			bytes, err := readFile(string(rs), CertFile, KeyFile, CAFile)
			return string(bytes), err
		}
		if err := strvals.ParseIntoFile(value, base, reader); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set-file data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

// Merges source and destination map, preferring values from the source map
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

//readFile load a file from the local directory or a remote file with a url.
func readFile(filePath, CertFile, KeyFile, CAFile string) ([]byte, error) {
	u, _ := url.Parse(filePath)
	p := getter.All(helmSettings)

	// FIXME: maybe someone handle other protocols like ftp.
	getterConstructor, err := p.ByScheme(u.Scheme)

	if err != nil {
		return ioutil.ReadFile(filePath)
	}

	getterIns, err := getterConstructor(filePath, CertFile, KeyFile, CAFile)
	if err != nil {
		return []byte{}, err
	}
	data, err := getterIns.Get(filePath)
	return data.Bytes(), err
}

// Options are options for this simple local render
type renderOptions struct {
	ReleaseOptions chartutil.ReleaseOptions
	KubeVersion    string
}

func render(c *chart.Chart, config *chart.Config, opts renderOptions) (map[string]string, error) {
	if req, err := chartutil.LoadRequirements(c); err == nil {
		if err := renderutil.CheckDependencies(c, req); err != nil {
			return nil, err
		}
	} else if err != chartutil.ErrRequirementsNotFound {
		return nil, fmt.Errorf("cannot load requirements: %v", err)
	}

	err := chartutil.ProcessRequirementsEnabled(c, config)
	if err != nil {
		return nil, err
	}

	err = chartutil.ProcessRequirementsImportValues(c)
	if err != nil {
		return nil, err
	}

	// Set up engine.
	engine := environment.DefaultEngine
	if c.Metadata.Engine != "" {
		engine = c.Metadata.Engine
	}
	renderer := tillerSettings.EngineYard[engine]

	if renderer == nil {
		return nil, fmt.Errorf("unknown engine '%s': check engine field in Chart.yaml", engine)
	}

	caps := &chartutil.Capabilities{
		APIVersions:   chartutil.DefaultVersionSet,
		KubeVersion:   chartutil.DefaultKubeVersion,
		TillerVersion: version.GetVersionProto(),
	}

	if opts.KubeVersion != "" {
		kv, verErr := semver.NewVersion(opts.KubeVersion)
		if verErr != nil {
			return nil, fmt.Errorf("could not parse a Kubernetes version: %v", verErr)
		}
		caps.KubeVersion.Major = fmt.Sprint(kv.Major())
		caps.KubeVersion.Minor = fmt.Sprint(kv.Minor())
		caps.KubeVersion.GitVersion = fmt.Sprintf("v%d.%d.0", kv.Major(), kv.Minor())
	}

	vals, err := chartutil.ToRenderValuesCaps(c, config, opts.ReleaseOptions, caps)
	if err != nil {
		return nil, err
	}

	return renderer.Render(c, vals)
}
