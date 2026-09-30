package e2e_build_test

import (
	"fmt"
	"path"
)

type gitOwnershipTestOptions struct {
	setupEnvOptions
	OwnerFirst bool
}

func (opts gitOwnershipTestOptions) env() setupEnvOptions {
	return opts.setupEnvOptions
}

const (
	gitOwnershipDefaultImage = "img-default"
	gitOwnershipOwnedImage   = "img-owned"
)

func gitOwnershipChecks(ownerGroup, fileContent string, extraPaths ...string) []string {
	checks := []string{
		fmt.Sprintf("test %q = \"$(cat /app/file)\"", fileContent),
		fmt.Sprintf("test %q = \"$(stat -c %%u:%%g /app/file)\"", ownerGroup),
		"test -L /app/link",
		fmt.Sprintf("test %q = \"$(stat -c %%u:%%g /app/link)\"", ownerGroup),
		`test "3001:3002" = "$(stat -c %u:%g /outside/target)"`,
		`test "4001:4002" = "$(stat -c %u:%g /app/base-file)"`,
		`test "installed" = "$(cat /sentinel)"`,
		`test "0:0" = "$(stat -c %u:%g /sentinel)"`,
	}

	for _, extraPath := range extraPaths {
		checks = append(checks,
			fmt.Sprintf("test -e %s", extraPath),
			fmt.Sprintf("test %q = \"$(stat -c %%u:%%g %s)\"", ownerGroup, extraPath),
		)

		for dir := path.Dir(extraPath); dir != "/app" && dir != "/"; dir = path.Dir(dir) {
			checks = append(checks,
				fmt.Sprintf("test -d %s", dir),
				fmt.Sprintf("test %q = \"$(stat -c %%u:%%g %s)\"", ownerGroup, dir),
			)
		}
	}

	return checks
}
