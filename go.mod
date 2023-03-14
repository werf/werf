module github.com/werf/werf

go 1.20

require (
	bou.ke/monkey v1.0.1
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/alessio/shellescape v0.0.0-20190409004728-b115ca0f9053
	github.com/aws/aws-sdk-go v1.43.16
	github.com/bmatcuk/doublestar v1.1.5
	github.com/containerd/containerd v1.6.15
	github.com/containers/buildah v1.26.1
	github.com/containers/common v0.48.0
	github.com/containers/image/v5 v5.21.1
	github.com/containers/storage v1.40.2
	github.com/deislabs/oras v0.12.0
	github.com/djherbis/buffer v1.1.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/docker/cli v20.10.21+incompatible
	github.com/docker/distribution v2.8.1+incompatible
	github.com/docker/docker v20.10.21+incompatible
	github.com/docker/go-connections v0.4.1-0.20210727194412-58542c764a11
	github.com/docker/go-units v0.4.0
	github.com/dustin/go-humanize v1.0.0
	github.com/fluxcd/flagger v1.8.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-git/go-git/v5 v5.4.2
	github.com/go-openapi/spec v0.19.5
	github.com/go-openapi/strfmt v0.19.5
	github.com/go-openapi/validate v0.19.8
	github.com/google/go-containerregistry v0.5.1
	github.com/google/uuid v1.3.0
	github.com/gookit/color v1.3.7
	github.com/gophercloud/gophercloud v0.25.0
	github.com/gosuri/uitable v0.0.4
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.2.0
	github.com/helm/helm-2to3 v0.8.1
	github.com/minio/minio v0.0.0-20210311070216-f92b7a562103
	github.com/mitchellh/copystructure v1.2.0
	github.com/moby/buildkit v0.8.2
	github.com/moby/term v0.0.0-20221205130635-1aeaba878587
	github.com/onsi/ginkgo/v2 v2.9.1
	github.com/onsi/gomega v1.27.4
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0-rc2
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/otiai10/copy v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/prashantv/gostub v1.0.0
	github.com/rodaine/table v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.9.0
	github.com/spaolacci/murmur3 v1.1.0
	github.com/spf13/cobra v1.6.1
	github.com/spf13/pflag v1.0.5
	github.com/werf/copy-recurse v0.2.4
	github.com/werf/kubedog v0.9.9
	github.com/werf/lockgate v0.0.0-20200729113342-ec2c142f71ea
	github.com/werf/logboek v0.5.4
	go.opentelemetry.io/otel v1.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.7.1-0.20220624212736-ef6c0da0de3b
	go.opentelemetry.io/otel/sdk v1.10.0
	go.opentelemetry.io/otel/trace v1.10.0
	golang.org/x/crypto v0.5.0
	golang.org/x/mod v0.9.0
	golang.org/x/net v0.8.0
	gopkg.in/errgo.v2 v2.1.0
	gopkg.in/ini.v1 v1.66.2
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.26.2
	k8s.io/apimachinery v0.26.2
	k8s.io/cli-runtime v0.26.2
	k8s.io/client-go v0.26.2
	k8s.io/helm v2.17.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.80.1
	k8s.io/kubectl v0.26.2
	mvdan.cc/xurls v1.1.0
	oras.land/oras-go v1.2.2
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/semver/v3 v3.2.0 // indirect
	github.com/Masterminds/squirrel v1.5.3 // indirect
	github.com/Masterminds/vcs v1.13.3 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Microsoft/hcsshim v0.9.6 // indirect
	github.com/ProtonMail/go-crypto v0.0.0-20220407094043-a94812496cf5 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/acomagu/bufpipe v1.0.3 // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/avelino/slugify v0.0.0-20180501145920-855f152bd774 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/cenkalti/backoff/v4 v4.1.3 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/cloudflare/cfssl v1.4.1 // indirect
	github.com/containerd/cgroups v1.0.4 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.11.4 // indirect
	github.com/containerd/typeurl v1.0.2 // indirect
	github.com/containernetworking/cni v1.1.1 // indirect
	github.com/containernetworking/plugins v1.1.1 // indirect
	github.com/containers/libtrust v0.0.0-20200511145503-9c3a6c22cd9a // indirect
	github.com/containers/ocicrypt v1.1.4-0.20220428134531-566b808bdf6f // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/daviddengcn/go-colortext v1.0.0 // indirect
	github.com/disiqueira/gotree/v3 v3.0.2 // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/libnetwork v0.8.0-dev.2.0.20200917202933-d0951081b35f // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsouza/go-dockerclient v1.7.11 // indirect
	github.com/fvbommel/sortorder v1.0.1 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-gorp/gorp/v3 v3.0.2 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.19.5 // indirect
	github.com/go-openapi/errors v0.19.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/loads v0.19.4 // indirect
	github.com/go-openapi/runtime v0.19.4 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20210107165309-348f09dbbbc0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-intervals v0.0.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.0 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/huandu/xstrings v1.3.3 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/ishidawataru/sctp v0.0.0-20210226210310-f2269e66cdee // indirect
	github.com/jaguilar/vt100 v0.0.0-20150826170717-2703a27b14ea // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/jinzhu/gorm v1.9.12 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/klauspost/compress v1.15.2 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/manifoldco/promptui v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/mattn/go-shellwords v1.0.12 // indirect
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/moby/sys/mount v0.2.0 // indirect
	github.com/moby/sys/mountinfo v0.6.1 // indirect
	github.com/moby/sys/symlink v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/montanaflynn/stats v0.5.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mvdan/xurls v1.1.0 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/ncw/directio v1.0.5 // indirect
	github.com/oleiade/reflections v1.0.1 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/opencontainers/runtime-tools v0.9.0 // indirect
	github.com/opencontainers/selinux v1.10.1 // indirect
	github.com/openshift/imagebuilder v1.2.4-0.20220502172744-009dbc6cb805 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/ostreedev/ostree-go v0.0.0-20210805093236-719684c64e4f // indirect
	github.com/otiai10/mint v1.3.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/proglottis/gpgme v0.1.1 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/rubenv/sql-migrate v1.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/seccomp/libseccomp-golang v0.9.2-0.20210429002308-3879420cc921 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/stefanberger/go-pkcs11uri v0.0.0-20201008174630-78d3cae3a980 // indirect
	github.com/sylabs/sif/v2 v2.7.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/tchap/go-patricia v2.3.0+incompatible // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20201103201449-0834f99b7b85 // indirect
	github.com/tonistiigi/go-rosetta v0.0.0-20200727161949-f79598599c5d // indirect
	github.com/tonistiigi/units v0.0.0-20180711220420-6950e57a87ea // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/vbatts/tar-split v0.11.2 // indirect
	github.com/vbauerster/mpb/v7 v7.4.1 // indirect
	github.com/vishvananda/netlink v1.1.1-0.20210330154013-f5de75959ad5 // indirect
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f // indirect
	github.com/xanzy/ssh-agent v0.3.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/xlab/treeprint v1.1.0 // indirect
	go.etcd.io/bbolt v1.3.6 // indirect
	go.mongodb.org/mongo-driver v1.5.1 // indirect
	go.mozilla.org/pkcs7 v0.0.0-20200128120323-432b2356ecb1 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.10.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.starlark.net v0.0.0-20200306205701-8dd3e2ee1dd5 // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.org/x/tools v0.7.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220502173005-c8bf987b8c21 // indirect
	google.golang.org/grpc v1.49.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/apiextensions-apiserver v0.26.2 // indirect
	k8s.io/apiserver v0.26.2 // indirect
	k8s.io/component-base v0.26.2 // indirect
	k8s.io/component-helpers v0.26.2 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	k8s.io/metrics v0.26.2 // indirect
	k8s.io/utils v0.0.0-20221107191617-1a15be271d1d // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/kustomize/api v0.12.1 // indirect
	sigs.k8s.io/kustomize/kustomize/v4 v4.5.7 // indirect
	sigs.k8s.io/kustomize/kyaml v0.13.9 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace k8s.io/helm => github.com/werf/helm v0.0.0-20210202111118-81e74d46da0f

replace github.com/deislabs/oras => github.com/werf/third-party-oras v0.9.1-0.20210927171747-6d045506f4c8

replace github.com/go-git/go-git/v5 => github.com/ZauberNerd/go-git/v5 v5.4.3-0.20220315170230-29ec1bc1e5db

replace helm.sh/helm/v3 => github.com/werf/3p-helm/v3 v3.0.0-20230309133321-09d8a1e63ebd
