package util

import "fmt"

const DAPPDEPS_BASE_VERSION = "0.2.3"

func DappdepsBaseBin(cmd string) string {
	return fmt.Sprintf("/.dapp/deps/base/%s/embedded/bin/%s", DAPPDEPS_BASE_VERSION, cmd)
}
