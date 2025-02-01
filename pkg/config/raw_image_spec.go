package config

type rawImageSpec struct {
	Author       string              `yaml:"author,omitempty"`
	ClearHistory bool                `yaml:"clearHistory,omitempty"`
	Config       *rawImageSpecConfig `yaml:"config,omitempty"`

	rawStapelImage         *rawStapelImage         `yaml:"-"` // possible parent
	rawImageFromDockerfile *rawImageFromDockerfile `yaml:"-"` // possible parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type rawImageSpecConfig struct {
	ClearWerfLabels bool     `yaml:"clearWerfLabels,omitempty"`
	RemoveLabels    []string `yaml:"removeLabels,omitempty"`
	RemoveVolumes   []string `yaml:"removeVolumes,omitempty"`
	RemoveEnv       []string `yaml:"removeEnv,omitempty"`
	ClearCmd        bool     `yaml:"clearCmd,omitempty"`
	ClearEntrypoint bool     `yaml:"clearEntrypoint,omitempty"`

	Volumes     []string          `yaml:"volumes,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Expose      []string          `yaml:"expose,omitempty"`
	User        string            `yaml:"user,omitempty"`
	Cmd         []string          `yaml:"cmd,omitempty"`
	Entrypoint  []string          `yaml:"entrypoint,omitempty"`
	WorkingDir  string            `yaml:"workingDir,omitempty"`
	StopSignal  string            `yaml:"stopSignal,omitempty"`
	Healthcheck *healthConfig     `yaml:"healthcheck,omitempty"`

	rawImageSpec *rawImageSpec `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type healthConfig struct {
	Test        []string `json:"test,omitempty"`
	Interval    int      `json:"interval,omitempty"`
	Timeout     int      `json:"timeout,omitempty"`
	StartPeriod int      `json:"startPeriod,omitempty"`
	Retries     int      `json:"retries,omitempty"`
}

func (s *rawImageSpec) doc() *doc {
	if s.rawStapelImage != nil {
		return s.rawStapelImage.doc
	} else if s.rawImageFromDockerfile != nil {
		return s.rawImageFromDockerfile.doc
	} else {
		panic("runtime error")
	}
}

func (s *rawImageSpec) UnmarshalYAML(unmarshal func(interface{}) error) error {
	switch parent := parentStack.Peek().(type) {
	case *rawStapelImage:
		s.rawStapelImage = parent
	case *rawImageFromDockerfile:
		s.rawImageFromDockerfile = parent
	}

	parentStack.Push(s)
	type plain rawImageSpec
	err := unmarshal((*plain)(s))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(s.UnsupportedAttributes, s, s.doc()); err != nil {
		return err
	}

	return nil
}

func (s *rawImageSpecConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawImageSpec); ok {
		s.rawImageSpec = parent
	}

	parentStack.Push(s)
	type plain rawImageSpecConfig
	err := unmarshal((*plain)(s))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(s.UnsupportedAttributes, nil, s.rawImageSpec.doc()); err != nil {
		return err
	}

	return nil
}

func (s *rawImageSpec) toDirective() *ImageSpec {
	imageSpec := &ImageSpec{raw: s}

	imageSpec.Author = s.Author
	imageSpec.ClearHistory = s.ClearHistory

	if s.Config != nil {
		imageSpec.ClearWerfLabels = s.Config.ClearWerfLabels
		imageSpec.RemoveLabels = s.Config.RemoveLabels
		imageSpec.RemoveVolumes = s.Config.RemoveVolumes
		imageSpec.RemoveEnv = s.Config.RemoveEnv
		imageSpec.ClearCmd = s.Config.ClearCmd
		imageSpec.ClearEntrypoint = s.Config.ClearEntrypoint

		imageSpec.Volumes = s.Config.Volumes
		imageSpec.Labels = s.Config.Labels
		imageSpec.User = s.Config.User
		imageSpec.Cmd = s.Config.Cmd
		imageSpec.Env = s.Config.Env
		imageSpec.Entrypoint = s.Config.Entrypoint
		imageSpec.WorkingDir = s.Config.WorkingDir
		imageSpec.StopSignal = s.Config.StopSignal
		imageSpec.Expose = s.Config.Expose

		imageSpec.Healthcheck = s.Config.Healthcheck
	}

	return imageSpec
}

type rawImageSpecGlobal struct {
	Author       string                    `yaml:"author,omitempty"`
	ClearHistory bool                      `yaml:"clearHistory,omitempty"`
	Config       *rawImageSpecGlobalConfig `yaml:"config,omitempty"`

	rawMetaBuild *rawMetaBuild `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type rawImageSpecGlobalConfig struct {
	ClearWerfLabels bool              `yaml:"clearWerfLabels,omitempty"`
	RemoveLabels    []string          `yaml:"removeLabels,omitempty"`
	Labels          map[string]string `yaml:"labels,omitempty"`

	rawImageSpecGlobal *rawImageSpecGlobal `yaml:"-"` // parent

	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

func (s *rawImageSpecGlobal) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMetaBuild); ok {
		s.rawMetaBuild = parent
	}

	parentStack.Push(s)
	type plain rawImageSpecGlobal
	err := unmarshal((*plain)(s))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(s.UnsupportedAttributes, nil, s.rawMetaBuild.rawMeta.doc); err != nil {
		return err
	}

	return nil
}

func (s *rawImageSpecGlobalConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawImageSpecGlobal); ok {
		s.rawImageSpecGlobal = parent
	}

	parentStack.Push(s)
	type plain rawImageSpecGlobalConfig
	err := unmarshal((*plain)(s))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(s.UnsupportedAttributes, nil, s.rawImageSpecGlobal.rawMetaBuild.rawMeta.doc); err != nil {
		return err
	}

	return nil
}

func (s *rawImageSpecGlobal) toDirective() *ImageSpec {
	imageSpec := &ImageSpec{rawGlobal: s}
	imageSpec.Author = s.Author
	imageSpec.ClearHistory = s.ClearHistory

	if s.Config != nil {
		imageSpec.ClearWerfLabels = s.Config.ClearWerfLabels
		imageSpec.RemoveLabels = s.Config.RemoveLabels
		imageSpec.Labels = s.Config.Labels
	}

	return imageSpec
}
