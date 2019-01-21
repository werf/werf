package slug

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flant/werf/pkg/util"
)

const slugSeparator = "-"

var (
	slugMaxSize = 42

	dockerTagRegexp  = regexp.MustCompile(`^[\w][\w.-]*$`)
	dockerTagMaxSize = 128

	projectNameRegex   = regexp.MustCompile(`^(?:[a-z0-9]|[a-z0-9][a-z0-9-]*[a-z0-9])$`)
	projectNameMaxSize = 50

	dnsLabelRegex   = regexp.MustCompile(`^(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])$`)
	dnsLabelMaxSize = 63

	helmReleaseRegexp  = regexp.MustCompile(`^(?:(?:[A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])+$`)
	helmReleaseMaxSize = 53
)

func Slug(data string) string {
	if len(data) == 0 || slugify(data) == data && len(data) < slugMaxSize {
		return data
	}

	return slug(data, slugMaxSize)
}

func Project(name string) string {
	if shouldNotBeSlugged(name, projectNameRegex, projectNameMaxSize) {
		return name
	}

	res := slugify(name)

	if len(res) > projectNameMaxSize {
		res = res[:projectNameMaxSize]
	}

	return res
}

func ValidateProject(name string) error {
	if shouldNotBeSlugged(name, projectNameRegex, projectNameMaxSize) {
		return nil
	}
	return fmt.Errorf("Project name should comply with regex '%s' and be maximum %d chars", projectNameRegex, projectNameMaxSize)
}

func DockerTag(tag string) string {
	if shouldNotBeSlugged(tag, dockerTagRegexp, dockerTagMaxSize) {
		return tag
	}

	return slug(tag, dockerTagMaxSize)
}

func ValidateDockerTag(tag string) error {
	if shouldNotBeSlugged(tag, dockerTagRegexp, dockerTagMaxSize) {
		return nil
	}
	return fmt.Errorf("Docker tag should comply with regex '%s' and be maximum %d chars", dockerTagRegexp, dockerTagMaxSize)
}

func KubernetesNamespace(namespace string) string {
	if shouldNotBeSlugged(namespace, dnsLabelRegex, dnsLabelMaxSize) {
		return namespace
	}

	return slug(namespace, dnsLabelMaxSize)
}

func ValidateKubernetesNamespace(namespace string) error {
	if shouldNotBeSlugged(namespace, dnsLabelRegex, dnsLabelMaxSize) {
		return nil
	}
	return fmt.Errorf("Kubernetes namespace should comply with DNS Label requirements: %s and %d bytes max", dnsLabelRegex, dnsLabelMaxSize)
}

func HelmRelease(name string) string {
	if shouldNotBeSlugged(name, helmReleaseRegexp, helmReleaseMaxSize) {
		return name
	}

	return slug(name, helmReleaseMaxSize)
}

func ValidateHelmRelease(name string) error {
	if shouldNotBeSlugged(name, helmReleaseRegexp, helmReleaseMaxSize) {
		return nil
	}
	return fmt.Errorf("Helm release name should comply with regex '%s' and be maximum %d chars", helmReleaseRegexp, helmReleaseMaxSize)
}

func shouldNotBeSlugged(data string, regexp *regexp.Regexp, maxSize int) bool {
	return len(data) == 0 || regexp.Match([]byte(data)) && len(data) < maxSize
}

func slug(data string, maxSize int) string {
	sluggedData := slugify(data)
	murmurHash := util.MurmurHash(data)

	var slugParts []string
	if sluggedData != "" {
		croppedSluggedData := cropSluggedData(sluggedData, maxSize)
		if strings.HasPrefix(croppedSluggedData, "-") {
			slugParts = append(slugParts, croppedSluggedData[:len(croppedSluggedData)-1])
		} else {
			slugParts = append(slugParts, croppedSluggedData)
		}
	}
	slugParts = append(slugParts, murmurHash)

	consistentUniqSlug := strings.Join(slugParts, slugSeparator)

	return consistentUniqSlug
}

func cropSluggedData(data string, maxSize int) string {
	var index int
	maxLength := maxSize - len(util.MurmurHash()) - len(slugSeparator)
	if len(data) > maxLength {
		index = maxLength
	} else {
		index = len(data)
	}

	return data[:index]
}

func slugify(data string) string {
	var result []rune

	var isCursorDash bool
	var isPreviousDash bool
	var isStartedDash, isDoubledDash bool

	isResultEmpty := true
	for _, r := range data {
		cursor := algorithm(string(r))
		if cursor == "" {
			continue
		}

		isCursorDash = cursor == "-"
		isStartedDash = isCursorDash && isResultEmpty
		isDoubledDash = isCursorDash && !isResultEmpty && isPreviousDash

		if isStartedDash || isDoubledDash {
			continue
		}

		result = append(result, []rune(cursor)...)
		isPreviousDash = isCursorDash
		isResultEmpty = false
	}

	isEndedDash := !isResultEmpty && isCursorDash
	if isEndedDash {
		return string(result[:len(result)-1])
	}
	return string(result)
}

func algorithm(data string) string {
	var result string
	for ind := range data {
		char, ok := mapping[string([]rune(data)[ind])]
		if ok {
			result += char
		}
	}

	return result
}
