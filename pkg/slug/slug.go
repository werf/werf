package slug

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/werf/werf/pkg/util"
)

const slugSeparator = "-"

var (
	DefaultSlugMaxSize = 42 // legacy

	dockerTagRegexp  = regexp.MustCompile(`^[\w][\w.-]*$`)
	DockerTagMaxSize = 128

	projectNameRegex   = regexp.MustCompile(`^(?:[a-z0-9]|[a-z0-9][a-z0-9-]*[a-z0-9])$`)
	projectNameMaxSize = 50

	kubernetesNamespaceMaxSize = 63
	helmReleaseMaxSize         = 53
)

func Slug(data string) string {
	return slugify(data)
}

func LimitedSlug(data string, slugMaxSize int) string {
	if !shouldBeSlugged(data, slugMaxSize) {
		return data
	}

	return slug(data, slugMaxSize)
}

func shouldBeSlugged(data string, slugMaxSize int) bool {
	// nothing to slug
	if len(data) == 0 {
		return false
	}

	// valid data
	if slugify(data) == data && len(data) <= slugMaxSize {
		return false
	}

	// legacy: this code provides idempotence for slugged data with sequence of two hyphens in a specific place
	// Remove this block in the following major releases.
	{
		// data length cannot be more than max size
		if len(data) > slugMaxSize {
			return true
		}

		// data must contain only one sequence
		if strings.Count(data, "--") != 1 {
			return true
		}

		// data without sequence must be valid
		{
			formattedData := strings.Replace(data, "--", "-", 1)
			if slugify(formattedData) != formattedData {
				return true
			}
		}

		return false
	}
}

func Project(name string) string {
	if err := validateProject(name); err != nil {
		res := slugify(name)
		if len(res) > projectNameMaxSize {
			res = res[:projectNameMaxSize]
		}
		return res
	}
	return name
}

func ValidateProject(name string) error {
	return validateProject(name)
}

func validateProject(name string) error {
	if shouldNotBeSlugged(name, projectNameRegex, projectNameMaxSize) {
		return nil
	}
	return fmt.Errorf("project name should comply with regex %q and be maximum %d chars", projectNameRegex, projectNameMaxSize)
}

func DockerTag(name string) string {
	if err := ValidateDockerTag(name); err != nil {
		return slug(name, DockerTagMaxSize)
	}
	return name
}

func IsValidDockerTag(name string) bool {
	if shouldNotBeSlugged(name, dockerTagRegexp, DockerTagMaxSize) {
		return true
	}

	return false
}

func ValidateDockerTag(name string) error {
	if IsValidDockerTag(name) {
		return nil
	}

	return fmt.Errorf(`%q is not a valid docker tag

 - a tag name must be valid ASCII and may contain lowercase and uppercase letters, digits, underscores, periods and dashes;
 - a tag name may not start with a period or a dash and may contain a maximum of 128 characters.`, name)
}

func KubernetesNamespace(name string) string {
	if err := validateKubernetesNamespace(name); err != nil {
		return slug(name, kubernetesNamespaceMaxSize)
	}
	return name
}

func ValidateKubernetesNamespace(namespace string) error {
	return validateKubernetesNamespace(namespace)
}

func validateKubernetesNamespace(name string) error {
	errorMsgPrefix := fmt.Sprintf("kubernetes namespace should be a valid DNS-1123 label")
	if len(name) == 0 {
		return nil
	} else if len(name) > kubernetesNamespaceMaxSize {
		return fmt.Errorf("%s: %q is %d chars long", errorMsgPrefix, name, len(name))
	} else if msgs := validation.IsDNS1123Label(name); len(msgs) > 0 {
		return fmt.Errorf("%s: %s", errorMsgPrefix, strings.Join(msgs, ", "))
	}
	return nil
}

func HelmRelease(name string) string {
	if err := validateHelmRelease(name); err != nil {
		return slug(name, helmReleaseMaxSize)
	}
	return name
}

func ValidateHelmRelease(name string) error {
	return validateHelmRelease(name)
}

func validateHelmRelease(name string) error {
	errorMsgPrefix := fmt.Sprintf("helm release name should be a valid DNS-1123 subdomain and be maximum %d chars", helmReleaseMaxSize)
	if len(name) == 0 {
		return nil
	} else if len(name) > helmReleaseMaxSize {
		return fmt.Errorf("%s: %q is %d chars long", errorMsgPrefix, name, len(name))
	} else if msgs := validation.IsDNS1123Subdomain(name); len(msgs) > 0 {
		return fmt.Errorf("%s: %s", errorMsgPrefix, strings.Join(msgs, ", "))
	}
	return nil
}

func shouldNotBeSlugged(data string, regexp *regexp.Regexp, maxSize int) bool {
	return len(data) == 0 || regexp.Match([]byte(data)) && len(data) <= maxSize
}

func slug(data string, maxSize int) string {
	sluggedData := slugify(data)
	murmurHash := util.LegacyMurmurHash(data)

	var slugParts []string
	if sluggedData != "" {
		croppedSluggedData := cropSluggedData(sluggedData, murmurHash, maxSize)

		// legacy: this check cannot be fixed without breaking reproducibility.
		// Fix by replacing HasPrefix on HasSuffix in the following major releases.
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

func cropSluggedData(data, hash string, maxSize int) string {
	var index int
	maxLength := maxSize - len(hash) - len(slugSeparator)
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
