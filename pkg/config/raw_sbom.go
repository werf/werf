package config

type rawSbom bool

func (s *rawSbom) toDirective() *Sbom {
	return &Sbom{
		Use: true,
	}
}
