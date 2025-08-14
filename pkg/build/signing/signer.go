package signing

import (
	"context"
	"fmt"

	"github.com/deckhouse/delivery-kit-sdk/pkg/signver"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
)

type SignerOptions struct {
	KeyRef           string
	CertRef          string
	IntermediatesRef string
}

func (o SignerOptions) IsZero() bool {
	return o.KeyRef == "" || o.CertRef == ""
}

type Signer struct {
	sv *signver.SignerVerifier
}

func (s *Signer) SignerVerifier() *signver.SignerVerifier {
	return s.sv
}

func (s *Signer) Cert() string {
	if s.sv == nil {
		return ""
	}
	return string(s.sv.Cert)
}

func (s *Signer) Chain() string {
	if s.sv == nil {
		return ""
	}
	return string(s.sv.Chain)
}

func NewSigner(ctx context.Context, opts SignerOptions) (*Signer, error) {
	if opts.IsZero() {
		return &Signer{nil}, nil
	}
	sv, err := signver.NewSignerVerifier(ctx, opts.CertRef, opts.IntermediatesRef, signver.KeyOpts{
		KeyRef:   opts.KeyRef,
		PassFunc: cryptoutils.SkipPassword,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create signer verifier: %w", err)
	}
	return &Signer{
		sv: sv,
	}, nil
}
