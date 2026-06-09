package image

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/in-toto/in-toto-golang/in_toto"
	"github.com/secure-systems-lab/go-securesystemslib/dsse"
)

type inTotoStatement struct {
	in_toto.StatementHeader
	Predicate json.RawMessage `json:"predicate"`
}

// WrapInDSSE wraps the payload into a DSSE envelope with the specified payloadType.
func WrapInDSSE(payload []byte, payloadType string) ([]byte, error) {
	envelope := dsse.Envelope{
		PayloadType: payloadType,
		Payload:     base64.StdEncoding.EncodeToString(payload),
		Signatures:  []dsse.Signature{},
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal DSSE envelope: %w", err)
	}

	return envelopeBytes, nil
}

// UnwrapDSSE unwraps the payload from a DSSE envelope and verifies the payloadType.
func UnwrapDSSE(envelopeJSON []byte, expectedPayloadType string) ([]byte, error) {
	var envelope dsse.Envelope
	if err := json.Unmarshal(envelopeJSON, &envelope); err != nil {
		return nil, fmt.Errorf("unmarshal DSSE envelope: %w", err)
	}

	if envelope.PayloadType != expectedPayloadType {
		return nil, fmt.Errorf("unexpected DSSE payloadType %q, expected %q", envelope.PayloadType, expectedPayloadType)
	}

	payload, err := envelope.DecodeB64Payload()
	if err != nil {
		return nil, fmt.Errorf("decode DSSE payload: %w", err)
	}

	return payload, nil
}

// WrapInInTotoStatement wraps the predicate into an in-toto v1 statement.
func WrapInInTotoStatement(predicate []byte, predicateType, repo, digestHex string) ([]byte, error) {
	stmt := inTotoStatement{
		StatementHeader: in_toto.StatementHeader{
			Type:          "https://in-toto.io/Statement/v1",
			PredicateType: predicateType,
			Subject: []in_toto.Subject{{
				Name:   repo,
				Digest: map[string]string{"sha256": digestHex},
			}},
		},
		Predicate: json.RawMessage(predicate),
	}

	stmtBytes, err := json.Marshal(stmt)
	if err != nil {
		return nil, fmt.Errorf("marshal in-toto statement: %w", err)
	}

	return stmtBytes, nil
}

// UnwrapInTotoStatement unwraps the predicate from an in-toto v1 statement and verifies the statement type.
func UnwrapInTotoStatement(statementJSON []byte) (json.RawMessage, string, error) {
	var stmt inTotoStatement
	if err := json.Unmarshal(statementJSON, &stmt); err != nil {
		return nil, "", fmt.Errorf("unmarshal in-toto statement: %w", err)
	}

	if stmt.Type != "https://in-toto.io/Statement/v1" {
		return nil, "", fmt.Errorf("unexpected in-toto statement type %q, expected %q", stmt.Type, "https://in-toto.io/Statement/v1")
	}

	return stmt.Predicate, stmt.PredicateType, nil
}
