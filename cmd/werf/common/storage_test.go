package common

import "testing"

func TestMetaRepoSafeguardBypassIsOptIn(t *testing.T) {
	c := &NewStorageManagerConfig{}
	for _, opt := range []NewStorageManagerOption{WithHostPurge()} {
		opt(c)
	}
	if c.skipMetaRepoSafeguard {
		t.Fatal("ordinary storage manager construction must not skip the meta-repo safeguard")
	}

	WithSkipMetaRepoSafeguard()(c)
	if !c.skipMetaRepoSafeguard {
		t.Fatal("WithSkipMetaRepoSafeguard must enable the bypass")
	}
}
