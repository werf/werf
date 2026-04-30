module github.com/werf/werf/v2

go 1.25.5

toolchain go1.25.8

require (
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/alessio/shellescape v1.4.2
	github.com/aws/aws-sdk-go-v2/config v1.32.12
	github.com/aws/aws-sdk-go-v2/service/ecr v1.36.2
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/containerd/containerd v1.7.28
	github.com/containers/buildah v1.43.0
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/deislabs/oras v1.1.0
	github.com/djherbis/buffer v1.2.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/docker/buildx v0.33.0
	github.com/docker/cli v29.3.1+incompatible
	github.com/docker/docker v28.5.2+incompatible
	github.com/docker/go-connections v0.6.0
	github.com/docker/go-units v0.5.0
	github.com/dustin/go-humanize v1.0.1
	github.com/go-git/go-billy/v5 v5.6.0
	github.com/go-git/go-git/v5 v5.12.0
	github.com/go-openapi/spec v0.22.4
	github.com/go-openapi/strfmt v0.26.1
	github.com/go-openapi/validate v0.25.2
	github.com/google/go-containerregistry v0.20.7
	github.com/google/uuid v1.6.0
	github.com/gookit/color v1.6.0
	github.com/gosuri/uitable v0.0.4
	github.com/goware/urlx v0.3.2
	github.com/hofstadter-io/cinful v1.0.0
	github.com/mitchellh/copystructure v1.2.0
	github.com/moby/buildkit v0.29.0
	github.com/moby/moby/client v0.3.0
	github.com/moby/patternmatcher v0.6.1
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.39.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/otiai10/copy v1.14.0
	github.com/pkg/errors v0.9.1
	github.com/prashantv/gostub v1.1.0
	github.com/prometheus/procfs v0.20.1
	github.com/rodaine/table v1.1.1
	github.com/samber/lo v1.53.0
	github.com/sirupsen/logrus v1.9.4
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/werf/common-go v0.0.0-20260421143942-0695d8576092
	github.com/werf/copy-recurse v0.2.7
	github.com/werf/lockgate v0.1.1
	github.com/werf/logboek v0.6.1
	github.com/werf/nelm v1.23.3-0.20260429183632-3a2264e5ca5b
	go.opentelemetry.io/otel v1.42.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.42.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.42.0
	go.opentelemetry.io/otel/sdk v1.42.0
	go.opentelemetry.io/otel/trace v1.42.0
	go.podman.io/common v0.67.0
	go.podman.io/image/v5 v5.39.1
	go.podman.io/storage v1.62.1-0.20260407174525-18bc25cc910c
	go.uber.org/mock v0.5.0
	golang.org/x/crypto v0.49.0
	golang.org/x/net v0.52.0
	golang.org/x/sys v0.42.0
	gopkg.in/ini.v1 v1.67.1
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.35.3
	k8s.io/apimachinery v0.35.3
	k8s.io/cli-runtime v0.35.3
	k8s.io/client-go v0.35.3
	k8s.io/component-base v0.35.3
	k8s.io/kubectl v0.35.3
	k8s.io/utils v0.0.0-20260319190234-28399d86e0b5
	mvdan.cc/xurls v1.1.0
	sigs.k8s.io/yaml v1.6.0
)

