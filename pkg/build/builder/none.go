package builder

type None struct{}

func NewNoneBuilder() *None {
	return &None{}
}

func (b *None) IsBeforeInstallEmpty() bool { return true }
func (b *None) IsInstallEmpty() bool       { return true }
func (b *None) IsBeforeSetupEmpty() bool   { return true }
func (b *None) IsSetupEmpty() bool         { return true }
func (b *None) IsBuildArtifactEmpty() bool { return true }

func (b *None) BeforeInstall(_ Container) error { return nil }
func (b *None) Install(_ Container) error       { return nil }
func (b *None) BeforeSetup(_ Container) error   { return nil }
func (b *None) Setup(_ Container) error         { return nil }
func (b *None) BuildArtifact(_ Container) error { return nil }

func (b *None) BeforeInstallChecksum() string { return "" }
func (b *None) InstallChecksum() string       { return "" }
func (b *None) BeforeSetupChecksum() string   { return "" }
func (b *None) SetupChecksum() string         { return "" }
func (b *None) BuildArtifactChecksum() string { return "" }
