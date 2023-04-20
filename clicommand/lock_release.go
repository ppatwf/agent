package clicommand

import (
	"context"
	"fmt"
	"os"

	"github.com/buildkite/agent/v3/leaderapi"
	"github.com/urfave/cli"
)

const lockReleaseHelpDescription = `Usage:

   buildkite-agent lock release [key]

Description:
   Releases the lock for the given key. This should only be called by the
   process that acquired the lock.

Examples:

   $ buildkite-agent lock acquire llama
   $ critical_section()
   $ buildkite-agent lock release llama

`

type LockReleaseConfig struct{}

var LockReleaseCommand = cli.Command{
	Name:        "release",
	Usage:       "Releases a previously-acquired lock",
	Description: lockReleaseHelpDescription,
	Action:      lockReleaseAction,
}

func lockReleaseAction(c *cli.Context) error {
	if c.NArg() != 1 {
		fmt.Fprint(c.App.ErrWriter, lockReleaseHelpDescription)
		os.Exit(1)
	}
	key := c.Args()[0]

	ctx := context.Background()

	cli, err := leaderapi.NewClient(leaderapi.LeaderSocketPath)
	if err != nil {
		fmt.Fprintf(c.App.ErrWriter, lockClientErrMessage, err)
		os.Exit(1)
	}

	val, done, err := cli.CompareAndSwap(ctx, key, "acquired", "")
	if err != nil {
		fmt.Fprintf(c.App.ErrWriter, "Error performing compare-and-swap: %v\n", err)
		os.Exit(1)
	}

	if !done {
		fmt.Fprintf(c.App.ErrWriter, "Lock in invalid state %q to release - investigate with 'lock get'\n", val)
		os.Exit(1)
	}
	return nil
}
