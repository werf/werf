package container_backend

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type commitChangesAndRunArgsTest struct {
	CommitChangeOptions LegacyCommitChangeOptions
	Options             *LegacyStageImageContainerOptions

	ExpectedCommitChanges        []string
	ExpectedPrepareCommitChanges []string
}

var _ = Describe("LegacyStageImageContainerOptions", func() {
	DescribeTable("Docker server commit changes and run args generator", func(tst commitChangesAndRunArgsTest) {
		ctx := context.Background()

		commitChanges := tst.Options.toCommitChanges(tst.CommitChangeOptions)

		prepareCommitChanges, err := tst.Options.prepareCommitChanges(ctx, tst.CommitChangeOptions)
		Expect(err).To(Succeed())

		fmt.Printf("Commit changes result: %v\n", commitChanges)
		fmt.Printf("Prepare changes result: %v\n", prepareCommitChanges)

		dockerfile, err := parser.Parse(bytes.NewBufferString(strings.Join(prepareCommitChanges, "\n")))
		Expect(err).To(Succeed())

		for _, n := range dockerfile.AST.Children {
			_, err = instructions.ParseCommand(n)
			Expect(err).To(Succeed())
		}

		Expect(commitChanges).To(Equal(tst.ExpectedCommitChanges))
		Expect(prepareCommitChanges).To(Equal(tst.ExpectedPrepareCommitChanges))
	},

		Entry("empty", commitChangesAndRunArgsTest{
			Options: &LegacyStageImageContainerOptions{
				dockerServerVersion: "20.10.16",
			},
			ExpectedPrepareCommitChanges: []string{`ENTRYPOINT [""]`, `CMD []`},
			ExpectedCommitChanges:        nil,
		}),

		Entry("without exact interpretation (legacy)", commitChangesAndRunArgsTest{
			Options: &LegacyStageImageContainerOptions{
				dockerServerVersion: "20.10.16",
				Volume:              []string{"/volume1", "/volume2", `/my\ volume\ with\ spaces`},
				VolumesFrom:         []string{"container-1", "container-2"},
				Expose:              []string{"80/tcp", "449/upd"},
				Env: map[string]string{
					"MYVAR":  "V1",
					"MYVAR2": `"value with spaces"`,
				},
				Label: map[string]string{
					"version":    "v1.2.3",
					"maintainer": "vasilyi.ivanovich@gmail.com",
					"mylabel":    `"my value with spaces"`,
				},
				Cmd:         "server run",
				Workdir:     `"/work dir"`,
				User:        "1000:1000",
				Entrypoint:  `"/dir with spaces/bin/bash"`,
				HealthCheck: "--interval=30s --timeout=3s CMD curl -f http://localhost/ || exit 1",
			},
			ExpectedCommitChanges: []string{
				`VOLUME /volume1`,
				`VOLUME /volume2`,
				`VOLUME /my\ volume\ with\ spaces`,
				`EXPOSE 80/tcp`,
				`EXPOSE 449/upd`,
				`ENV MYVAR=V1`,
				`ENV MYVAR2="value with spaces"`,
				`LABEL maintainer=vasilyi.ivanovich@gmail.com`,
				`LABEL mylabel="my value with spaces"`,
				`LABEL version=v1.2.3`,
				`CMD server run`,
				`WORKDIR "/work dir"`,
				`USER 1000:1000`,
				`ENTRYPOINT "/dir with spaces/bin/bash"`,
				`HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost/ || exit 1`,
			},
			ExpectedPrepareCommitChanges: []string{
				`VOLUME /volume1`,
				`VOLUME /volume2`,
				`VOLUME /my\ volume\ with\ spaces`,
				`EXPOSE 80/tcp`,
				`EXPOSE 449/upd`,
				`ENV MYVAR=V1`,
				`ENV MYVAR2="value with spaces"`,
				`LABEL maintainer=vasilyi.ivanovich@gmail.com`,
				`LABEL mylabel="my value with spaces"`,
				`LABEL version=v1.2.3`,
				`WORKDIR "/work dir"`,
				`USER 1000:1000`,
				`ENTRYPOINT "/dir with spaces/bin/bash"`,
				`CMD server run`,
				`HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost/ || exit 1`,
			},
		}),

		Entry("with exact interpretation", commitChangesAndRunArgsTest{
			CommitChangeOptions: LegacyCommitChangeOptions{ExactValues: true},
			Options: &LegacyStageImageContainerOptions{
				dockerServerVersion: "20.10.16",
				Volume:              []string{"/volume1", "/volume2", "/my volume with spaces"},
				VolumesFrom:         []string{"container-1", "container-2"},
				Expose:              []string{"80/tcp", "449/upd"},
				Env: map[string]string{
					"MYVAR":  "V1",
					"MYVAR2": "value with spaces",
				},
				Label: map[string]string{
					"version":    "v1.2.3",
					"maintainer": "vasilyi.ivanovich@gmail.com",
					"mylabel":    "my value with spaces",
				},
				Cmd:         "server run",
				Workdir:     "/work dir",
				User:        "1000:1000",
				Entrypoint:  "/dir with spaces/bin/bash",
				HealthCheck: "--interval=30s --timeout=3s CMD curl -f http://localhost/ || exit 1",
			},
			ExpectedCommitChanges: []string{
				`VOLUME ["/volume1"]`,
				`VOLUME ["/volume2"]`,
				`VOLUME ["/my volume with spaces"]`,
				`EXPOSE "80/tcp"`,
				`EXPOSE "449/upd"`,
				`ENV MYVAR="V1"`,
				`ENV MYVAR2="value with spaces"`,
				`LABEL maintainer="vasilyi.ivanovich@gmail.com"`,
				`LABEL mylabel="my value with spaces"`,
				`LABEL version="v1.2.3"`,
				`CMD server run`,
				`WORKDIR /work dir`,
				`USER 1000:1000`,
				`ENTRYPOINT /dir with spaces/bin/bash`,
				`HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost/ || exit 1`,
			},
			ExpectedPrepareCommitChanges: []string{
				`VOLUME ["/volume1"]`,
				`VOLUME ["/volume2"]`,
				`VOLUME ["/my volume with spaces"]`,
				`EXPOSE "80/tcp"`,
				`EXPOSE "449/upd"`,
				`ENV MYVAR="V1"`,
				`ENV MYVAR2="value with spaces"`,
				`LABEL maintainer="vasilyi.ivanovich@gmail.com"`,
				`LABEL mylabel="my value with spaces"`,
				`LABEL version="v1.2.3"`,
				`WORKDIR /work dir`,
				`USER 1000:1000`,
				`ENTRYPOINT /dir with spaces/bin/bash`,
				`CMD server run`,
				`HEALTHCHECK --interval=30s --timeout=3s CMD curl -f http://localhost/ || exit 1`,
			},
		}),
	)
})
