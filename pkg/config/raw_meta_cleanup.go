package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type rawMetaCleanup struct {
	Policies []*rawMetaCleanupPolicy `yaml:"policies,omitempty"`

	rawMeta               *rawMeta
	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type rawMetaCleanupPolicy struct {
	Tag                string                 `yaml:"tag,omitempty"`
	Branch             string                 `yaml:"branch,omitempty"`
	RefsToKeepImagesIn *rawRefsToKeepImagesIn `yaml:"refsToKeepImagesIn,omitempty"`
	ImageDepthToKeep   int                    `yaml:"imageDepthToKeep,omitempty"`

	TagRegexp    *regexp.Regexp
	BranchRegexp *regexp.Regexp

	rawMetaCleanup        *rawMetaCleanup
	UnsupportedAttributes map[string]interface{} `yaml:",inline"`
}

type rawRefsToKeepImagesIn struct {
	Last       int           `yaml:"last,omitempty"`
	ModifiedIn time.Duration `yaml:"modifiedIn,omitempty"`
	Operator   string        `yaml:"operator,omitempty"`
}

func (c *rawMetaCleanup) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMeta); ok {
		c.rawMeta = parent
	}

	parentStack.Push(c)
	type plain rawMetaCleanup
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawMeta.doc); err != nil {
		return err
	}

	return nil
}

func (c *rawMetaCleanupPolicy) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if parent, ok := parentStack.Peek().(*rawMetaCleanup); ok {
		c.rawMetaCleanup = parent
	}

	parentStack.Push(c)
	type plain rawMetaCleanupPolicy
	err := unmarshal((*plain)(c))
	parentStack.Pop()
	if err != nil {
		return err
	}

	if err := checkOverflow(c.UnsupportedAttributes, c, c.rawMetaCleanup.rawMeta.doc); err != nil {
		return err
	}

	if c.Tag == "" && c.Branch == "" {
		return newDetailedConfigError("tag `tag: string|REGEX` or branch `branch: string|REGEX` required for cleanup policy!", c, c.rawMetaCleanup.rawMeta.doc)
	} else if c.Tag != "" && c.Branch != "" {
		return newDetailedConfigError("specify only tag `tag: string|REGEX` or branch `branch: string|REGEX` for cleanup policy!", c, c.rawMetaCleanup.rawMeta.doc)
	}

	if c.Branch != "" {
		regex, err := c.processRegexpString("branch", c.Branch)
		if err != nil {
			return err
		}

		c.BranchRegexp = regex
	} else {
		regex, err := c.processRegexpString("tag", c.Tag)
		if err != nil {
			return err
		}

		c.TagRegexp = regex
	}

	if c.Tag != "" && c.ImageDepthToKeep != 0 {
		return newDetailedConfigError("imageDepthToKeep cannot be defined for tag reference!", c, c.rawMetaCleanup.rawMeta.doc)
	} else {
		c.ImageDepthToKeep = 1
	}

	if c.RefsToKeepImagesIn != nil {
		if c.RefsToKeepImagesIn.Operator != "" {
			if c.RefsToKeepImagesIn.Operator != "Or" && c.RefsToKeepImagesIn.Operator != "And" {
				return newDetailedConfigError(fmt.Sprintf("unsupported value '%s' for operator `operator: Or|And`!", c.RefsToKeepImagesIn.Operator), c, c.rawMetaCleanup.rawMeta.doc)
			}
		} else {
			c.RefsToKeepImagesIn.Operator = "Or"
		}
	}

	return nil
}

func (c *rawMetaCleanupPolicy) processRegexpString(name, value string) (*regexp.Regexp, error) {
	if strings.HasPrefix(value, "/") && strings.HasSuffix(value, "/") {
		value = strings.TrimPrefix(value, "/")
	}

	regex, err := regexp.Compile(c.Tag)
	if err != nil {
		return nil, newDetailedConfigError(fmt.Sprintf("invalid value '%s' for `%s: string|REGEX`!", value, name), c, c.rawMetaCleanup.rawMeta.doc)
	}

	return regex, nil
}

func (c *rawMetaCleanup) toMetaCleanup() MetaCleanup {
	metaCleanup := MetaCleanup{}

	for _, policy := range c.Policies {
		metaCleanup.Policies = append(metaCleanup.Policies, policy.toMetaCleanupPolicy())
	}

	return metaCleanup
}

func (c *rawMetaCleanupPolicy) toMetaCleanupPolicy() *MetaCleanupPolicy {
	metaCleanupPolicy := &MetaCleanupPolicy{}

	metaCleanupPolicy.BranchRegexp = c.BranchRegexp
	metaCleanupPolicy.TagRegexp = c.TagRegexp

	if c.RefsToKeepImagesIn != nil {
		metaCleanupPolicy.RefsToKeepImagesIn = c.RefsToKeepImagesIn.toRefsToKeepImagesIn()
	}

	metaCleanupPolicy.ImageDepthToKeep = &c.ImageDepthToKeep

	return metaCleanupPolicy
}

func (c *rawRefsToKeepImagesIn) toRefsToKeepImagesIn() *RefsToKeepImagesIn {
	refsToKeepImagesIn := &RefsToKeepImagesIn{}

	refsToKeepImagesIn.Last = &c.Last

	if c.Operator == "Or" {
		refsToKeepImagesIn.Operator = OrOperator
	} else if c.Operator == "And" {
		refsToKeepImagesIn.Operator = AndOperator
	}

	refsToKeepImagesIn.ModifiedIn = &c.ModifiedIn

	return refsToKeepImagesIn
}