require (
	cyphar.com/go-pathrs v0.2.4 // indirect
	github.com/Ladicle/tabwriter v1.0.0 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/alecthomas/chroma/v2 v2.23.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.8 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/chainguard-dev/git-urls v1.0.2 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/containerd/containerd/api v1.10.0 // indirect
	github.com/containerd/containerd/v2 v2.2.2 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/platforms v1.0.0-rc.2 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/digitorus/pkcs7 v0.0.0-20230818184609-3a137a874352 // indirect
	github.com/digitorus/timestamp v0.0.0-20231217203849-220c5c2851b7 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/go-events v0.0.0-20250808211157-605354379745 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fluxcd/cli-utils v0.37.2-flux.1 // indirect
	github.com/fluxcd/flagger v1.36.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-openapi/runtime v0.29.2 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.5 // indirect
	github.com/go-openapi/swag/conv v0.25.5 // indirect
	github.com/go-openapi/swag/fileutils v0.25.5 // indirect
	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.5 // indirect
	github.com/go-openapi/swag/loading v0.25.5 // indirect
	github.com/go-openapi/swag/mangling v0.25.5 // indirect
	github.com/go-openapi/swag/netutils v0.25.5 // indirect
	github.com/go-openapi/swag/stringutils v0.25.5 // indirect
	github.com/go-openapi/swag/typeutils v0.25.5 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.5 // indirect
	github.com/go-resty/resty/v2 v2.17.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-task/template v0.1.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/certificate-transparency-go v1.3.2 // indirect
	github.com/google/go-dap v0.12.1-0.20250904181021-d7a2259b058b // indirect
	github.com/gosimple/slug v1.15.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hiddeco/sshsig v0.2.0 // indirect
	github.com/in-toto/attestation v1.1.2 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/dsig v1.0.0 // indirect
	github.com/lestrrat-go/dsig-secp256k1 v1.0.0 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.1 // indirect
	github.com/lestrrat-go/jwx/v3 v3.0.11 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/mattn/go-zglob v0.0.6 // indirect
	github.com/miekg/dns v1.1.61 // indirect
	github.com/mistifyio/go-zfs/v4 v4.0.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/moby/go-archive v0.2.0 // indirect
	github.com/moby/moby/api v1.54.0 // indirect
	github.com/moby/policy-helpers v0.0.0-20260324161837-b7c0b994300b // indirect
	github.com/moby/swarmkit/v2 v2.1.1 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/sys/capability v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/ohler55/ojg v1.28.1 // indirect
	github.com/oklog/ulid/v2 v2.1.1 // indirect
	github.com/open-policy-agent/opa v1.10.1 // indirect
	github.com/opencontainers/cgroups v0.0.5 // indirect
	github.com/package-url/packageurl-go v0.1.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/sajari/fuzzy v1.0.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/sigstore/protobuf-specs v0.5.0 // indirect
	github.com/sigstore/rekor v1.5.0 // indirect
	github.com/sigstore/rekor-tiles/v2 v2.0.1 // indirect
	github.com/sigstore/sigstore-go v1.1.4 // indirect
	github.com/sigstore/timestamp-authority/v2 v2.0.3 // indirect
	github.com/smallstep/pkcs7 v0.1.1 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/theupdateframework/go-tuf/v2 v2.4.1 // indirect
	github.com/tonistiigi/dchapes-mode v0.0.0-20250318174251-73d941a28323 // indirect
	github.com/tonistiigi/go-csvvalue v0.0.0-20240814133006-030d3b2625d0 // indirect
	github.com/tonistiigi/go-rosetta v0.0.0-20220804170347-3f4430f2d346 // indirect
	github.com/tonistiigi/jaeger-ui-rest v0.0.0-20250408171107-3dd17559e117 // indirect
	github.com/transparency-dev/formats v0.0.0-20251017110053-404c0d5b696c // indirect
	github.com/transparency-dev/merkle v0.0.2 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/vektah/gqlparser/v2 v2.5.30 // indirect
	github.com/werf/kubedog v0.13.1-0.20260320165832-7d97aaf7aab9 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	github.com/yannh/kubeconform v0.7.0 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.16 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.34.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	mvdan.cc/sh/v3 v3.10.0 // indirect
	oras.land/oras-go/v2 v2.6.0 // indirect
	sigs.k8s.io/controller-runtime v0.23.3 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
	tags.cncf.io/container-device-interface/specs-go v1.1.0 // indirect
)

