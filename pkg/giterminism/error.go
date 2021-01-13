package giterminism

import "fmt"

type ConfigNotFoundError error

const GiterminismDocPageURL = "https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html"

func IsConfigNotFoundError(err error) bool {
	switch err.(type) {
	case ConfigNotFoundError:
		return true
	default:
		return false
	}
}

func NewConfigNotFoundError(msg string) ConfigNotFoundError {
	return ConfigNotFoundError(NewError(msg))
}

func NewUncommittedConfigurationError(msg string) error {
	return NewError(fmt.Sprintf("the uncommitted configuration found in the project directory: %s", msg))
}

func NewError(msg string) error {
	return fmt.Errorf(`%s

To provide a strong guarantee of reproducibility, werf reads the configuration and build context files from the project git repository.
We strongly recommend to follow this approach. But if you need, you can allow reading particular files directly from the project directory.
Read more about giterminism and how to manage it here: %s.`, msg, GiterminismDocPageURL)
}
