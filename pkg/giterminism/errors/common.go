package errors

import "fmt"

const docPageURL = "https://werf.io/v1.2-alpha/documentation/advanced/configuration/giterminism.html"

func NewError(msg string) error {
	return fmt.Errorf(`%s

To provide a strong guarantee of reproducibility, werf reads the configuration and build context files from the project git repository and eliminates external dependencies. 
We strongly recommend to follow this approach. But if necessary, you can allow the reading of specific files directly from the project directory and enable the features that require careful use. 
Read more about giterminism and how to manage it here: %s.`, msg, docPageURL)
}
