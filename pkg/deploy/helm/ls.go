package helm

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"google.golang.org/grpc"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/timeconv"
)

type LsOptions struct {
	Short        bool
	Offset       string
	Reverse      bool
	Max          int64
	All          bool
	Deleted      bool
	Deleting     bool
	Pending      bool
	Namespace    string
	ColWidth     uint
	OutputFormat string
}

func Ls(out io.Writer, filter string, opts LsOptions) error {
	sortBy := services.ListSort_LAST_RELEASED
	sortOrder := services.ListSort_DESC
	if opts.Reverse {
		sortOrder = services.ListSort_ASC
	}

	store := &listReleasesStore{}
	err := tillerReleaseServer.ListReleases(&services.ListReleasesRequest{
		Limit:       opts.Max,
		Offset:      opts.Offset,
		SortBy:      sortBy,
		Filter:      filter,
		SortOrder:   sortOrder,
		StatusCodes: statusCodes(opts),
		Namespace:   opts.Namespace,
	}, store)
	if err != nil {
		return err
	}

	for _, resp := range store.Responses {
		rels := filterList(resp.GetReleases())
		result := getListResult(rels, resp.Next)

		output, err := formatResult(opts.OutputFormat, opts.Short, result, opts.ColWidth)
		if err != nil {
			return err
		}

		fmt.Fprintln(out, output)
	}

	return nil
}

// statusCodes gets the list of status codes that are to be included in the results.
func statusCodes(opts LsOptions) []release.Status_Code {
	if opts.All {
		return []release.Status_Code{
			release.Status_UNKNOWN,
			release.Status_DEPLOYED,
			release.Status_DELETED,
			release.Status_DELETING,
			release.Status_FAILED,
			release.Status_PENDING_INSTALL,
			release.Status_PENDING_UPGRADE,
			release.Status_PENDING_ROLLBACK,
		}
	}
	status := []release.Status_Code{}
	if opts.Deleted {
		status = append(status, release.Status_DELETED)
	}
	if opts.Deleting {
		status = append(status, release.Status_DELETING)
	}
	if opts.Pending {
		status = append(status, release.Status_PENDING_INSTALL, release.Status_PENDING_UPGRADE, release.Status_PENDING_ROLLBACK)
	}

	// Default case.
	if len(status) == 0 {
		status = append(status, release.Status_DEPLOYED, release.Status_FAILED)
	}
	return status
}

type listReleasesStore struct {
	grpc.ServerStream
	Responses []*services.ListReleasesResponse
}

func (store *listReleasesStore) Send(resp *services.ListReleasesResponse) error {
	store.Responses = append(store.Responses, resp)
	return nil
}

type listResult struct {
	Next     string
	Releases []listRelease
}

type listRelease struct {
	Name       string
	Revision   int32
	Updated    string
	Status     string
	Chart      string
	AppVersion string
	Namespace  string
}

// filterList returns a list scrubbed of old releases.
func filterList(rels []*release.Release) []*release.Release {
	idx := map[string]int32{}

	for _, r := range rels {
		name, version := r.GetName(), r.GetVersion()
		if max, ok := idx[name]; ok {
			// check if we have a greater version already
			if max > version {
				continue
			}
		}
		idx[name] = version
	}

	uniq := make([]*release.Release, 0, len(idx))
	for _, r := range rels {
		if idx[r.GetName()] == r.GetVersion() {
			uniq = append(uniq, r)
		}
	}
	return uniq
}

func getListResult(rels []*release.Release, next string) listResult {
	listReleases := []listRelease{}
	for _, r := range rels {
		md := r.GetChart().GetMetadata()
		t := "-"
		if tspb := r.GetInfo().GetLastDeployed(); tspb != nil {
			t = timeconv.String(tspb)
		}

		lr := listRelease{
			Name:       r.GetName(),
			Revision:   r.GetVersion(),
			Updated:    t,
			Status:     r.GetInfo().GetStatus().GetCode().String(),
			Chart:      fmt.Sprintf("%s-%s", md.GetName(), md.GetVersion()),
			AppVersion: md.GetAppVersion(),
			Namespace:  r.GetNamespace(),
		}
		listReleases = append(listReleases, lr)
	}

	return listResult{
		Releases: listReleases,
		Next:     next,
	}
}

func shortenListResult(result listResult) []string {
	names := []string{}
	for _, r := range result.Releases {
		names = append(names, r.Name)
	}

	return names
}

func formatTextShort(shortResult []string) string {
	return strings.Join(shortResult, "\n")
}

func formatText(result listResult, colWidth uint) string {
	nextOutput := ""
	if result.Next != "" {
		nextOutput = fmt.Sprintf("\tnext: %s\n", result.Next)
	}

	table := uitable.New()
	table.MaxColWidth = colWidth
	table.AddRow("NAME", "REVISION", "UPDATED", "STATUS", "CHART", "APP VERSION", "NAMESPACE")
	for _, lr := range result.Releases {
		table.AddRow(lr.Name, lr.Revision, lr.Updated, lr.Status, lr.Chart, lr.AppVersion, lr.Namespace)
	}

	return fmt.Sprintf("%s%s", nextOutput, table.String())
}

func formatResult(format string, short bool, result listResult, colWidth uint) (string, error) {
	var output string
	var err error

	var shortResult []string
	var finalResult interface{}
	if short {
		shortResult = shortenListResult(result)
		finalResult = shortResult
	} else {
		finalResult = result
	}

	switch format {
	case "table":
		if short {
			output = formatTextShort(shortResult)
		} else {
			output = formatText(result, colWidth)
		}
	case "json":
		o, e := json.Marshal(finalResult)
		if e != nil {
			err = fmt.Errorf("Failed to Marshal JSON output: %s", e)
		} else {
			output = string(o)
		}
	case "yaml":
		o, e := yaml.Marshal(finalResult)
		if e != nil {
			err = fmt.Errorf("Failed to Marshal YAML output: %s", e)
		} else {
			output = string(o)
		}
	default:
		err = fmt.Errorf("Unknown output format \"%s\"", format)
	}
	return output, err
}
