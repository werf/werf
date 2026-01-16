package report

import (
	"context"
	"encoding/json"
	"os"

	. "github.com/onsi/gomega"

	"github.com/werf/3p-helm/pkg/release"
	"github.com/werf/werf/v2/pkg/build"

	werftest "github.com/werf/werf/v2/test/pkg/werf"
)

type Project struct {
	*werftest.Project
}

func NewProjectWithReport(p *werftest.Project) *Project {
	return &Project{Project: p}
}

func (p *Project) BuildWithReport(ctx context.Context, buildReportPath string, opts *werftest.WithReportOptions) (string, build.ImagesReport) {
	if opts == nil {
		opts = &werftest.WithReportOptions{}
	}

	args := append([]string{"build", "--save-build-report", "--build-report-path", buildReportPath}, opts.ExtraArgs...)
	out := p.RunCommand(ctx, args, opts.CommonOptions)
	buildReportRaw, err := os.ReadFile(buildReportPath)
	Expect(err).NotTo(HaveOccurred())

	var buildReport build.ImagesReport
	Expect(json.Unmarshal(buildReportRaw, &buildReport)).To(Succeed())

	return out, buildReport
}

func (p *Project) ConvergeWithReport(ctx context.Context, deployReportPath string, opts *werftest.WithReportOptions) (combinedOut string, report release.DeployReport) {
	if opts == nil {
		opts = &werftest.WithReportOptions{}
	}

	args := append([]string{"converge", "--save-deploy-report", "--deploy-report-path", deployReportPath}, opts.ExtraArgs...)
	out := p.RunCommand(ctx, args, opts.CommonOptions)

	deployReportRaw, err := os.ReadFile(deployReportPath)
	Expect(err).NotTo(HaveOccurred())

	var deployReport release.DeployReport
	Expect(json.Unmarshal(deployReportRaw, &deployReport)).To(Succeed())

	return out, deployReport
}

func (p *Project) BundlePublishWithReport(ctx context.Context, buildReportPath string, opts *werftest.WithReportOptions) (string, build.ImagesReport) {
	if opts == nil {
		opts = &werftest.WithReportOptions{}
	}

	args := append([]string{"bundle", "publish", "--save-build-report", "--build-report-path", buildReportPath}, opts.ExtraArgs...)
	out := p.RunCommand(ctx, args, opts.CommonOptions)

	buildReportRaw, err := os.ReadFile(buildReportPath)
	Expect(err).NotTo(HaveOccurred())

	var buildReport build.ImagesReport
	Expect(json.Unmarshal(buildReportRaw, &buildReport)).To(Succeed())

	return out, buildReport
}

func (p *Project) BundleApplyWithReport(ctx context.Context, releaseName, namespace, deployReportPath string, opts *werftest.WithReportOptions) (string, release.DeployReport) {
	if opts == nil {
		opts = &werftest.WithReportOptions{}
	}

	args := append([]string{"bundle", "apply", "--release", releaseName, "--namespace", namespace, "--save-deploy-report", "--deploy-report-path", deployReportPath}, opts.ExtraArgs...)
	out := p.RunCommand(ctx, args, opts.CommonOptions)

	deployReportRaw, err := os.ReadFile(deployReportPath)
	Expect(err).NotTo(HaveOccurred())

	var deployReport release.DeployReport
	Expect(json.Unmarshal(deployReportRaw, &deployReport)).To(Succeed())

	return out, deployReport
}
