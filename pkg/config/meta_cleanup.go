package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type MetaCleanup struct {
	KeepPolicies []*MetaCleanupKeepPolicy
}

type MetaCleanupKeepPolicy struct {
	References         MetaCleanupKeepPolicyReferences
	ImagesPerReference MetaCleanupKeepPolicyImagesPerReference
}

func (p *MetaCleanupKeepPolicy) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("references={%s}", p.References.String()))
	parts = append(parts, fmt.Sprintf("imagesPerReference={%s}", p.ImagesPerReference.String()))

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
	Last        *int
	PublishedIn *time.Duration
	Operator    *Operator
}

func (c *MetaCleanupKeepPolicyImagesPerReference) String() string {
	var parts []string

	if c.PublishedIn != nil {
		parts = append(parts, fmt.Sprintf("publshedIn=%s", c.PublishedIn.String()))
	}

	if c.Last != nil {
		parts = append(parts, fmt.Sprintf("last=%d", *c.Last))
	}

	if c.Operator != nil {
		parts = append(parts, fmt.Sprintf("operator=%s", *c.Operator))
	}

	return strings.Join(parts, " ")
}

type MetaCleanupKeepPolicyLimit struct {
	Last      *int
	CreatedIn *time.Duration
	Operator  *Operator
}

func (c *MetaCleanupKeepPolicyLimit) String() string {
	var parts []string

	if c.CreatedIn != nil {
		parts = append(parts, fmt.Sprintf("createdIn=%s", c.CreatedIn.String()))
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
