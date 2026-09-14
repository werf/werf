package config

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("werf.yaml JSON schema", func() {
	var schema *jsonschema.Schema

	BeforeEach(func() {
		schema = compileWerfSchema()
	})

	It("accepts every non-templated werf.yaml fixture in the repository", func() {
		fixtures := nonTemplatedWerfYamlFixtures("../../test")
		Expect(fixtures).NotTo(BeEmpty())

		for _, path := range fixtures {
			data, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())

			for i, document := range yamlDocuments(data) {
				if err := schema.Validate(document); err != nil {
					Fail(fmt.Sprintf("%s: document %d: %v", path, i, err))
				}
			}
		}
	})

	DescribeTable("accepts a document",
		func(document string) {
			Expect(schema.Validate(yamlDocument(document))).To(Succeed())
		},
		Entry("meta with every section", `
configVersion: 1
project: app
build:
  cacheVersion: "1"
  platform: [linux/amd64, linux/arm64]
  staged: true
  imageSpec:
    author: me
    clearHistory: true
    config:
      keepEssentialWerfLabels: true
      removeLabels: ["/^io\\.k8s\\..*/"]
      labels:
        team: backend
deploy:
  helmChartDir: .helm
  helmChartConfig:
    appVersion: "1.2.3"
  helmRelease: "[[ project ]]-[[ env ]]"
  helmReleaseSlug: false
  namespace: "[[ project ]]"
  namespaceSlug: true
cleanup:
  disable: false
  disableKubernetesBasedPolicy: true
  disableGitHistoryBasedPolicy: true
  disableBuiltWithinLastNHoursPolicy: true
  keepImagesBuiltWithinLastNHours: 24
  keepPolicies:
  - references:
      branch: /.*/
      limit:
        last: 10
        in: 168h
        operator: Or
    imagesPerReference:
      last: 2
      in: 720h
  - references:
      tag: /v.*/
gitWorktree:
  forceShallowClone: true
  allowUnshallow: false
  allowFetchOriginBranchesAndTags: false
`),
		Entry("dockerfile image with every directive", `
image: [backend, backend-alias]
final: false
dockerfile: Dockerfile
staged: true
cacheVersion: "2"
context: backend
contextAddFiles: [dist]
platform: [linux/amd64]
target: production
args:
  VERSION: "1.0"
  RETRIES: 3
  DEBUG: true
addHost: registry.local:10.0.0.1
network: host
secrets:
- env: GITHUB_TOKEN
- src: ~/.npmrc
- id: api-key
  value: secret
dependencies:
- from: base
  imports:
  - type: ImageName
    targetBuildArg: BASE_IMAGE
imageSpec:
  author: me
  clearHistory: true
  config:
    clearCmd: true
    clearEntrypoint: true
    clearUser: true
    clearWorkingDir: true
    removeLabels: [a, /b.*/]
    removeVolumes: [/data]
    removeEnv: [/^DEBUG_.*/]
    volumes: [/cache]
    labels: {a: b}
    env: {A: b}
    expose: ["8080/tcp"]
    user: app
    cmd: [run]
    entrypoint: [/bin/sh, -c]
    workingDir: /app
    stopSignal: SIGTERM
    healthcheck:
      test: [CMD, curl, -f, http://localhost/]
      interval: 30
      timeout: 5
      startperiod: 10
      retries: 3
`),
		Entry("stapel image with every directive", `
image: app
final: true
cacheVersion: "1"
platform: [linux/amd64]
from: alpine:3.20
fromLatest: true
fromCacheVersion: "1"
network: host
git:
- to: /app
  includePaths: src
  excludePaths: [docs, "*.md"]
  owner: 1000
  group: app
  stageDependencies:
    install: package.json
    beforeSetup: [config/**]
    setup: "*.env"
- url: https://github.com/werf/werf.git
  basicAuth:
    username: bot
    password:
      env: GIT_PASSWORD
  branch: main
  add: /cmd
  to: /werf
- url: https://github.com/werf/werf.git
  tag: v2.0.0
  to: /werf-tag
- url: https://github.com/werf/werf.git
  commit: 0123456789abcdef0123456789abcdef01234567
  to: /werf-commit
shell:
  beforeInstall: apk add curl
  install: [npm ci]
  beforeSetup: []
  setup: [npm run build]
  cacheVersion: "1"
  beforeInstallCacheVersion: "1"
  installCacheVersion: "1"
  beforeSetupCacheVersion: "1"
  setupCacheVersion: "1"
mount:
- from: tmp_dir
  to: /tmp
- from: build_dir
  to: /root/.cache
- fromPath: ~/.npm
  to: /root/.npm
import:
- from: builder
  before: install
  add: /build/app
  to: /usr/local/bin/app
  owner: root
  group: root
  includePaths: [bin]
  excludePaths: [bin/test]
- from: nginx:1.25
  after: setup
  add: /etc/nginx
dependencies:
- from: builder
  after: install
  imports:
  - type: ImageDigest
    targetEnv: BUILDER_DIGEST
secrets:
- env: NPM_TOKEN
imageSpec:
  config:
    cmd: [/usr/local/bin/app]
`),
		Entry("stapel image with deprecated directives", `
image: app
fromImage: base
import:
- image: builder
  before: install
  add: /app
dependencies:
- image: builder
  before: setup
imageSpec:
  config:
    clearWerfLabels: true
`),
	)

	DescribeTable("rejects a document",
		func(document string) {
			Expect(schema.Validate(yamlDocument(document))).NotTo(Succeed())
		},
		Entry("meta without project", `
configVersion: 1
`),
		Entry("meta with unsupported configVersion", `
configVersion: 2
project: app
`),
		Entry("meta with unknown directive", `
configVersion: 1
project: app
unknown: true
`),
		Entry("meta with unknown nested directive", `
configVersion: 1
project: app
deploy:
  helmChart: .helm
`),
		Entry("cleanup keep policy without references", `
configVersion: 1
project: app
cleanup:
  keepPolicies:
  - imagesPerReference:
      last: 1
`),
		Entry("cleanup keep policy with both branch and tag", `
configVersion: 1
project: app
cleanup:
  keepPolicies:
  - references:
      branch: main
      tag: /v.*/
`),
		Entry("cleanup limit with unknown operator", `
configVersion: 1
project: app
cleanup:
  keepPolicies:
  - references:
      branch: main
      limit:
        operator: Xor
`),
		Entry("document that is neither meta nor image", `
from: alpine
`),
		Entry("dockerfile image without image name", `
dockerfile: Dockerfile
`),
		Entry("dockerfile image with empty image list", `
image: []
dockerfile: Dockerfile
`),
		Entry("dockerfile image with stapel directive", `
image: app
dockerfile: Dockerfile
shell:
  install: ls
`),
		Entry("dockerfile dependency with before", `
image: app
dockerfile: Dockerfile
dependencies:
- from: base
  before: install
`),
		Entry("dockerfile dependency import with targetEnv", `
image: app
dockerfile: Dockerfile
dependencies:
- from: base
  imports:
  - type: ImageName
    targetEnv: BASE
`),
		Entry("dependency import with unknown type", `
image: app
dockerfile: Dockerfile
dependencies:
- from: base
  imports:
  - type: ImageSize
    targetBuildArg: BASE
`),
		Entry("secret with two sources", `
image: app
dockerfile: Dockerfile
secrets:
- env: A
  src: ./a
`),
		Entry("secret value without id", `
image: app
dockerfile: Dockerfile
secrets:
- value: secret
`),
		Entry("stapel image without from", `
image: app
shell:
  install: ls
`),
		Entry("stapel image with both from and fromImage", `
image: app
from: alpine
fromImage: alpine
`),
		Entry("stapel image with dockerfile directive", `
image: app
from: alpine
context: .
`),
		Entry("local git mapping with branch", `
image: app
from: alpine
git:
- to: /app
  branch: main
`),
		Entry("git mapping without to", `
image: app
from: alpine
git:
- add: /src
`),
		Entry("git mapping password with two sources", `
image: app
from: alpine
git:
- url: https://example.com/repo.git
  to: /app
  basicAuth:
    password:
      env: A
      value: b
`),
		Entry("stage dependencies for unknown stage", `
image: app
from: alpine
git:
- to: /app
  stageDependencies:
    beforeInstall: ["*"]
`),
		Entry("shell with unknown stage", `
image: app
from: alpine
shell:
  afterSetup: ls
`),
		Entry("mount with unknown service directory", `
image: app
from: alpine
mount:
- from: home_dir
  to: /root
`),
		Entry("mount with both from and fromPath", `
image: app
from: alpine
mount:
- from: tmp_dir
  fromPath: /tmp
  to: /tmp
`),
		Entry("import without add", `
image: app
from: alpine
import:
- from: builder
  before: install
`),
		Entry("import without stage", `
image: app
from: alpine
import:
- from: builder
  add: /app
`),
		Entry("import with both before and after", `
image: app
from: alpine
import:
- from: builder
  before: install
  after: install
  add: /app
`),
		Entry("import with unknown stage", `
image: app
from: alpine
import:
- from: builder
  before: beforeInstall
  add: /app
`),
		Entry("stapel dependency without stage", `
image: app
from: alpine
dependencies:
- from: builder
`),
		Entry("stapel dependency import with targetBuildArg", `
image: app
from: alpine
dependencies:
- from: builder
  after: install
  imports:
  - type: ImageName
    targetBuildArg: BUILDER
`),
		Entry("imageSpec with unknown config directive", `
image: app
from: alpine
imageSpec:
  config:
    healthcheck:
      startPeriod: 10
`),
		Entry("global imageSpec with per-image directive", `
configVersion: 1
project: app
build:
  imageSpec:
    config:
      cmd: [run]
`),
	)
})
