package main

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
)

func startServer() {
	cmd := exec.Command("tmux", "start-server")
	if err := cmd.Run(); err != nil {
		os.Stderr.WriteString("could not start tmux server")
		os.Exit(1)
	}
}

func getSessions() map[string]struct{} {
	sessions := make(map[string]struct{})

	cmd := exec.Command("tmux", "list-sessions", "-F", "'#{session_name}'")
	output, err := cmd.Output()
	// no active sessions, return the empty map
	if err != nil {
		return sessions
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(output))
	for scanner.Scan() {
		sessions[scanner.Text()] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		os.Stderr.WriteString("could not scan list of sessions")
		os.Exit(1)
	}
	return sessions
}

func main() {
	startServer()
	sessions := getSessions()
}
