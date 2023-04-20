package clicommand

import (
	"context"
	"fmt"
	"os"

	"github.com/buildkite/agent/v3/leaderapi"
	"github.com/urfave/cli"
)

const lockDoneHelpDescription = `Usage:

   buildkite-agent lock release [key]

Description:
   Completes a do-once lock. This should only be used by the process performing
   the work.

Examples:

   #!/bin/bash
   if [ $(buildkite-agent lock do llama) = 'do' ] ; then
	  setup_code()
	  buildkite-agent lock done llama
   fi


`

type LockDoneConfig struct{}

var LockDoneCommand = cli.Command{
	Name:        "done",
	Usage:       "Completes a do-once lock",
	Description: lockDoneHelpDescription,
	Action:      lockDoneAction,
}

func lockDoneAction(c *cli.Context) error {
	if c.NArg() != 1 {
		fmt.Fprint(c.App.ErrWriter, lockDoneHelpDescription)
		os.Exit(1)
	}
	key := c.Args()[0]

	ctx := context.Background()

	cli, err := leaderapi.NewClient(leaderapi.LeaderSocketPath)
	if err != nil {
		fmt.Fprintf(c.App.ErrWriter, lockClientErrMessage, err)
		os.Exit(1)
	}

	val, done, err := cli.CompareAndSwap(ctx, key, "doing", "done")
	if err != nil {
		fmt.Fprintf(c.App.ErrWriter, "Error performing compare-and-swap: %v\n", err)
		os.Exit(1)
	}

	if !done {
		fmt.Fprintf(c.App.ErrWriter, "Lock in invalid state %q to mark complete - investigate with 'lock get'\n", val)
		os.Exit(1)
	}
	return nil
}
