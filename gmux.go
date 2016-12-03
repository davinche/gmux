package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/davinche/gmux/config"
	"github.com/urfave/cli"
)

var VERSION string

func startServer() {
	cmd := exec.Command("tmux", "start-server")
	if err := cmd.Run(); err != nil {
		os.Stderr.WriteString("could not start tmux server")
		os.Exit(1)
	}
}

func hasSession(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	return err == nil
}

func init() {
	startServer()
}

func createNew(c *cli.Context) error {
	projectName := c.Args().First()
	if projectName == "" {
		return showHelp(c)
	}
	if config.Exists(projectName) {
		return fmt.Errorf("project with the same name already exists")
	}

	newConfig := config.New(projectName)
	if err := newConfig.Write(); err != nil {
		return err
	}
	return config.EditProject(projectName)
}

func start(c *cli.Context) error {
	projectName := c.Args().First()
	if projectName == "" {
		return showHelp(c)
	}

	if hasSession(projectName) {
		if err := config.AttachToSession(projectName); err != nil {
			return cli.NewExitError(fmt.Sprintf("could not attach to session %q", projectName), 1)
		}
	}

	if err := config.GetAndRun(projectName, c.Bool("debug")); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

func stop(c *cli.Context) error {
	sessionName := c.Args().First()

	if sessionName == "" {
		cmd := exec.Command("tmux", "display-message", "-p", "#S")
		output, err := cmd.Output()
		if err != nil {
			return cli.NewExitError("could not determine current tmux session", 1)
		}
		sessionName = strings.TrimSpace(string(output))
	}
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	return nil
}

func list(c *cli.Context) error {
	return config.ListProjects()
}

func edit(c *cli.Context) error {
	projectName := c.Args().First()
	if projectName == "" {
		return showHelp(c)
	}
	return config.EditProject(projectName)
}

func showHelp(c *cli.Context) error {
	args := append(os.Args[:], "-h")
	return c.App.Run(args)
}

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
			Action:    createNew,
			ArgsUsage: "project_name",
		},
		{
			Name:      "start",
			Usage:     "start a tmux session using a gmux config",
			Action:    start,
			ArgsUsage: "project_name",
		},
		{
			Name:        "stop",
			Usage:       "stops a tmux session",
			Description: "Removes a tmux session by running `tmux kill-session -t sessionname`.",
			ArgsUsage:   "session_name",
			Action:      stop,
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "lists all available gmux projects",
			Action:  list,
		},
		{
			Name:      "edit",
			Usage:     "edit a gmux project configuration",
			ArgsUsage: "project_name",
			Action:    edit,
		},
	}

	// Default action to show the help menu
	app.Action = func(c *cli.Context) error {
		projectName := c.Args().First()
		if projectName != "" {
			return start(c)
		}
		return showHelp(c)
	}
	app.Run(os.Args)
}
