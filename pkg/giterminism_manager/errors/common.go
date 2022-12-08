package errors

import "fmt"

const docPageURL = "https://werf.io/documentation/usage/project_configuration/giterminism.html"

func NewError(msg string) error {
	return fmt.Errorf(`%s

To provide a strong guarantee of reproducibility, werf reads the configuration and build's context files from the project git repository, and eliminates external dependencies. We strongly recommend following this approach, but if necessary, you can allow the reading of specific files directly from the file system and enable the features that require careful use. Read more about giterminism and how to manage it here: %s.`, msg, docPageURL)
}
