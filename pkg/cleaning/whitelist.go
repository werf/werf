package cleaning

import mapset "github.com/deckarep/golang-set/v2"

type Whitelist mapset.Set[string]

func NewWhitelist(val ...string) Whitelist {
	return mapset.NewSet[string](val...)
}

func NewWhitelistWithSize(n int) Whitelist {
	return mapset.NewSetWithSize[string](n)
}
