package chart_ts_init

import (
	"context"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/werf/common-go/pkg/graceful"
	"github.com/werf/werf/v3/pkg/logging"
)

func TestChartTSInit(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "TypeScript chart initialization")
}

var _ = ginkgo.Describe("chart ts init", func() {
	ginkgo.It("creates the embedded scaffold with the werf render context", func() {
		dir := initProject()
		ctx := graceful.WithTermination(context.Background())
		defer graceful.Shutdown(ctx, func(_ context.Context, _ graceful.TerminationDescriptor) {})
		ctx = logging.WithLogger(ctx)
		cmd := NewCmd(ctx)
		cmd.SetArgs([]string{dir, "--home-dir", ginkgo.GinkgoT().TempDir()})
		gomega.Expect(cmd.Execute()).To(gomega.Succeed())

		tsDir := filepath.Join(dir, ".helm", "ts")
		var files []string
		gomega.Expect(filepath.WalkDir(tsDir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() {
				relPath, err := filepath.Rel(tsDir, path)
				if err != nil {
					return err
				}
				files = append(files, filepath.ToSlash(relPath))
			}
			return nil
		})).To(gomega.Succeed())
		gomega.Expect(files).To(gomega.ConsistOf(
			"input.example.yaml", "package.json", "tsconfig.json", "tsconfig.test.json",
			"src/context.ts", "src/deployment.ts", "src/deployment.test.ts",
			"src/helpers.ts", "src/index.ts", "src/service.ts",
		))

		gomega.Expect(readScaffoldFile(tsDir, "src/context.ts")).To(gomega.Equal(
			"export type { WerfRenderContext as ChartContext } from '@nelm/chart-ts-sdk';\n",
		))
		var input map[string]any
		gomega.Expect(yaml.Unmarshal([]byte(readScaffoldFile(tsDir, "input.example.yaml")), &input)).To(gomega.Succeed())
		gomega.Expect(input).To(gomega.HaveKeyWithValue("Chart", gomega.HaveKeyWithValue("Name", "scaffold-test")))
		gomega.Expect(input).To(gomega.HaveKeyWithValue("Values", gomega.HaveKeyWithValue("global", gomega.HaveKey("werf"))))

		var pkg struct {
			Name            string            `json:"name"`
			Scripts         map[string]string `json:"scripts"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		gomega.Expect(json.Unmarshal([]byte(readScaffoldFile(tsDir, "package.json")), &pkg)).To(gomega.Succeed())
		gomega.Expect(pkg.Name).To(gomega.Equal("scaffold-test-ts"))
		gomega.Expect(pkg.Scripts).To(gomega.HaveKeyWithValue("test", gomega.ContainSubstring("--test")))
		gomega.Expect(pkg.Scripts).To(gomega.HaveKeyWithValue("typecheck", gomega.And(
			gomega.ContainSubstring("tsconfig.json"), gomega.ContainSubstring("tsconfig.test.json"),
		)))
		gomega.Expect(pkg.DevDependencies).To(gomega.HaveKey("tsx"))
		gomega.Expect(pkg.DevDependencies).To(gomega.HaveKey("@types/node"))
	})
})
