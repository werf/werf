package image

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DSSE Envelope", func() {
	payload := []byte(`{"bomFormat":"CycloneDX","version":1}`)
	payloadType := "application/vnd.dsse.envelope.v1+json"

	DescribeTable("WrapInDSSE / UnwrapDSSE round-trip",
		func(payload []byte, payloadType string) {
			envelopeJSON, err := WrapInDSSE(payload, payloadType)
			Expect(err).To(Succeed())
			Expect(envelopeJSON).ToNot(BeEmpty())

			result, err := UnwrapDSSE(envelopeJSON, payloadType)
			Expect(err).To(Succeed())
			Expect(result).To(Equal(payload))
		},
		Entry("simple JSON payload", payload, payloadType),
		Entry("empty payload", []byte{}, payloadType),
		Entry("binary payload", []byte{0x00, 0x01, 0x02}, payloadType),
	)

	It("should fail on wrong payloadType", func() {
		envelopeJSON, err := WrapInDSSE(payload, payloadType)
		Expect(err).To(Succeed())

		_, err = UnwrapDSSE(envelopeJSON, "application/vnd.wrong+json")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unexpected DSSE payloadType"))
	})

	It("should fail on malformed JSON", func() {
		_, err := UnwrapDSSE([]byte("{bad json}"), payloadType)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("In-Toto Statement", func() {
	predicate := []byte(`{"bomFormat":"CycloneDX","version":1}`)
	predicateType := "https://cyclonedx.org/bom/v1.6"
	repo := "registry.example.com/project"
	digestHex := "abc123def456"

	DescribeTable("WrapInInTotoStatement / UnwrapInTotoStatement round-trip",
		func(predicate []byte, predicateType, repo, digestHex string) {
			stmtJSON, err := WrapInInTotoStatement(predicate, predicateType, repo, digestHex)
			Expect(err).To(Succeed())
			Expect(stmtJSON).ToNot(BeEmpty())

			resultPredicate, resultType, err := UnwrapInTotoStatement(stmtJSON)
			Expect(err).To(Succeed())
			Expect(resultType).To(Equal(predicateType))
			Expect(json.RawMessage(resultPredicate)).To(MatchJSON(predicate))
		},
		Entry("CycloneDX predicate", predicate, predicateType, repo, digestHex),
		Entry("different digest", []byte(`{"key":"val"}`), predicateType, "other.io/repo", "xyz789"),
		Entry("empty predicate", []byte(`{}`), predicateType, repo, digestHex),
	)

	It("should produce valid in-toto v1 statement structure", func() {
		stmtJSON, err := WrapInInTotoStatement(predicate, predicateType, repo, digestHex)
		Expect(err).To(Succeed())

		var stmt map[string]interface{}
		Expect(json.Unmarshal(stmtJSON, &stmt)).To(Succeed())

		Expect(stmt["_type"]).To(Equal("https://in-toto.io/Statement/v1"))
		Expect(stmt["predicateType"]).To(Equal(predicateType))

		subjects, ok := stmt["subject"].([]interface{})
		Expect(ok).To(BeTrue())
		Expect(subjects).To(HaveLen(1))

		subj, ok := subjects[0].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(subj["name"]).To(Equal(repo))

		digests, ok := subj["digest"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(digests["sha256"]).To(Equal(digestHex))
	})

	It("should fail on unknown statement type", func() {
		stmtJSON, err := WrapInInTotoStatement(predicate, predicateType, repo, digestHex)
		Expect(err).To(Succeed())

		// Replace _type with wrong value
		var raw map[string]json.RawMessage
		Expect(json.Unmarshal(stmtJSON, &raw)).To(Succeed())
		raw["_type"] = json.RawMessage(`"https://in-toto.io/Statement/v0.1"`)
		modified, err := json.Marshal(raw)
		Expect(err).To(Succeed())

		_, _, err = UnwrapInTotoStatement(modified)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unexpected in-toto statement type"))
	})
})
