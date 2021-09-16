package buildah

import (
	"context"
	"fmt"
	"os"

	"github.com/containers/buildah"
	buildahcli "github.com/containers/buildah/pkg/cli"

	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"

	"github.com/pkg/errors"
)

/*
InitProcess() {
	if reexec.Init() {
		return
	}
	
	unshare.MaybeReexecUsingUserNamespace(false)
}

Init(Options) -> BuildahObject {
	parse options & init store
}
BuildahObject.Bud(ctx, ...) // обязательно принимать dockerfile через []byte, т.к. гитерминизм 
BuildahObject.Run(ctx)
...
*/

var store storage.Store

func Init() error {
	options, err := storage.DefaultStoreOptions(unshare.IsRootless(), unshare.GetRootlessUID())
	if err != nil {
		return err
	}

	//fmt.Printf("OPTIONS: %#v\n", options)

	s, err := storage.GetStore(options)
	if err != nil {
		return fmt.Errorf("unable to get storage: %s", err)
	}
	if s != nil {
		is.Transport.SetStore(store)
	}
	store = s

	return nil
}

type runInputOptions struct {
	addHistory  bool
	capAdd      []string
	capDrop     []string
	env         []string
	hostname    string
	isolation   string
	mounts      []string
	runtime     string
	runtimeFlag []string
	noPivot     bool
	terminal    bool
	volumes     []string
	workingDir  string
	*buildahcli.NameSpaceResults
}

func NewRunInputOptions() *runInputOptions {
	return &runInputOptions{
		NameSpaceResults: &buildahcli.NameSpaceResults{},
	}
}

func Run(ctx context.Context, container string, command []string, opts *runInputOptions) error {
	options := buildah.RunOptions{
		Hostname: opts.hostname,
		Runtime:  opts.runtime,
		// Args:             runtimeFlags,
		// NoPivot:          noPivot,
		// User:             c.Flag("user").Value.String(),
		// Isolation:        isolation,
		// NamespaceOptions: namespaceOptions,
		// ConfigureNetwork: networkPolicy,
		CNIPluginPath:    opts.CNIPlugInPath,
		CNIConfigDir:     opts.CNIConfigDir,
		AddCapabilities:  opts.capAdd,
		DropCapabilities: opts.capDrop,
		Env:              opts.env,
		WorkingDir:       opts.workingDir,
	}

	//fmt.Printf("HELO\n")
	b, err := buildah.OpenBuilder(store, container)
	//fmt.Printf("OpenBuilder -- %#v\n", b)
	if os.IsNotExist(errors.Cause(err)) {
		b, err = buildah.ImportBuilder(ctx, store, buildah.ImportOptions{
			Container: container,
		})
		if err != nil {
			return fmt.Errorf("unable to import builder for container %q: %s", container, err)
		}
	} else if err != nil {
		return fmt.Errorf("unable to open builder for container %q: %s", container, err)
	}

	return b.Run(command, options)
}
