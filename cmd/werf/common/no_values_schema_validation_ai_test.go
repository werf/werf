//go:build ai_tests

package common

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("SetupNoValuesSchemaValidationFlag", func() {
	It("registers the --no-values-schema-validation flag defaulting to false", func() {
		cmdData := &CmdData{}
		cmd := &cobra.Command{}

		Expect(SetupNoValuesSchemaValidationFlag(cmdData, cmd)).To(Succeed())

		flag := cmd.Flags().Lookup("no-values-schema-validation")
		Expect(flag).NotTo(BeNil())
		Expect(flag.DefValue).To(Equal("false"))
	})

	It("honors the $WERF_NO_VALUES_SCHEMA_VALIDATION env var default", func() {
		GinkgoT().Setenv("WERF_NO_VALUES_SCHEMA_VALIDATION", "true")

		cmdData := &CmdData{}
		cmd := &cobra.Command{}

		Expect(SetupNoValuesSchemaValidationFlag(cmdData, cmd)).To(Succeed())

		flag := cmd.Flags().Lookup("no-values-schema-validation")
		Expect(flag).NotTo(BeNil())
		Expect(flag.DefValue).To(Equal("true"))
		Expect(cmdData.NoValuesSchemaValidation).To(BeTrue())
	})

	It("binds the parsed value into CmdData.NoValuesSchemaValidation", func() {
		cmdData := &CmdData{}
		cmd := &cobra.Command{}

		Expect(SetupNoValuesSchemaValidationFlag(cmdData, cmd)).To(Succeed())
		Expect(cmd.Flags().Parse([]string{"--no-values-schema-validation"})).To(Succeed())
		Expect(cmdData.NoValuesSchemaValidation).To(BeTrue())
	})

	It("is registered as part of SetupResourceValidationFlags for the resource-validating commands", func() {
		cmdData := &CmdData{}
		cmd := &cobra.Command{}

		Expect(SetupResourceValidationFlags(cmdData, cmd)).To(Succeed())
		Expect(cmd.Flags().Lookup("no-values-schema-validation")).NotTo(BeNil())
	})
})
