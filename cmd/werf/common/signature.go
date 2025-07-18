package common

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/werf/common-go/pkg/util"
	"github.com/werf/werf/v2/pkg/util/option"
)

func SetupSigningOptions(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SignManifest = new(bool)
	cmd.Flags().BoolVarP(cmdData.SignManifest, "sign-manifest", "", util.GetBoolEnvironmentDefaultFalse("WERF_SIGN_MANIFEST"),
		`Enable image manifest signing (default $WERF_SIGN_MANIFEST).
When enabled,
the private signing key must be specified with --sign-key option and
the certificate must be specified with --sign-cert option`)

	cmdData.SignKey = new(string)
	cmd.Flags().StringVarP(cmdData.SignKey, "sign-key", "", os.Getenv("WERF_SIGN_KEY"),
		"The private signing key as path to PEM file, base64-encoded PEM or hashivault://[KEY] (default $WERF_SIGN_KEY)")

	cmdData.SignCert = new(string)
	cmd.Flags().StringVarP(cmdData.SignCert, "sign-cert", "", os.Getenv("WERF_SIGN_CERT"),
		"The leaf certificate as path to PEM file or base64-encoded PEM (default $WERF_SIGN_CERT)")

	cmdData.SignIntermediates = new(string)
	cmd.Flags().StringVarP(cmdData.SignIntermediates, "sign-intermediates", "", os.Getenv("WERF_SIGN_INTERMEDIATES"),
		"The intermediate certificates as path to PEM file or base64-encoded PEM (default $WERF_SIGN_INTERMEDIATES)")
}

func GetSignManifest(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.SignManifest, false)
}

func SetupELFSigningOptions(cmdData *CmdData, cmd *cobra.Command) {
	cmdData.SignELFFiles = new(bool)
	cmd.Flags().BoolVarP(cmdData.SignELFFiles, "sign-elf-files", "", util.GetBoolEnvironmentDefaultFalse("WERF_SIGN_ELF_FILES"),
		`Enable ELF files signing (default $WERF_SIGN_ELF_FILES).
When enabled, the private signing key must be specified with --sign-key option and the certificate must be specified with --sign-cert option`)

	cmdData.BSignELFFiles = new(bool)
	cmd.Flags().BoolVarP(cmdData.BSignELFFiles, "bsign-elf-files", "", util.GetBoolEnvironmentDefaultFalse("WERF_BSIGN_ELF_FILES"),
		`Enable ELF files signing with bsign (default $WERF_BSIGN_ELF_FILES).
When enabled, the private elf key must be specified with --elf-pgp-private-key-base64 or --elf-pgp-private-key-fingerprint option`)

	cmdData.ELFPGPPrivateKeyBase64 = new(string)
	cmd.Flags().StringVarP(cmdData.ELFPGPPrivateKeyBase64, "elf-pgp-private-key-base64", "", os.Getenv("WERF_ELF_PGP_PRIVATE_KEY_BASE64"), "Base64-encoded PGP private key (default $WERF_ELF_PGP_PRIVATE_KEY_BASE64)")

	cmdData.ELFPGPPrivateKeyFingerprint = new(string)
	cmd.Flags().StringVarP(cmdData.ELFPGPPrivateKeyFingerprint, "elf-pgp-private-key-fingerprint", "", os.Getenv("WERF_ELF_PGP_PRIVATE_KEY_FINGERPRINT"), "PGP private key fingerprint (default $WERF_ELF_PGP_PRIVATE_KEY_FINGERPRINT)")

	cmdData.ELFPGPPrivateKeyPassphrase = new(string)
	cmd.Flags().StringVarP(cmdData.ELFPGPPrivateKeyPassphrase, "elf-pgp-private-key-passphrase", "", os.Getenv("WERF_ELF_PGP_PRIVATE_KEY_PASSPHRASE"), "Passphrase for the PGP private key (default $WERF_ELF_PGP_PRIVATE_KEY_PASSPHRASE)")
}

func GetSignELFFiles(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.SignELFFiles, false)
}

func GetBSignELFFiles(cmdData *CmdData) bool {
	return option.PtrValueOrDefault(cmdData.BSignELFFiles, false)
}
