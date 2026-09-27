package builder

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v3/pkg/config"
)

func TestBuilder(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Builder Suite")
}

var _ = ginkgo.Describe("Shell stage checksums", func() {
	ginkgo.DescribeTable("preserves commands and cache versions across repeated reads",
		func(ctx ginkgo.SpecContext, stage string, shell config.Shell, expected string) {
			b := NewShellBuilder(&shell, &Extra{}, []config.Secret{}, "")
			for i := 0; i < 3; i++ {
				gomega.Expect(b.stageChecksum(ctx, stage)).To(gomega.Equal(expected))
			}
			gomega.Expect(b.stageChecksum(ctx, "Setup")).To(gomega.BeEmpty())
		},
		ginkgo.Entry("empty", "Install", config.Shell{}, ""),
		ginkgo.Entry("commands", "Install", config.Shell{Install: []string{"echo one", "echo two"}}, util.Sha256Hash("echo one", "echo two")),
		ginkgo.Entry("stage version", "BeforeInstall", config.Shell{BeforeInstall: []string{"echo one"}, BeforeInstallCacheVersion: "v1"}, util.Sha256Hash("echo one", util.Sha256Hash("v1"))),
	)
	ginkgo.It("keeps stage and global versions distinct", func(ctx ginkgo.SpecContext) {
		b := NewShellBuilder(&config.Shell{Install: []string{"echo one"}, CacheVersion: "global", InstallCacheVersion: "stage"}, &Extra{}, []config.Secret{}, "")
		gomega.Expect(b.InstallChecksum(ctx)).To(gomega.Equal(util.Sha256Hash("echo one", util.Sha256Hash("stage", "global"))))
		gomega.Expect(b.SetupChecksum(ctx)).To(gomega.Equal(util.Sha256Hash(util.Sha256Hash("global"))))
		gomega.Expect(b.InstallChecksum(ctx)).To(gomega.Equal(util.Sha256Hash("echo one", util.Sha256Hash("stage", "global"))))
	})
	ginkgo.It("owns the commands and versions used by checksums and execution", func(ctx ginkgo.SpecContext) {
		commands := []string{"echo original"}
		shell := &config.Shell{
			BeforeInstall: commands, Install: commands, BeforeSetup: commands, Setup: commands,
			CacheVersion: "global", BeforeInstallCacheVersion: "stage", InstallCacheVersion: "stage",
			BeforeSetupCacheVersion: "stage", SetupCacheVersion: "stage",
		}
		b := NewShellBuilder(shell, &Extra{}, []config.Secret{}, "")
		expected := util.Sha256Hash("echo original", util.Sha256Hash("stage", "global"))
		gomega.Expect(b.InstallChecksum(ctx)).To(gomega.Equal(expected))
		commands[0] = "echo modified"
		shell.Install = []string{"echo replaced"}
		shell.CacheVersion = "modified"
		shell.BeforeInstallCacheVersion = "modified"
		for _, name := range []string{"BeforeInstall", "Install", "BeforeSetup", "Setup"} {
			gomega.Expect(b.stageChecksum(ctx, name)).To(gomega.Equal(expected))
			gomega.Expect(b.stageCommands(name)).To(gomega.Equal([]string{"echo original"}))
		}
	})

	ginkgo.It("supports concurrent first checksum reads", func(ctx ginkgo.SpecContext) {
		b := NewShellBuilder(&config.Shell{CacheVersion: "global"}, &Extra{}, []config.Secret{}, "")
		names := []string{"BeforeInstall", "Install", "BeforeSetup", "Setup"}
		start := make(chan struct{})
		results := make(chan string, 32)
		for i := range cap(results) {
			go func() {
				<-start
				results <- b.stageChecksum(ctx, names[i%len(names)])
			}()
		}
		close(start)
		for range cap(results) {
			gomega.Expect(<-results).To(gomega.Equal(util.Sha256Hash(util.Sha256Hash("global"))))
		}
	})
})

func BenchmarkShellStageChecksum(b *testing.B) {
	builder := NewShellBuilder(&config.Shell{Install: []string{"echo one", "echo two"}, CacheVersion: "global", InstallCacheVersion: "stage"}, &Extra{}, []config.Secret{}, "")
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		builder.InstallChecksum(ctx)
	}
}