require (
	dario.cat/mergo v1.0.2 // indirect
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/BurntSushi/toml v1.6.0
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.14.0-rc.1 // indirect
	github.com/ProtonMail/go-crypto v1.4.1 // indirect
	github.com/Shopify/logrus-bugsnag v0.0.0-20230117174420-439a4b8ba167 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/aead/serpent v0.0.0-20160714141033-fba169763ea6 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/avelino/slugify v0.0.0-20180501145920-855f152bd774 // indirect
	github.com/aws/aws-sdk-go-v2 v1.41.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.12 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.20 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.20 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.20 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.20 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.9 // indirect
	github.com/aws/smithy-go v1.24.2 // indirect
	github.com/aymanbagabas/go-udiff v0.4.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.9.1
	github.com/bugsnag/bugsnag-go v2.2.0+incompatible // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.3 // indirect
	github.com/chanced/caps v1.0.2 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cloudflare/circl v1.6.3 // indirect
	github.com/compose-spec/compose-go/v2 v2.9.1 // indirect
	github.com/containerd/console v1.0.5 // indirect
	github.com/containerd/continuity v0.4.5 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.18.2 // indirect
	github.com/containerd/ttrpc v1.2.8 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/containernetworking/cni v1.3.0 // indirect
	github.com/containernetworking/plugins v1.9.1 // indirect
	github.com/containers/libtrust v0.0.0-20230121012942-c1716e8a8d01 // indirect
	github.com/containers/luksy v0.0.0-20250910190358-2cf5bc928957 // indirect
	github.com/containers/ocicrypt v1.2.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20241213102144-19d51d7fe467 // indirect
	github.com/cyphar/filepath-securejoin v0.6.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/disiqueira/gotree/v3 v3.0.2 // indirect
	github.com/distribution/reference v0.6.0
	github.com/docker/cli-docs-tool v0.11.0 // indirect
	github.com/docker/docker-credential-helpers v0.9.5 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/dominikbraun/graph v0.23.0 // indirect
	github.com/elazarl/goproxy v0.0.0-20231117061959-7cc037d33fb5 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fsouza/go-dockerclient v1.12.4 // indirect
	github.com/fvbommel/sortorder v1.1.0 // indirect
	github.com/garyburd/redigo v1.6.4 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.24.3 // indirect
	github.com/go-openapi/errors v0.22.7 // indirect
	github.com/go-openapi/jsonpointer v0.22.5 // indirect
	github.com/go-openapi/jsonreference v0.21.5 // indirect
	github.com/go-openapi/loads v0.23.3 // indirect
	github.com/go-openapi/swag v0.25.5 // indirect
	github.com/go-task/task/v3 v3.40.1
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-intervals v0.0.2 // indirect
	github.com/google/pprof v0.0.0-20260302011040-a15ffb7f9dcc // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty-funcs v0.0.0-20250818135842-6aab67130928 // indirect
	github.com/hashicorp/hcl/v2 v2.24.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/in-toto/in-toto-golang v0.10.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jedib0t/go-pretty/v6 v6.7.8 // indirect
	github.com/jellydator/ttlcache/v3 v3.4.0 // indirect
	github.com/jinzhu/copier v0.4.0 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.12.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/looplab/fsm v1.0.3 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.21 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/signal v0.7.1 // indirect
	github.com/moby/sys/symlink v0.3.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/morikuni/aec v1.1.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mvdan/xurls v1.1.0 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/oleiade/reflections v1.0.1 // indirect
	github.com/opencontainers/runc v1.3.4 // indirect
	github.com/opencontainers/runtime-tools v0.9.1-0.20251114084447-edf4cb3d2116 // indirect
	github.com/opencontainers/selinux v1.13.1 // indirect
	github.com/openshift/imagebuilder v1.2.20 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/proglottis/gpgme v0.1.5 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/rubenv/sql-migrate v1.8.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/seccomp/libseccomp-golang v0.11.1 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.10.0 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sigstore/fulcio v1.7.1 // indirect
	github.com/sigstore/sigstore v1.10.4 // indirect
	github.com/skeema/knownhosts v1.3.2 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/stefanberger/go-pkcs11uri v0.0.0-20230803200340-78284954bff6 // indirect
	github.com/stretchr/testify v1.11.1
	github.com/sylabs/sif/v2 v2.22.0 // indirect
	github.com/tchap/go-patricia/v2 v2.3.3 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20251211185533-a2aa163d723f // indirect
	github.com/tonistiigi/units v0.0.0-20180711220420-6950e57a87ea // indirect
	github.com/tonistiigi/vt100 v0.0.0-20240514184818-90bafcd6abab // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/vbatts/tar-split v0.12.2 // indirect
	github.com/vbauerster/mpb/v8 v8.10.2 // indirect
	github.com/vishvananda/netlink v1.3.1 // indirect
	github.com/vishvananda/netns v0.0.5 // indirect
	github.com/wI2L/jsondiff v0.7.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940 // indirect
	github.com/yvasiyarov/gorelic v0.0.7 // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	github.com/zclconf/go-cty v1.17.0 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.42.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.42.0 // indirect
	go.opentelemetry.io/otel/metric v1.42.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.42.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.20.0
	golang.org/x/term v0.41.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/grpc v1.79.3 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/apiextensions-apiserver v0.35.3 // indirect
	k8s.io/component-helpers v0.35.3 // indirect
	k8s.io/kube-openapi v0.0.0-20260319004828-5883c5ee87b9 // indirect
	k8s.io/metrics v0.35.3 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kustomize/v5 v5.7.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
	tags.cncf.io/container-device-interface v1.1.0 // indirect
)

replace (
	github.com/deislabs/oras => github.com/werf/3p-oras v0.9.1-0.20260408144000-3b8c77eb09e8 // used by bundles, not maintained
	github.com/spf13/cobra => github.com/werf/3p-cobra v0.0.0-20260403075225-552c82797324 // adds EnableErrorOnUnknownSubcommand, not yet in upstream
	oras.land/oras-go => github.com/werf/3p-oras-go v1.2.8-0.20260408140625-72dd516ce0aa // used by bundles, not maintained
)
