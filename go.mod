module github.com/flant/werf

require (
	cloud.google.com/go v0.34.0
	github.com/Masterminds/semver v1.4.2
	github.com/Masterminds/sprig v2.20.0+incompatible
	github.com/agl/ed25519 v0.0.0-20170116200512-5312a6153412 // indirect
	github.com/asaskevich/govalidator v0.0.0-20160518190739-766470278477
	github.com/bmatcuk/doublestar v1.1.5
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.0-0.20190212172151-f5b0fa220df8
	github.com/containerd/continuity v0.0.0-20181203112020-004b46473808
	github.com/containerd/fifo v0.0.0-20180307165137-3d5202aec260
	github.com/containerd/go-runc v0.0.0-20180907222934-5a6d9f37cfa3
	github.com/containerd/typeurl v0.0.0-20180627222232-a93fcdb778cd
	github.com/docker/cli v0.0.0-20190321234815-f40f9c240ab0
	github.com/docker/compose-on-kubernetes v0.4.23 // indirect
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/docker-credential-helpers v0.6.1
	github.com/docker/go v1.5.1-1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-events v0.0.0-20170721190031-9461782956ad
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82
	github.com/docker/go-units v0.3.3
	github.com/docker/libnetwork v0.0.0-20180913200009-36d3bed0e9f4
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/docker/licensing v0.0.0-20190320170819-9781369abdb5 // indirect
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96
	github.com/docker/swarmkit v0.0.0-20180705210007-199cf49cd996
	github.com/fatih/color v1.7.0
	github.com/flant/go-containerregistry v0.0.0-20190712094650-0cfc503dc51a
	github.com/flant/kubedog v0.3.4-0.20190821145513-2cd6191586fb
	github.com/flant/logboek v0.2.6-0.20190726104558-c32b60bb4a37
	github.com/flant/logboek_py v0.0.0-20190418220715-388556f27301
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/ghodss/yaml v0.0.0-20180820084758-c7ce16629ff4
	github.com/google/btree v1.0.0
	github.com/google/go-cmp v0.3.0
	github.com/google/go-containerregistry v0.0.0-20190623150931-ca8b66cb1b79
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf
	github.com/google/shlex v0.0.0-20150127133951-6f45313302b9
	github.com/google/uuid v1.0.0
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/gosuri/uitable v0.0.0-20160404203958-36ee7e946282
	github.com/hashicorp/go-version v1.2.0
	github.com/mattn/go-shellwords v1.0.5 // indirect
	github.com/moby/buildkit v0.4.0
	github.com/moby/moby v0.7.3-0.20190411110308-fc52433fa677
	github.com/otiai10/copy v1.0.1
	github.com/satori/go.uuid v1.2.0
	github.com/spaolacci/murmur3 v1.1.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/theupdateframework/notary v0.6.1 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190809123943-df4f5c81cb3b // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20170512152554-8a8cc2c7e54a
	golang.org/x/crypto v0.0.0-20190530122614-20be4c3c3ed5
	golang.org/x/net v0.0.0-20190603091049-60506f45cf65
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190602015325-4c4f7f33c9ed
	golang.org/x/text v0.3.2
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4
	gopkg.in/ini.v1 v1.46.0
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/src-d/go-billy.v4 v4.3.0 // indirect
	gopkg.in/src-d/go-git-fixtures.v3 v3.5.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.11.0
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190409092523-d687e77c8ae9
	k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery v0.0.0-20190409092423-760d1845f48b
	k8s.io/apiserver v0.0.0-20190313205120-8b27c41bdbb1
	k8s.io/cli-runtime v0.0.0-20190409093718-11d55751678d
	k8s.io/client-go v0.0.0-20190411052641-7a6b4715b709
	k8s.io/cloud-provider v0.0.0-20190323031113-9c9d72d1bf90
	k8s.io/helm v2.13.1+incompatible
	k8s.io/klog v0.2.0
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kubernetes v1.14.1
	k8s.io/utils v0.0.0-20190308190857-21c4ce38f2a7
	mvdan.cc/xurls v1.1.0
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.1.0
)

replace k8s.io/helm => github.com/flant/helm v0.0.0-20190816114256-f4d27cf06ce1
