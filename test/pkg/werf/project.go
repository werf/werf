package werf

import (
	"encoding/json"
	"os"

	. "github.com/onsi/gomega"

	iutils "github.com/werf/werf/integration/pkg/utils"
	"github.com/werf/werf/pkg/build"
)

func NewProject(werfBinPath, repoPath string) *Project {
	return &Project{
		WerfBinPath: werfBinPath,
		RepoPath:    repoPath,
	}
}

type Project struct {
	RepoPath    string
	WerfBinPath string
}

func (p *Project) Build(optArgs ...string) (combinedOut string) {
	optArgs = append([]string{"build", "--debug"}, optArgs...)
	return iutils.SucceedCommandOutputString(p.RepoPath, p.WerfBinPath, optArgs...)
}

func (p *Project) BuildWithReport(buildReportPath string, optsArgs ...string) (combinedOut string, buildReport build.ImagesReport) {
	optsArgs = append([]string{"build", "--debug", "--report-path", buildReportPath}, optsArgs...)
	combinedOut = iutils.SucceedCommandOutputString(p.RepoPath, p.WerfBinPath, optsArgs...)

	buildReportRaw, err := os.ReadFile(buildReportPath)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(buildReportRaw, &buildReport)).To(Succeed())

	return combinedOut, buildReport
}
