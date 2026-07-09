package buildkit

import (
	"fmt"
	"strings"

	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
)

// ParseSecretSpecs parses docker-CLI-style secret specs ("id=myid,src=/path" or
// "id=myid,env=MYENV") into secretsprovider sources.
func ParseSecretSpecs(specs []string) ([]secretsprovider.Source, error) {
	var sources []secretsprovider.Source
	for _, spec := range specs {
		source := secretsprovider.Source{}
		for _, field := range strings.Split(spec, ",") {
			parts := strings.SplitN(field, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid secret spec %q: expected comma-separated key=value fields", spec)
			}
			key, value := parts[0], parts[1]
			switch key {
			case "id":
				source.ID = value
			case "src", "source":
				source.FilePath = value
			case "env":
				source.Env = value
			default:
				return nil, fmt.Errorf("invalid secret spec %q: unexpected field %q", spec, key)
			}
		}
		if source.ID == "" {
			return nil, fmt.Errorf("invalid secret spec %q: id is required", spec)
		}
		sources = append(sources, source)
	}
	return sources, nil
}

// ParseSSHSpec parses a docker-CLI-style ssh spec ("default", "id", "id=/path/to/sock"
// or "id=/path/to/key1,/path/to/key2") into sshprovider agent configs. defaultAgentSock
// is used for specs without explicit paths.
func ParseSSHSpec(spec, defaultAgentSock string) ([]sshprovider.AgentConfig, error) {
	if spec == "" {
		return nil, nil
	}

	var configs []sshprovider.AgentConfig
	for _, item := range strings.Split(spec, ";") {
		conf := sshprovider.AgentConfig{}
		parts := strings.SplitN(item, "=", 2)
		conf.ID = parts[0]
		if len(parts) == 2 {
			conf.Paths = strings.Split(parts[1], ",")
		} else if defaultAgentSock != "" {
			conf.Paths = []string{defaultAgentSock}
		} else {
			return nil, fmt.Errorf("invalid ssh spec %q: no socket path given and no ssh agent socket available", spec)
		}
		configs = append(configs, conf)
	}
	return configs, nil
}
