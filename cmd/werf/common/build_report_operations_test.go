package common

import (
	"context"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/werf/logboek"
	"github.com/werf/logboek/pkg/level"
	"github.com/werf/werf/v2/pkg/config"
	"github.com/werf/werf/v2/pkg/opstats"
)

var _ = Describe("build report operations option", func() {
	Describe("SetupBuildReportOperations", func() {
		It("defaults to false", func() {
			cmdData := &CmdData{}
			SetupBuildReportOperations(cmdData, &cobra.Command{})
			Expect(GetBuildReportOperations(cmdData)).To(BeFalse())
		})

		It("is enabled by the flag", func() {
			cmdData := &CmdData{}
			cmd := &cobra.Command{}
			SetupBuildReportOperations(cmdData, cmd)
			Expect(cmd.Flags().Parse([]string{"--build-report-operations"})).To(Succeed())
			Expect(GetBuildReportOperations(cmdData)).To(BeTrue())
		})

		It("is enabled by $WERF_BUILD_REPORT_OPERATIONS", func() {
			GinkgoT().Setenv("WERF_BUILD_REPORT_OPERATIONS", "1")
			cmdData := &CmdData{}
			SetupBuildReportOperations(cmdData, &cobra.Command{})
			Expect(GetBuildReportOperations(cmdData)).To(BeTrue())
		})
	})

	DescribeTable("wiring into build options",
		func(enabled bool) {
			cmdData := &CmdData{
				Dev:                   lo.ToPtr(false),
				BuildReportOperations: lo.ToPtr(enabled),
			}
			werfConfig := &config.WerfConfig{Meta: &config.Meta{}}

			buildOptions, err := GetBuildOptions(context.Background(), cmdData, werfConfig, config.ImagesToProcess{})
			Expect(err).NotTo(HaveOccurred())
			Expect(buildOptions.ReportOperations).To(Equal(enabled))

			builtOptions, err := GetShouldBeBuiltOptions(cmdData, werfConfig, config.ImagesToProcess{})
			Expect(err).NotTo(HaveOccurred())
			Expect(builtOptions.ReportOperations).To(Equal(enabled))
		},
		Entry("enabled", true),
		Entry("disabled", false),
	)

	Describe("InitOperationsStatistics", func() {
		newCtx := func(acceptedLevel level.Level) context.Context {
			logger := logboek.NewLogger(io.Discard, io.Discard)
			logger.SetAcceptedLevel(acceptedLevel)
			return logboek.NewContext(context.Background(), logger)
		}

		It("does not install a collector when disabled", func() {
			cmdData := &CmdData{BuildReportOperations: lo.ToPtr(false)}
			ctx, finish := InitOperationsStatistics(newCtx(level.Default), cmdData)
			Expect(opstats.FromContext(ctx)).To(BeNil())
			Expect(finish).NotTo(BeNil())
		})

		It("installs a command-scoped collector when the flag is set", func() {
			cmdData := &CmdData{BuildReportOperations: lo.ToPtr(true)}
			ctx, finish := InitOperationsStatistics(newCtx(level.Default), cmdData)
			Expect(opstats.FromContext(ctx)).NotTo(BeNil())
			Expect(finish).NotTo(BeNil())
		})

		It("installs a collector under debug logging without the flag", func() {
			cmdData := &CmdData{BuildReportOperations: lo.ToPtr(false)}
			ctx, _ := InitOperationsStatistics(newCtx(level.Debug), cmdData)
			Expect(opstats.FromContext(ctx)).NotTo(BeNil())
		})
	})
})
