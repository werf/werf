// THIS WAS COPIED FROM github.com/containers/image/v5/pkg/manifest

package manifest

import (
	"time"

	"github.com/werf/werf/test/pkg/thirdparty/contruntime/strslice"
)

// Schema2Port is a Port, a string containing port number and protocol in the
// format "80/tcp", from docker/go-connections/nat.
type Schema2Port string

// Schema2PortSet is a PortSet, a collection of structs indexed by Port, from
// docker/go-connections/nat.
type Schema2PortSet map[Schema2Port]struct{}

// Schema2HealthConfig is a HealthConfig, which holds configuration settings
// for the HEALTHCHECK feature, from docker/docker/api/types/container.
type Schema2HealthConfig struct {
	// Test is the test to perform to check that the container is healthy.
	// An empty slice means to inherit the default.
	// The options are:
	// {} : inherit healthcheck
	// {"NONE"} : disable healthcheck
	// {"CMD", args...} : exec arguments directly
	// {"CMD-SHELL", command} : run command with system's default shell
	Test []string `json:",omitempty"`

	// Zero means to inherit. Durations are expressed as integer nanoseconds.
	StartPeriod time.Duration `json:",omitempty"` // StartPeriod is the time to wait after starting before running the first check.
	Interval    time.Duration `json:",omitempty"` // Interval is the time to wait between checks.
	Timeout     time.Duration `json:",omitempty"` // Timeout is the time to wait before considering the check to have hung.

	// Retries is the number of consecutive failures needed to consider a container as unhealthy.
	// Zero means inherit.
	Retries int `json:",omitempty"`
}

// Schema2Config is a Config in docker/docker/api/types/container.
type Schema2Config struct {
	Hostname        string               // Hostname
	Domainname      string               // Domainname
	User            string               // User that will run the command(s) inside the container, also support user:group
	AttachStdin     bool                 // Attach the standard input, makes possible user interaction
	AttachStdout    bool                 // Attach the standard output
	AttachStderr    bool                 // Attach the standard error
	ExposedPorts    Schema2PortSet       `json:",omitempty"` // List of exposed ports
	Tty             bool                 // Attach standard streams to a tty, including stdin if it is not closed.
	OpenStdin       bool                 // Open stdin
	StdinOnce       bool                 // If true, close stdin after the 1 attached client disconnects.
	Env             []string             // List of environment variable to set in the container
	Cmd             strslice.StrSlice    // Command to run when starting the container
	Healthcheck     *Schema2HealthConfig `json:",omitempty"` // Healthcheck describes how to check the container is healthy
	ArgsEscaped     bool                 `json:",omitempty"` // True if command is already escaped (Windows specific)
	Image           string               // Name of the image as it was passed by the operator (e.g. could be symbolic)
	Volumes         map[string]struct{}  // List of volumes (mounts) used for the container
	WorkingDir      string               // Current directory (PWD) in the command will be launched
	Entrypoint      strslice.StrSlice    // Entrypoint to run when starting the container
	NetworkDisabled bool                 `json:",omitempty"` // Is network disabled
	MacAddress      string               `json:",omitempty"` // Mac Address of the container
	OnBuild         []string             // ONBUILD metadata that were defined on the image Dockerfile
	Labels          map[string]string    // List of labels set to this container
	StopSignal      string               `json:",omitempty"` // Signal to stop a container
	StopTimeout     *int                 `json:",omitempty"` // Timeout (in seconds) to stop a container
	Shell           strslice.StrSlice    `json:",omitempty"` // Shell for shell-form of RUN, CMD, ENTRYPOINT
}
