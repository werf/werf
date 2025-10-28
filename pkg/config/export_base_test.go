package config_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/werf/werf/v2/pkg/config"
)

var _ = Describe("ExportBase", func() {
	DescribeTable(
		"AutoExcludeExportAndCheck()",
		func(left, right *config.ExportBase, expected types.GomegaMatcher) {
			Expect(left.AutoExcludeExportAndCheck(right)).To(expected)
		},
		Entry(
			"should return false and don't auto exclude",
			&config.ExportBase{
				Add: "/a",
				To:  "/",
			},
			&config.ExportBase{
				Add: "/b",
				To:  "/",
			},
			BeFalse(),
		),
		Entry(
			"should return false and don't auto exclude if include paths contains **/*",
			&config.ExportBase{
				Add: "/a",
				To:  "/",
				IncludePaths: []string{
					"**/*",
				},
			},
			&config.ExportBase{
				Add: "/a",
				To:  "/",
			},
			BeFalse(),
		),
		Entry(
			"should return false and don't auto exclude if include paths contains **/* (the same as above, symmetric)",
			&config.ExportBase{
				Add: "/a",
				To:  "/",
			},
			&config.ExportBase{
				Add: "/a",
				To:  "/",
				IncludePaths: []string{
					"**/*",
				},
			},
			BeFalse(),
		),
	)
})
