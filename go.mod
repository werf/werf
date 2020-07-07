module github.com/werf/werf

require (
	bou.ke/monkey v1.0.1
	github.com/Masterminds/goutils v1.1.0
	github.com/Masterminds/semver v1.4.2
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d // indirect
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/alessio/shellescape v0.0.0-20190409004728-b115ca0f9053
	github.com/aws/aws-sdk-go v1.27.1
	github.com/bitly/go-hostpool v0.1.0 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bmatcuk/doublestar v1.1.5
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/bugsnag/bugsnag-go v1.5.3 // indirect
	github.com/bugsnag/panicwrap v1.2.0 // indirect
	github.com/containerd/cgroups v0.0.0-20181219155423-39b18af02c41 // indirect
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50 // indirect
	github.com/coreos/go-systemd v0.0.0-20181031085051-9002847aa142 // indirect
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/compose-on-kubernetes v0.4.23 // indirect
	github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7
	github.com/docker/go v1.5.1-1 // indirect
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/go-units v0.4.0
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/fatih/color v1.9.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git/v5 v5.1.0
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/gofrs/uuid v3.3.0+incompatible // indirect
	github.com/golang/example v0.0.0-20170904185048-46695d81d1fa
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/google/go-containerregistry v0.0.0-20200320200342-35f57d7d4930
	github.com/google/shlex v0.0.0-20150127133951-6f45313302b9 // indirect
	github.com/google/uuid v1.1.1
	github.com/gosuri/uitable v0.0.0-20160404203958-36ee7e946282
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/jinzhu/gorm v1.9.12 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.11
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/moby/buildkit v0.3.3
	github.com/oleiade/reflections v1.0.0 // indirect
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/otiai10/copy v1.0.1
	github.com/otiai10/curr v1.0.0 // indirect
	github.com/prashantv/gostub v1.0.0
	github.com/rodaine/table v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spaolacci/murmur3 v1.1.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/syndtr/gocapability v0.0.0-20170704070218-db04d3cc01c8 // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20190130224639-b4281fa67095 // indirect
	github.com/tonistiigi/units v0.0.0-20180711220420-6950e57a87ea // indirect
	github.com/werf/kubedog v0.3.5-0.20200707154239-8015c267710f
	github.com/werf/lockgate v0.0.0-20200625122100-41c30943229f
	github.com/werf/logboek v0.3.5-0.20200608145450-5b5f18fe7009
	github.com/xeipuuv/gojsonpointer v0.0.0-20190809123943-df4f5c81cb3b // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20170512152554-8a8cc2c7e54a // indirect
	github.com/ziutek/mymysql v1.5.4 // indirect
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	google.golang.org/genproto v0.0.0-20200117163144-32f20d992d24 // indirect
	google.golang.org/grpc v1.26.0
	gopkg.in/dancannon/gorethink.v3 v3.0.5 // indirect
	gopkg.in/fatih/pool.v2 v2.0.0 // indirect
	gopkg.in/gorethink/gorethink.v3 v3.0.5 // indirect
	gopkg.in/ini.v1 v1.46.0
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.4
	k8s.io/apiextensions-apiserver v0.16.6 // indirect
	k8s.io/apimachinery v0.17.4
	k8s.io/cli-runtime v0.16.7
	k8s.io/client-go v0.17.4
	k8s.io/helm v2.13.1+incompatible
	k8s.io/klog v1.0.0
	k8s.io/utils v0.0.0-20200327001022-6496210b90e8 // indirect
	mvdan.cc/xurls v1.1.0
	sigs.k8s.io/yaml v1.2.0
)

replace k8s.io/api => k8s.io/api v0.16.7

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.7

replace k8s.io/apimachinery => k8s.io/apimachinery v0.16.8-beta.0

replace k8s.io/apiserver => k8s.io/apiserver v0.16.7

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.7

replace k8s.io/client-go => k8s.io/client-go v0.16.7

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.7

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.7

replace k8s.io/code-generator => k8s.io/code-generator v0.16.8-beta.0

replace k8s.io/component-base => k8s.io/component-base v0.16.7

replace k8s.io/cri-api => k8s.io/cri-api v0.16.8-beta.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.7

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.7

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.7

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.7

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.7

replace k8s.io/kubectl => k8s.io/kubectl v0.16.7

replace k8s.io/kubelet => k8s.io/kubelet v0.16.7

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.7

replace k8s.io/metrics => k8s.io/metrics v0.16.7

replace k8s.io/node-api => k8s.io/node-api v0.16.7

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.7

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.16.7

replace k8s.io/sample-controller => k8s.io/sample-controller v0.16.7

replace k8s.io/helm => github.com/werf/helm v0.0.0-20200707170917-c558867a0a34

replace github.com/containerd/containerd => github.com/containerd/containerd v1.2.3

replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190319215453-e7b5f7dbe98c

replace github.com/docker/cli => github.com/docker/cli v0.0.0-20190321234815-f40f9c240ab0

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a

go 1.13

replace github.com/go-git/go-git/v5 => github.com/distorhead/go-git/v5 v5.1.1-0.20200611092215-45f2d6a6e110
