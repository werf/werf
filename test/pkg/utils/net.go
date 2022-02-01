package utils

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/gomega"
)

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
