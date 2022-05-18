module github.com/werf/werf

go 1.17

require (
	bou.ke/monkey v1.0.1
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/alessio/shellescape v0.0.0-20190409004728-b115ca0f9053
	github.com/aws/aws-sdk-go v1.31.6
	github.com/bmatcuk/doublestar v1.1.5
	github.com/docker/cli v0.0.0-20200803101120-f4f962292d47
	github.com/docker/docker v17.12.0-ce-rc1.0.20200728121027-0f41a77c6993+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git/v5 v5.1.1-0.20200721083337-cded5b685b8a
	github.com/golang/example v0.0.0-20170904185048-46695d81d1fa
	github.com/google/go-containerregistry v0.1.1
	github.com/google/uuid v1.1.1
	github.com/gosuri/uitable v0.0.4
	github.com/hashicorp/go-version v1.2.0
	github.com/helm/helm-2to3 v0.8.1
	github.com/moby/buildkit v0.7.1-0.20200615045306-df35e9818d1f
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/otiai10/copy v1.0.1
	github.com/prashantv/gostub v1.0.0
	github.com/rodaine/table v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spaolacci/murmur3 v1.1.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/werf/kubedog v0.4.1-0.20220222103049-afc67463c609
	github.com/werf/lockgate v0.0.0-20200729113342-ec2c142f71ea
	github.com/werf/logboek v0.4.7
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	google.golang.org/grpc v1.29.1
	gopkg.in/ini.v1 v1.56.0
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm/v3 v3.5.1
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/cli-runtime v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/helm v2.17.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.18.6
	mvdan.cc/xurls v1.1.0
	sigs.k8s.io/yaml v1.2.0
)

require (
	cloud.google.com/go v0.57.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest/autorest v0.10.2 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.8.3 // indirect
	github.com/Azure/go-autorest/autorest/date v0.2.0 // indirect
	github.com/Azure/go-autorest/logger v0.1.0 // indirect
	github.com/Azure/go-autorest/tracing v0.5.0 // indirect
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/squirrel v1.2.0 // indirect
	github.com/Masterminds/vcs v1.13.1 // indirect
	github.com/Microsoft/go-winio v0.4.15-0.20190919025122-fc70bd9a86b5 // indirect
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535 // indirect
	github.com/avelino/slugify v0.0.0-20180501145920-855f152bd774 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/chai2010/gettext-go v0.0.0-20160711120539-c6fed771bfd5 // indirect
	github.com/cloudflare/cfssl v1.4.1 // indirect
	github.com/containerd/cgroups v0.0.0-20200327175542-b44481373989 // indirect
	github.com/containerd/console v1.0.0 // indirect
	github.com/containerd/containerd v1.4.0-0 // indirect
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deislabs/oras v0.8.1 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-metrics v0.0.1 // indirect
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96 // indirect
	github.com/emicklei/go-restful v2.13.0+incompatible // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/flant/logboek v0.3.1 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.3 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gofrs/flock v0.7.1 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/gogo/googleapis v1.4.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/shlex v0.0.0-20150127133951-6f45313302b9 // indirect
	github.com/googleapis/gnostic v0.2.2 // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jaguilar/vt100 v0.0.0-20150826170717-2703a27b14ea // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/gorm v1.9.12 // indirect
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/jmoiron/sqlx v1.2.1-0.20190826204134-d7d95172beb5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.3 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lib/pq v1.3.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mailru/easyjson v0.7.2 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.3.1 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/moby/sys/mount v0.1.0 // indirect
	github.com/moby/sys/mountinfo v0.1.3 // indirect
	github.com/moby/term v0.0.0-20200611042045-63b9a826fb74 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/oleiade/reflections v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/opencontainers/runc v1.0.0-rc8.0.20190926000215-3e425f80a8c9 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/otiai10/mint v1.3.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.3.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.7.0 // indirect
	github.com/prometheus/procfs v0.0.8 // indirect
	github.com/rubenv/sql-migrate v0.0.0-20200616145509-8d140a17f351 // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20200724193237-c3ed55f3b481 // indirect
	github.com/tonistiigi/go-rosetta v0.0.0-20200727161949-f79598599c5d // indirect
	github.com/tonistiigi/units v0.0.0-20180711220420-6950e57a87ea // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208 // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	google.golang.org/genproto v0.0.0-20200527145253-8367513e4ece // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/gorp.v1 v1.7.2 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.3.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	k8s.io/apiextensions-apiserver v0.18.6 // indirect
	k8s.io/apiserver v0.18.6 // indirect
	k8s.io/component-base v0.18.6 // indirect
	k8s.io/klog/v2 v2.0.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6 // indirect
	k8s.io/kubernetes v1.18.6 // indirect
	k8s.io/utils v0.0.0-20200731180307-f00132d28269 // indirect
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)

replace k8s.io/api => k8s.io/api v0.18.6

replace k8s.io/apimachinery => k8s.io/apimachinery v0.18.6

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.6

replace k8s.io/client-go => k8s.io/client-go v0.18.6

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.6

replace k8s.io/apiserver => k8s.io/apiserver v0.18.6

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.6

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.6

replace k8s.io/code-generator => k8s.io/code-generator v0.16.8-beta.0

replace k8s.io/component-base => k8s.io/component-base v0.18.6

replace k8s.io/cri-api => k8s.io/cri-api v0.18.6

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.6

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.6

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.6

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.6

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.6

replace k8s.io/kubelet => k8s.io/kubelet v0.18.6

replace k8s.io/kubectl => k8s.io/kubectl v0.18.6

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.6

replace k8s.io/metrics => k8s.io/metrics v0.18.6

replace k8s.io/node-api => k8s.io/node-api v0.18.6

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.6

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.18.6

replace k8s.io/sample-controller => k8s.io/sample-controller v0.18.6

replace github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200728121027-0f41a77c6993+incompatible

replace github.com/containerd/containerd => github.com/containerd/containerd v1.3.1-0.20200508210449-c80284d4b529

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace github.com/google/go-containerregistry => github.com/alexey-igrychev/go-containerregistry v0.1.3-0.20200901133051-a73cc6cd741c

replace k8s.io/helm => github.com/werf/helm v0.0.0-20210202092607-295e772eabb1

replace helm.sh/helm/v3 => github.com/werf/helm/v3 v3.0.0-20210218074430-84227989c996
