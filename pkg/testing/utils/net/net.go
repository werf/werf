package net

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/gomega"

	"github.com/flant/shluz"

	"github.com/flant/werf/pkg/util"
)

func init() {
	if err := shluz.Init(util.ExpandPath(filepath.Join("~/.werf", "service", "locks"))); err != nil {
		panic(err)
	}
}

const DefaultWaitTillHostReadyToRespondMaxAttempts = 60

func WaitTillHostReadyToRespond(url string, maxAttempts int) {
	var attemptCounter int

	for {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				break
			}
		}

		attemptCounter++
		Ω(attemptCounter).Should(BeNumerically("<", maxAttempts), fmt.Sprintf("max attempts reached %d (%s): %s", maxAttempts, url, err))
		time.Sleep(1 * time.Second)
	}
}

func GetFreeTCPHostPort() int {
	ln, err := net.Listen("tcp", "[::]:0")
	Ω(err).ShouldNot(HaveOccurred(), "net listen")

	port := ln.Addr().(*net.TCPAddr).Port
	Ω(ln.Close()).Should(Succeed(), "ln close")

	lockName := fmt.Sprintf("GetFreeTCPHostPort:%s", strconv.Itoa(port))
	isAcquired, err := shluz.TryLock(lockName, shluz.TryLockOptions{ReadOnly: false})
	Ω(err).ShouldNot(HaveOccurred(), "lock port")
	if !isAcquired {
		return GetFreeTCPHostPort()
	}

	return port
}
