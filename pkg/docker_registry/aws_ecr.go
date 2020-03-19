package docker_registry

const AwsEcrImplementationName = "ecr"

var awsEcrPatterns = []string{"^.*\\.dkr\\.ecr\\..*\\.amazonaws\\.com"}

type awsEcr struct {
	*defaultImplementation
}

type awsEcrOptions struct {
	defaultImplementationOptions
}

func newAwsEcr(options awsEcrOptions) (*awsEcr, error) {
	d, err := newDefaultImplementation(options.defaultImplementationOptions)
	if err != nil {
		return nil, err
	}

	awsEcr := &awsEcr{defaultImplementation: d}

	return awsEcr, nil
}

func (r *awsEcr) String() string {
	return AwsEcrImplementationName
}
