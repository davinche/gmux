package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/davinche/gmux/config"
	"github.com/urfave/cli"
)

func init() {
	startServer()
}

// New handles the creation of a new gmux config
func New(c *cli.Context) error {
	configName := c.Args().First()
	if configName == "" {
		return ShowHelp(c)
	}
	if config.Exists(configName) {
		return fmt.Errorf("config with the same name already exists")
	}

	newConfig := config.New(configName)
	if err := newConfig.Write(); err != nil {
		return err
	}
	return config.Edit(configName)
}

// Edit opens a gmux configuration inside the user's editor
func Edit(c *cli.Context) error {
	configName := c.Args().First()
	if configName == "" {
		return ShowHelp(c)
	}
	return config.Edit(configName)
}

// Start handles running a gmux config
func Start(c *cli.Context) error {
	configName := c.Args().First()
	if configName == "" {
		return ShowHelp(c)
	}

	if hasSession(configName) {
		if err := config.AttachToSession(configName); err != nil {
			return cli.NewExitError(fmt.Sprintf("could not attach to session %q", configName), 1)
		}
	}

	if err := config.GetAndRun(configName, c.Bool("debug")); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

// Stop handles terminating a tmux connection
func Stop(c *cli.Context) error {
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
	return cmd.Run()
}

// List shows all available gmux configurations
func List(c *cli.Context) error {
	return config.List()
}

// ShowHelp shows the help for the given command
func ShowHelp(c *cli.Context) error {
	args := append(os.Args[:], "-h")
	return c.App.Run(args)
}

// ----------------------------------------------------------------------------
// TMUX Helpers ---------------------------------------------------------------
// ----------------------------------------------------------------------------
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
