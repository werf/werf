package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/timeconv"
)

type HistoryOptions struct {
	Max          int64
	OutputFormat string
	ColWidth     uint
}

func History(out io.Writer, releaseName string, opts HistoryOptions) error {
	r, err := tillerReleaseServer.GetHistory(context.Background(), &services.GetHistoryRequest{
		Name: releaseName,
		Max:  int32(opts.Max),
	})
	if err != nil {
		return err
	}
	if len(r.Releases) == 0 {
		return nil
	}

	releases := getReleaseHistory(r.Releases)

	var history []byte
	var formattingError error

	switch opts.OutputFormat {
	case "yaml":
		history, formattingError = yaml.Marshal(releases)
	case "json":
		history, formattingError = json.Marshal(releases)
	case "table":
		history = formatAsTable(releases, opts.ColWidth)
	default:
		return fmt.Errorf("unknown output format %q", opts.OutputFormat)
	}

	if formattingError != nil {
		return formattingError
	}

	fmt.Fprintln(out, string(history))

	return nil
}

type releaseInfo struct {
	Revision    int32  `json:"revision"`
	Updated     string `json:"updated"`
	Status      string `json:"status"`
	Chart       string `json:"chart"`
	AppVersion  string `json:"appVersion"`
	Description string `json:"description"`
}

type releaseInfos []releaseInfo

func getReleaseHistory(rls []*release.Release) (history releaseInfos) {
	for i := len(rls) - 1; i >= 0; i-- {
		r := rls[i]
		c := formatChartname(r.Chart)
		a := appVersionFromChart(r.Chart)
		t := timeconv.String(r.Info.LastDeployed)
		s := r.Info.Status.Code.String()
		v := r.Version
		d := r.Info.Description
		rInfo := releaseInfo{
			Revision:    v,
			Updated:     t,
			Status:      s,
			Chart:       c,
			AppVersion:  a,
			Description: d,
		}
		history = append(history, rInfo)
	}

	return history
}

func formatAsTable(releases releaseInfos, colWidth uint) []byte {
	tbl := uitable.New()

	tbl.MaxColWidth = colWidth
	tbl.AddRow("REVISION", "UPDATED", "STATUS", "CHART", "APP VERSION", "DESCRIPTION")
	for i := 0; i <= len(releases)-1; i++ {
		r := releases[i]
		tbl.AddRow(r.Revision, r.Updated, r.Status, r.Chart, r.AppVersion, r.Description)
	}
	return tbl.Bytes()
}

func formatChartname(c *chart.Chart) string {
	if c == nil || c.Metadata == nil {
		// This is an edge case that has happened in prod, though we don't
		// know how: https://github.com/kubernetes/helm/issues/1347
		return "MISSING"
	}
	return fmt.Sprintf("%s-%s", c.Metadata.Name, c.Metadata.Version)
}

func appVersionFromChart(c *chart.Chart) string {
	if c == nil || c.Metadata == nil {
		return "MISSING"
	}
	return c.Metadata.AppVersion
}
