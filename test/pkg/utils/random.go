package utils

import (
	"strings"

	"github.com/Masterminds/goutils"
	. "github.com/onsi/gomega"
)

func GetRandomString(n int) string {
	str, err := goutils.CryptoRandomAlphaNumeric(n)
	Ω(err).ShouldNot(HaveOccurred())
	return strings.ToLower(str)
}
