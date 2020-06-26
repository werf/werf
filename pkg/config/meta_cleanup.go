package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type MetaCleanup struct {
	Policies []*MetaCleanupPolicy
}

type MetaCleanupPolicy struct {
	TagRegexp          *regexp.Regexp
	BranchRegexp       *regexp.Regexp
	RefsToKeepImagesIn *RefsToKeepImagesIn
	ImageDepthToKeep   *int
}

func (p *MetaCleanupPolicy) String() string {
	var parts []string

	if p.TagRegexp != nil {
		parts = append(parts, fmt.Sprintf("tag=%s", p.TagRegexp.String()))
	} else {
		parts = append(parts, fmt.Sprintf("branch=%s", p.BranchRegexp.String()))
	}

	if p.ImageDepthToKeep != nil {
		parts = append(parts, fmt.Sprintf("imageDepthToKeep=%d", *p.ImageDepthToKeep))
	}

	if p.RefsToKeepImagesIn != nil {
		parts = append(parts, fmt.Sprintf("refsToKeepImagesIn={%s}", p.RefsToKeepImagesIn.String()))
	}

	return strings.Join(parts, " ")
}

type RefsToKeepImagesIn struct {
	Last       *int
	ModifiedIn *time.Duration
	Operator   Operator
}

func (r *RefsToKeepImagesIn) String() string {
	var parts []string

	if r.ModifiedIn != nil {
		parts = append(parts, fmt.Sprintf("modefiedIn=%s", r.ModifiedIn.String()))
	}

	if r.Last != nil {
		parts = append(parts, fmt.Sprintf("last=%d", *r.Last))
	}

	parts = append(parts, fmt.Sprintf("operator=%s", r.Operator))

	return strings.Join(parts, " ")
}

type Operator string

const (
	OrOperator  Operator = "Or"
	AndOperator Operator = "And"
)
