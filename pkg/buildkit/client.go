package buildkit

import (
	"context"
	"fmt"

	"github.com/moby/buildkit/client"
	_ "github.com/moby/buildkit/client/connhelper/dockercontainer"
	_ "github.com/moby/buildkit/client/connhelper/kubepod"
	_ "github.com/moby/buildkit/client/connhelper/podmancontainer"
	_ "github.com/moby/buildkit/client/connhelper/ssh"
)

func NewClient(ctx context.Context, host string) (*client.Client, error) {
	c, err := client.New(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("create buildkit client for %q: %w", host, err)
	}
	return c, nil
}
