package helm

import (
	"io"
	"text/template"
	"time"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/timeconv"
)

var printReleaseTemplate = `REVISION: {{.Release.Version}}
RELEASED: {{.ReleaseDate}}
CHART: {{.Release.Chart.Metadata.Name}}-{{.Release.Chart.Metadata.Version}}
USER-SUPPLIED VALUES:
{{.Release.Config.Raw}}
COMPUTED VALUES:
{{.ComputedValues}}
HOOKS:
{{- range .Release.Hooks }}
---
# {{.Name}}
{{.Manifest}}
{{- end }}
MANIFEST:
{{.Release.Manifest}}
`

func tpl(t string, vals interface{}, out io.Writer) error {
	tt, err := template.New("_").Parse(t)
	if err != nil {
		return err
	}
	return tt.Execute(out, vals)
}

func printRelease(out io.Writer, rel *release.Release) error {
	if rel == nil {
		return nil
	}

	cfg, err := chartutil.CoalesceValues(rel.Chart, rel.Config)
	if err != nil {
		return err
	}
	cfgStr, err := cfg.YAML()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Release":        rel,
		"ComputedValues": cfgStr,
		"ReleaseDate":    timeconv.Format(rel.Info.LastDeployed, time.ANSIC),
	}
	return tpl(printReleaseTemplate, data, out)
}

