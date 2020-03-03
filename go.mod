module github.com/flant/werf

require (
	cloud.google.com/go v0.38.0
	github.com/Masterminds/goutils v1.1.0
	github.com/Masterminds/semver v1.4.2
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/alessio/shellescape v0.0.0-20190409004728-b115ca0f9053
	github.com/apache/thrift v0.0.0-20161221203622-b2a4d4ae21c7 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/bmatcuk/doublestar v1.1.5
	github.com/codahale/hdrhistogram v0.0.0-20160425231609-f8ad88b59a58 // indirect
	github.com/containerd/cgroups v0.0.0-20181219155423-39b18af02c41 // indirect
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.0
	github.com/containerd/continuity v0.0.0-20190827140505-75bee3e2ccb6
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c
	github.com/containerd/go-runc v0.0.0-20180907222934-5a6d9f37cfa3
	github.com/containerd/typeurl v0.0.0-20190228175220-2a93cfde8c20
	github.com/coreos/go-systemd v0.0.0-20181031085051-9002847aa142 // indirect
	github.com/docker/cli v0.0.0-20191017083524-a8ff7f821017
	github.com/docker/compose-on-kubernetes v0.4.23 // indirect
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.4.2-0.20190924003213-a8608b5b67c7
	github.com/docker/docker-credential-helpers v0.6.3
	github.com/docker/go v1.5.1-1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82
	github.com/docker/go-units v0.4.0
	github.com/docker/libnetwork v0.0.0-20180913200009-36d3bed0e9f4
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/docker/licensing v0.0.0-20190320170819-9781369abdb5 // indirect
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96
	github.com/docker/swarmkit v0.0.0-20180705210007-199cf49cd996
	github.com/fatih/color v1.9.0
	github.com/flant/kubedog v0.3.5-0.20200228135326-83b69f5024b7
	github.com/flant/logboek v0.3.3
	github.com/flant/shluz v0.0.0-20191223174507-c6152b298d53
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/gofrs/flock v0.7.1
	github.com/google/btree v1.0.0
	github.com/google/go-cmp v0.3.0
	github.com/google/go-containerregistry v0.0.0-20200227193449-ba53fa10e72c
	github.com/google/gofuzz v1.0.0
	github.com/google/shlex v0.0.0-20150127133951-6f45313302b9
	github.com/google/uuid v1.1.1
	github.com/gosuri/uitable v0.0.0-20160404203958-36ee7e946282
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/uuid v0.0.0-20160311170451-ebb0a03e909c // indirect
	github.com/ishidawataru/sctp v0.0.0-20180213033435-07191f837fed // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e
	github.com/mattn/go-isatty v0.0.11
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/mitchellh/hashstructure v0.0.0-20170609045927-2bca23e0e452 // indirect
	github.com/mjibson/esc v0.2.0 // indirect
	github.com/moby/buildkit v0.3.3
	github.com/moby/moby v0.7.3-0.20190411110308-fc52433fa677
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.5.0
	github.com/opentracing-contrib/go-stdlib v0.0.0-20171029140428-b1a47cfbdd75 // indirect
	github.com/opentracing/opentracing-go v0.0.0-20171003133519-1361b9cd60be // indirect
	github.com/otiai10/copy v1.0.1
	github.com/pkg/profile v1.2.1 // indirect
	github.com/prashantv/gostub v1.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spaolacci/murmur3 v1.1.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/syndtr/gocapability v0.0.0-20170704070218-db04d3cc01c8 // indirect
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/tonistiigi/fsutil v0.0.0-20190130224639-b4281fa67095 // indirect
	github.com/tonistiigi/units v0.0.0-20180711220420-6950e57a87ea // indirect
	github.com/uber/jaeger-client-go v0.0.0-20180103221425-e02c85f9069e // indirect
	github.com/uber/jaeger-lib v1.2.1 // indirect
	github.com/urfave/cli v0.0.0-20171014202726-7bc6a0acffa5 // indirect
	github.com/vishvananda/netlink v1.0.0 // indirect
	github.com/vishvananda/netns v0.0.0-20180720170159-13995c7128cc // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190809123943-df4f5c81cb3b // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20170512152554-8a8cc2c7e54a
	golang.org/x/crypto v0.0.0-20200210222208-86ce3cb69678
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20200120151820-655fe14d7479
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	google.golang.org/genproto v0.0.0-20200117163144-32f20d992d24
	google.golang.org/grpc v1.26.0
	gopkg.in/ini.v1 v1.46.0
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/src-d/go-billy.v4 v4.3.2
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.17.0
	k8s.io/apiserver v0.16.7
	k8s.io/cli-runtime v0.16.7
	k8s.io/client-go v0.17.0
	k8s.io/cloud-provider v0.16.7
	k8s.io/helm v2.13.1+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	k8s.io/kubernetes v1.16.7
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1
	mvdan.cc/xurls v1.1.0
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.1.0
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

replace k8s.io/helm => github.com/flant/helm v0.0.0-20200302114220-d40ba0adc0f1

replace github.com/containerd/containerd => github.com/containerd/containerd v1.2.3

replace github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190319215453-e7b5f7dbe98c

replace github.com/docker/cli => github.com/docker/cli v0.0.0-20190321234815-f40f9c240ab0

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a

go 1.13
