package docker_registry

const QuayImplementationName = "quay"

var quayPatterns = []string{"^quay\\.io"}

type quay struct {
	*defaultImplementation
}

type quayOptions struct {
	defaultImplementationOptions
}

func newQuay(options quayOptions) (*quay, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	quay := &quay{defaultImplementation: d}

	return quay, nil
}
