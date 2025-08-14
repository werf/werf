package signing

type ManifestSigningOptions struct {
	Enabled bool

	signer *Signer
}

func (o ManifestSigningOptions) Signer() *Signer {
	return o.signer
}

func NewManifestSigningOptions(signer *Signer) ManifestSigningOptions {
	return ManifestSigningOptions{
		signer: signer,
	}
}
