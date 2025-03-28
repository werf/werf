package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	// Unlimited references by default.
	defaultReferencesLimitLast     = -1
	defaultReferencesLimitOperator = AndOperator

	// Keep only the last image by default.
	defaultImagesPerReferenceLast     = 1
	defaultImagesPerReferenceOperator = AndOperator
)

type MetaCleanup struct {
	DisbaleCleanup                     bool
	DisableKubernetesBasedPolicy       bool
	DisableGitHistoryBasedPolicy       bool
	DisableBuiltWithinLastNHoursPolicy bool
	KeepImagesBuiltWithinLastNHours    uint64
	KeepPolicies                       []*MetaCleanupKeepPolicy
}

type MetaCleanupKeepPolicy struct {
	References         MetaCleanupKeepPolicyReferences
	ImagesPerReference MetaCleanupKeepPolicyImagesPerReference
}

func (p *MetaCleanupKeepPolicy) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("references={%s}", p.References.String()))

	imagesPerReferencePart := p.ImagesPerReference.String()
	if imagesPerReferencePart != "" {
		parts = append(parts, fmt.Sprintf("imagesPerReference={%s}", imagesPerReferencePart))
	}

	return strings.Join(parts, " ")
}

type MetaCleanupKeepPolicyReferences struct {
	TagRegexp    *regexp.Regexp
	BranchRegexp *regexp.Regexp
	Limit        *MetaCleanupKeepPolicyReferencesLimit
}

func (c *MetaCleanupKeepPolicyReferences) String() string {
	var parts []string

	if c.TagRegexp != nil {
		parts = append(parts, fmt.Sprintf("tag=%s", c.TagRegexp.String()))
	} else {
		parts = append(parts, fmt.Sprintf("branch=%s", c.BranchRegexp.String()))
	}

	if c.Limit != nil {
		parts = append(parts, fmt.Sprintf("limit={%s}", c.Limit.String()))
	}

	return strings.Join(parts, " ")
}

func NewMetaCleanupKeepPolicyReferencesLimit(last *int, in *time.Duration, operator *Operator) *MetaCleanupKeepPolicyReferencesLimit {
	return &MetaCleanupKeepPolicyReferencesLimit{
		metaCleanupKeepPolicyLimit{
			last:     last,
			in:       in,
			operator: operator,
		},
	}
}

type MetaCleanupKeepPolicyReferencesLimit struct {
	metaCleanupKeepPolicyLimit
}

func (c *MetaCleanupKeepPolicyReferencesLimit) Last() int {
	if c.last == nil {
		return defaultReferencesLimitLast
	}
	return *c.last
}

func (c *MetaCleanupKeepPolicyReferencesLimit) Operator() Operator {
	if c.operator == nil || *c.operator == "" {
		return defaultReferencesLimitOperator
	}
	return *c.operator
}

func (c *MetaCleanupKeepPolicyReferencesLimit) String() string {
	return metaCleanupKeepPolicyLimitToString(c.Last(), c.In(), c.Operator())
}

type MetaCleanupKeepPolicyImagesPerReference struct {
	metaCleanupKeepPolicyLimit
}

func NewMetaCleanupKeepPolicyImagesPerReference(last *int, in *time.Duration, operator *Operator) MetaCleanupKeepPolicyImagesPerReference {
	return MetaCleanupKeepPolicyImagesPerReference{
		metaCleanupKeepPolicyLimit{
			last:     last,
			in:       in,
			operator: operator,
		},
	}
}

func (c *MetaCleanupKeepPolicyImagesPerReference) Last() int {
	if c.last == nil {
		return defaultImagesPerReferenceLast
	}
	return *c.last
}

func (c *MetaCleanupKeepPolicyImagesPerReference) Operator() Operator {
	if c.operator == nil || *c.operator == "" {
		return defaultImagesPerReferenceOperator
	}
	return *c.operator
}

func (c *MetaCleanupKeepPolicyImagesPerReference) String() string {
	return metaCleanupKeepPolicyLimitToString(c.Last(), c.In(), c.Operator())
}

type metaCleanupKeepPolicyLimit struct {
	last     *int
	in       *time.Duration
	operator *Operator
}

func (c *metaCleanupKeepPolicyLimit) In() *time.Duration {
	return c.in
}

func (c *metaCleanupKeepPolicyLimit) String() string {
	panic("not implemented")
}

func metaCleanupKeepPolicyLimitToString(last int, in *time.Duration, operator Operator) string {
	var parts []string

	if in != nil {
		parts = append(parts, fmt.Sprintf("in=%s", in.String()))
	}

	parts = append(parts, fmt.Sprintf("last=%d", last))
	parts = append(parts, fmt.Sprintf("operator=%s", operator))

	return strings.Join(parts, " ")
}

type Operator string

var (
	OrOperator  Operator = "Or"
	AndOperator Operator = "And"
)
