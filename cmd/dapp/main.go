package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmdData struct {
	Name    string
	Dir     string
	SSHKeys []string
}

func main() {
	cmd := &cobra.Command{
		Use: "dapp",
	}

	cmd.AddCommand(
		newBuildCmd(),
		newPushCmd(),
		newBPCmd(),
	)

	cmd.PersistentFlags().StringVarP(&rootCmdData.Name, "name", "", "", `Use custom dapp name.
Chaging default name will cause full cache rebuild.
By default dapp name is the last element of remote.origin.url from project git,
or it is the name of the directory where Dappfile resides.`)
	cmd.PersistentFlags().StringVarP(&rootCmdData.Dir, "dir", "", "", "Change to the specified directory to find dappfile")
	cmd.PersistentFlags().StringVarP(&rootCmdData.Dir, "tmp-dir", "", "", "Use specified dir to store tmp dirs")
	cmd.PersistentFlags().StringArrayVarP(&rootCmdData.SSHKeys, "ssh-key", "", []string{}, "Enable only specified ssh keys (use system ssh-agent by default)")

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
