package render_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/flant/werf/pkg/testing/utils"
)

var _ = Describe("helm render with secrets", func() {
	BeforeEach(func() {
		utils.CopyIn(utils.FixturePath("secret"), testDirPath)
	})

	It("should be rendered", func() {
		output := utils.SucceedCommandOutputString(
			testDirPath,
			werfBinPath,
			"helm", "render",
		)

		for _, substr := range []string{
			"int: MTc=",
			`json:
    web:
      id: "220821453116"
      redirect_uris:
      - https://test.ru`,
			"quoted: '''password'''",
			"secret_file: TUlJSktBSUJBQUtDQWdFQXFQMEFDcHZKNmUrSFJjdndDMUlFM201TzJNQXo4ci90QU1RZldwK2w3L0g1TE1zTgpGTzJKSVJPUis1VE95R2VWc1dzekRZMTEzcFhhRTJoaDZjam16Qkhvc1h6azZSenpVd1B4TlZzUy9MN3ZkRy9iClZxTnB1UGhyRllMdVNWcFgzZC9Cc0wzUmZXVnpxS1JlYVpnb05SQnpUN044UWNFVHdCMEEzbExvYy9TcGxNY0UKOTJ5QUJJMUNwcEt4b1RWVnlGZ09vMVNhSlFBRlRwOUtLbDY4K2tkTTFhcUlhV1Y0eWxKbXB2Umo3bVR2eGlmYwp3SVdLc1FLQ0FRQmo4c05RTHl3bnJIY2pJSnpwemRKZVArTUlCN1BaUFFIVDQzWDBpZzRFaC9kS3BIVHlWb21ECituWjZLWktqRUcySmY1NnNKUysxa202cWNhYjU1MnIxTGtNcXNoRVZWb2s2K2ovanNidHROOHVCVGliMkRyUksKbFlWMmI5TTh2RnF6LzNxQmM2VkxISWNkZ29pRGlOdWl3QlhlZFdsRHBnNUN2V3BIRWx1anRKVjNMNmNpQjFRNwpqeG9lMGNtUXY5TGJYWUNCRjBBNVZVNWZqeHdJN0tQVmRCSm1rWE9RS3EvZ1ZuVGFQOG9DSlN0dk95Lzd6bm05CjgwemhRQUtJMkgzVjZMOG9ONG5vTm16TW50bkNUWUl5UWFpMlRlUVgyckswbGlpeXNjVWQxaGJGaWE5bUZMNWEKd2g2Y3EyVElNcDNpK1h2MWw4VU5pTXZOM0JGWEhKR3hBb0lCQUMrajZZY2wxRVZuMjJCOFk3bHd1Mmp1SFFuYQpiZEpTR2QrMk5mUnp3TGcxaGxDc3pmL3JPNzZkUlYyVkFnR0NTM1NUWE9iQ2s2UHR5RTJHLy9TeG5pTG1nNCtyCklkLzF6SVNFZERUdkg1TG1OejNqNmJKNXFBTUttdkYyVjRoeHhtMmYwQ2xZSjJBWFFZdGNrL2N1bzQrWUxjVkQKOTRqNXowZkd4Y0xhSVkxVVpDUXZpbHZzVmdFaEV0cEQ0enpaTGFuOXNqRVVLLzAwV2E3NG8vVzBKWmpUS3Q3RAoxaDAwbUJxcTJDN1J0OElnNitreE9QSVdCUGxTNk1CQkJuS2Q1L1VqeS9hanJERnlHVjZjNjRtRFJHR1FFKy8rCk5JcDcwZjcxaWFBQVFQSVBkaFJUdk9zaEhna2RNQTFqRlBOWUNUVlVmWk56NllCSVpzajhBTktqNnNFPQo=",
		} {
			Ω(output).Should(ContainSubstring(substr))
		}
	})
})
