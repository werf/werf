package cleaning

import mapset "github.com/deckarep/golang-set/v2"

type KeepList mapset.Set[string]

func NewKeepList(val ...string) KeepList {
	return mapset.NewSet[string](val...)
}

func NewKeepListWithSize(n int) KeepList {
	return mapset.NewSetWithSize[string](n)
}
