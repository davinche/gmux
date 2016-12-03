package main

import (
	"os"

	gmux "github.com/davinche/gmux/cli"
	"github.com/urfave/cli"
)

var VERSION string

func main() {
	app := cli.NewApp()
	app.Name = "GMux"
	app.Usage = "a tmux sessions manager"
	app.Version = VERSION

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug logging",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "new",
			Usage:     "create a new gmux project",
			Action:    gmux.New,
			ArgsUsage: "project_name",
		},
		{
			Name:      "start",
			Usage:     "start a tmux session using a gmux config",
			Action:    gmux.Start,
			ArgsUsage: "project_name",
		},
		{
			Name:        "stop",
			Usage:       "stops a tmux session",
			Description: "Removes a tmux session by running `tmux kill-session -t sessionname`.",
			ArgsUsage:   "session_name",
			Action:      gmux.Stop,
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "lists all available gmux projects",
			Action:  gmux.List,
		},
		{
			Name:      "edit",
			Usage:     "edit a gmux project configuration",
			ArgsUsage: "project_name",
			Action:    gmux.Edit,
		},
	}

	// Default action to show the help menu
	app.Action = func(c *cli.Context) error {
		projectName := c.Args().First()
		if projectName != "" {
			return gmux.Start(c)
		}
		return gmux.ShowHelp(c)
	}
	app.Run(os.Args)
}
