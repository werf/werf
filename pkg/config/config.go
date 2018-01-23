package config

type RubyType interface {
	GetRubyTypeTag() string
}


type Mount struct {
	_to string
	_from string
	_type string
}

func (cfg *Mount) GetRubyTypeTag() (string) {
	return “ruby/object:Dapp::Dimg::Config::Directive::Mount”
}


type Chef struct {
  _dimod []string
	_recipe []string
	_attributes []ChefAttributes
}

func (cfg *Chef) GetRubyTypeTag() (string) {
	return “ruby/object:Dapp::Dimg::Config::Directive::Mount”
}

type ChefAttributes map[interface{}]interface{}

func (cfg *ChefAttributes) GetRubyTypeTag() (string) {
	return “ruby/hash:Dapp::Dimg::Config::Directive::Chef::Attributes”
}


type ArtifactBase struct {
  Owner string `yaml:"_owner"`
	Group string `yaml:"_group"`
}

type Artifact struct {
  ArtifactBase
	Config string `yaml:"_config"`
}
