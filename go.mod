module github.com/werf/werf

go 1.14

require (
	bou.ke/monkey v1.0.1
	github.com/Masterminds/goutils v1.1.1
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/alessio/shellescape v0.0.0-20190409004728-b115ca0f9053
	github.com/aws/aws-sdk-go v1.37.32
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bmatcuk/doublestar v1.1.5
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/cloudflare/cfssl v1.4.1 // indirect
	github.com/containerd/containerd v1.5.10
	github.com/containers/buildah v1.24.2
	github.com/containers/common v0.47.4
	github.com/containers/image/v5 v5.19.1
	github.com/containers/storage v1.38.2
	github.com/deislabs/oras v0.12.0
	github.com/djherbis/buffer v1.1.0
	github.com/djherbis/nio/v3 v3.0.1
	github.com/docker/cli v20.10.11+incompatible
	github.com/docker/distribution v2.8.0+incompatible
	github.com/docker/docker v20.10.12+incompatible
	github.com/docker/go v1.5.1-1 // indirect
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
	github.com/gosuri/uitable v0.0.4
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-version v1.2.0
	github.com/helm/helm-2to3 v0.8.1
	github.com/jinzhu/gorm v1.9.12 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/minio/minio v0.0.0-20210311070216-f92b7a562103
	github.com/mitchellh/copystructure v1.2.0
	github.com/moby/buildkit v0.8.2
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6
	github.com/mvdan/xurls v1.1.0 // indirect
	github.com/oleiade/reflections v1.0.1 // indirect
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.0.0
	github.com/onsi/gomega v1.18.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.3-0.20211202193544-a5463b7f9c84
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/otiai10/copy v1.0.1
	github.com/otiai10/curr v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prashantv/gostub v1.0.0
	github.com/rodaine/table v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spaolacci/murmur3 v1.1.0
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/tonistiigi/go-rosetta v0.0.0-20200727161949-f79598599c5d // indirect
	github.com/werf/kubedog v0.6.4-0.20220222141823-4ca722ade0ef
	github.com/werf/lockgate v0.0.0-20200729113342-ec2c142f71ea
	github.com/werf/logboek v0.5.4
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	go.mongodb.org/mongo-driver v1.5.1 // indirect
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d
	golang.org/x/sys v0.0.0-20220403205710-6acee93ad0eb // indirect
	golang.org/x/tools v0.1.10 // indirect
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/errgo.v2 v2.1.0
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/ini.v1 v1.66.2
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/cli-runtime v0.23.1
	k8s.io/client-go v0.23.1
	k8s.io/helm v2.17.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.30.0
	k8s.io/kubectl v0.23.1
	mvdan.cc/xurls v1.1.0
	oras.land/oras-go v1.1.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.3.0
)

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

replace k8s.io/helm => github.com/werf/helm v0.0.0-20210202111118-81e74d46da0f

replace github.com/deislabs/oras => github.com/werf/third-party-oras v0.9.1-0.20210927171747-6d045506f4c8

replace helm.sh/helm/v3 => github.com/werf/3p-helm/v3 v3.0.0-20220323104627-79c3ac94b05f
