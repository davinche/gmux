package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli"
)

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

func stop(sessionName string) {
	if sessionName == "" {
		cmd := exec.Command("tmux", "display-message", "-p", "#S")
		output, err := cmd.Output()
		if err != nil {
			os.Stderr.WriteString("could not determine current tmux session")
			os.Exit(1)
		}
		sessionName = strings.TrimSpace(string(output))
	}
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func main() {
	app := cli.NewApp()
	app.Name = "GMux"
	app.Usage = "a tmux sessions manager"
	app.Run(os.Args)
}
