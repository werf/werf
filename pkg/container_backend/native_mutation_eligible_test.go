package container_backend

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/werf/werf/v2/pkg/image"
)

var _ = Describe("nativeMutationEligible", func() {
	base := func() *v1.ConfigFile {
		return &v1.ConfigFile{Config: v1.Config{
			Cmd:          []string{"/bin/sh"},
			Labels:       map[string]string{"a": "1"},
			Env:          []string{"FOO=bar"},
			Volumes:      map[string]struct{}{"/data": {}},
			ExposedPorts: map[string]struct{}{"80/tcp": {}},
		}}
	}

	DescribeTable("decides docker-commit eligibility",
		func(cfg image.SpecConfig, expected bool) {
			Expect(nativeMutationEligible(cfg, base())).To(Equal(expected))
		},
		Entry("additive superset labels+env", image.SpecConfig{
			Labels: map[string]string{"a": "1", "b": "2"},
			Env:    []string{"FOO=bar", "BAZ=qux"},
		}, true),
		Entry("nil labels/volumes/ports keep base (only env replaced)", image.SpecConfig{
			Env: []string{"FOO=other"},
		}, true),
		Entry("removed label", image.SpecConfig{
			Labels: map[string]string{"b": "2"},
			Env:    []string{"FOO=bar"},
		}, false),
		Entry("removed env var (nil env clears)", image.SpecConfig{
			Labels: map[string]string{"a": "1"},
		}, false),
		Entry("removed env var (non-nil env drops key)", image.SpecConfig{
			Env: []string{},
		}, false),
		Entry("changed env value keeps key", image.SpecConfig{
			Env: []string{"FOO=changed"},
		}, true),
		Entry("removed volume", image.SpecConfig{
			Env:     []string{"FOO=bar"},
			Volumes: map[string]struct{}{},
		}, false),
		Entry("removed exposed port", image.SpecConfig{
			Env:          []string{"FOO=bar"},
			ExposedPorts: map[string]struct{}{},
		}, false),
		Entry("ClearHistory", image.SpecConfig{
			Env:          []string{"FOO=bar"},
			ClearHistory: true,
		}, false),
		Entry("ClearCmd", image.SpecConfig{
			Env:      []string{"FOO=bar"},
			ClearCmd: true,
		}, false),
		Entry("ClearEntrypoint", image.SpecConfig{
			Env:             []string{"FOO=bar"},
			ClearEntrypoint: true,
		}, false),
		Entry("ClearUser", image.SpecConfig{
			Env:       []string{"FOO=bar"},
			ClearUser: true,
		}, false),
		Entry("ClearWorkingDir", image.SpecConfig{
			Env:             []string{"FOO=bar"},
			ClearWorkingDir: true,
		}, false),
	)

	DescribeTable("declines when the resulting image has no command",
		func(cfg image.SpecConfig, baseCfg *v1.ConfigFile, expected bool) {
			Expect(nativeMutationEligible(cfg, baseCfg)).To(Equal(expected))
		},
		Entry("base commandless, config adds Cmd", image.SpecConfig{
			Env: []string{"FOO=bar"}, Cmd: []string{"/bin/sh"},
		}, &v1.ConfigFile{Config: v1.Config{Env: []string{"FOO=bar"}}}, true),
		Entry("base has Entrypoint only", image.SpecConfig{
			Env: []string{"FOO=bar"},
		}, &v1.ConfigFile{Config: v1.Config{Env: []string{"FOO=bar"}, Entrypoint: []string{"/bin/app"}}}, true),
		Entry("both commandless", image.SpecConfig{
			Env: []string{"FOO=bar"},
		}, &v1.ConfigFile{Config: v1.Config{Env: []string{"FOO=bar"}}}, false),
	)
})
