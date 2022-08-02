package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type MetaCleanup struct {
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
	Limit        *MetaCleanupKeepPolicyLimit
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

type MetaCleanupKeepPolicyImagesPerReference struct {
	MetaCleanupKeepPolicyLimit
}

type MetaCleanupKeepPolicyLimit struct {
	Last     *int
	In       *time.Duration
	Operator *Operator
}

func (c *MetaCleanupKeepPolicyLimit) String() string {
	var parts []string

	if c.In != nil {
		parts = append(parts, fmt.Sprintf("in=%s", c.In.String()))
	}

	if c.Last != nil {
		parts = append(parts, fmt.Sprintf("last=%d", *c.Last))
	}

	if c.Operator != nil {
		parts = append(parts, fmt.Sprintf("operator=%s", *c.Operator))
	}

	return strings.Join(parts, " ")
}

type Operator string

var (
	OrOperator  Operator = "Or"
	AndOperator Operator = "And"
)
