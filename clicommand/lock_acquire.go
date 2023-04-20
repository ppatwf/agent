package clicommand

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/buildkite/agent/v3/internal/leaderapi"
	"github.com/urfave/cli"
)

const lockAcquireHelpDescription = `Usage:

   buildkite-agent lock acquire [key]

Description:
   Acquires the lock for the given key. ′lock acquire′ will wait (potentially
   forever) until it can acquire the lock, if the lock is already held by
   another process. If multiple processes are waiting for the same lock, there
   is no ordering guarantee of which one will be given the lock next.

Examples:

   $ buildkite-agent lock acquire llama
   $ critical_section()
   $ buildkite-agent lock release llama

`

type LockAcquireConfig struct {
	SocketsPath string `cli:"sockets-path" normalize:"filepath"`
}

var LockAcquireCommand = cli.Command{
	Name:        "acquire",
	Usage:       "Acquires a lock from the agent leader",
	Description: lockAcquireHelpDescription,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "sockets-path",
			Value:  defaultSocketsPath(),
			Usage:  "Directory where the agent will place sockets",
			EnvVar: "BUILDKITE_SOCKETS_PATH",
		},
	},
	Action: lockAcquireAction,
}

func lockAcquireAction(c *cli.Context) error {
	if c.NArg() != 1 {
		fmt.Fprint(c.App.ErrWriter, lockAcquireHelpDescription)
		os.Exit(1)
	}
	key := c.Args()[0]

	ctx := context.Background()

	cli, err := leaderapi.NewClient(leaderapi.LeaderSocketPath)
	if err != nil {
		fmt.Fprintf(c.App.ErrWriter, lockClientErrMessage, err)
		os.Exit(1)
	}

	for {
		_, done, err := cli.CompareAndSwap(ctx, key, "", "acquired")
		if err != nil {
			fmt.Fprintf(c.App.ErrWriter, "Error performing compare-and-swap: %v\n", err)
			os.Exit(1)
		}

		if done {
			return nil
		}

		// Not done.
		if err := sleep(ctx, 100*time.Millisecond); err != nil {
			fmt.Fprintf(c.App.ErrWriter, "Exceeded deadline or context cancelled: %v\n", err)
			os.Exit(1)
		}

	}
}
