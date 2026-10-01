package common

import (
	"bytes"
	"net/http"
	"os"

	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/werf/v3/pkg/storage"
	"github.com/werf/werf/v3/pkg/storage/synchronization/server"
)

var _ = ginkgo.Describe("synchronization selection", func() {
	ginkgo.DescribeTable("redacts credentials from the selected synchronization address", func(address, expected string) {
		gomega.Expect(synchronizationAddressForLog(address)).To(gomega.Equal(expected))
	},
		ginkgo.Entry("plain HTTP", "https://sync.example/path", "https://sync.example/path"),
		ginkgo.Entry("HTTP credentials", "https://user:password@sync.example/path?token=secret#secret", "https://sync.example/path"),
		ginkgo.Entry("embedded kubeconfig", "kubernetes://namespace:context@base64:secret", "kubernetes://namespace:context@base64:[REDACTED]"),
		ginkgo.Entry("kubeconfig path", "kubernetes://namespace@/config", "kubernetes://namespace@/config"),
		ginkgo.Entry("local", ":local", ":local"),
		ginkgo.Entry("invalid", "https://user:secret@%invalid", "[invalid address]"))

	ginkgo.It("redacts credentials in the configured synchronization log", func(ctx ginkgo.SpecContext) {
		store := synchronizationTestStorage(ctx)
		gomega.Expect(store.PostClientIDRecord(ctx, "project", &storage.ClientIDRecord{ClientID: "client", TimestampMillisec: 1})).To(gomega.Succeed())
		address := "https://sync-user:sync-password@sync.example/path?token=sync-token#sync-fragment"
		var output bytes.Buffer
		ctxLog := logboek.NewContext(ctx, logboek.NewLogger(&output, &output))
		_, err := GetSynchronization(ctxLog, &CmdData{Synchronization: &address}, "project", store)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(output.String()).To(gomega.ContainSubstring("Using sync server: https://sync.example/path\n"))
		for _, secret := range []string{"sync-user", "sync-password", "sync-token", "sync-fragment"} {
			gomega.Expect(output.String()).NotTo(gomega.ContainSubstring(secret))
		}
	})

	ginkgo.DescribeTable("uses strict synchronization and logs only explicit settings", func(ctx ginkgo.SpecContext, source string) {
		previousEnv, present := os.LookupEnv("WERF_SYNCHRONIZATION")
		ginkgo.DeferCleanup(func() {
			if present {
				gomega.Expect(os.Setenv("WERF_SYNCHRONIZATION", previousEnv)).To(gomega.Succeed())
			} else {
				gomega.Expect(os.Unsetenv("WERF_SYNCHRONIZATION")).To(gomega.Succeed())
			}
		})
		gomega.Expect(os.Unsetenv("WERF_SYNCHRONIZATION")).To(gomega.Succeed())
		if source == "env" {
			gomega.Expect(os.Setenv("WERF_SYNCHRONIZATION", server.DefaultAddress)).To(gomega.Succeed())
		}
		command := &cobra.Command{}
		data := &CmdData{}
		SetupSynchronization(data, command)
		if source == "flag" {
			gomega.Expect(command.Flags().Set("synchronization", server.DefaultAddress)).To(gomega.Succeed())
		}
		store := synchronizationTestStorage(ctx)

		original := http.DefaultTransport
		http.DefaultTransport = &synchronizationDNSFailureTransport{next: original}
		ginkgo.DeferCleanup(func() { http.DefaultTransport = original })
		var output bytes.Buffer
		ctxLog := logboek.NewContext(ctx, logboek.NewLogger(&output, &output))
		_, err := GetSynchronization(ctxLog, data, "project", store)
		if source == "implicit" {
			gomega.Expect(output.String()).NotTo(gomega.ContainSubstring("Using sync server:"))
		} else {
			gomega.Expect(output.String()).To(gomega.ContainSubstring("Using sync server: " + server.DefaultAddress))
		}
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("no such host"))
	},
		ginkgo.Entry("implicit public", "implicit"),
		ginkgo.Entry("explicit public flag", "flag"),
		ginkgo.Entry("explicit public environment", "env"))
})
