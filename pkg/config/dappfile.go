package config

type Dappfile struct {
	Meta  Meta
	Dimgs []*Dimg
}

type Meta struct {
	Project         string
	DeployTemplates DeployTemplates
}

type DeployTemplates struct {
	HelmRelease, KubernetesNamespace string
}
