package signing

type SigningOptions struct {
	KeyRef   string
	CertRef  string
	ChainRef string
}

type ELFSigningOptions struct {
	Enabled                  bool
	PGPPrivateKeyFingerprint string
	PGPPrivateKeyPassphrase  string
}
