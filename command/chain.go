package command

import (
	"log"
	"os/exec"
	"strings"
)

// Chain contains a list of commands to run consecutively
type Chain struct {
	commands [][]string
	Debug    bool
}

// Add to the chain of commands
func (c *Chain) Add(args ...string) {
	c.commands = append(c.commands, args)
}

// Run the chain of commands
func (c *Chain) Run() error {
	for _, command := range c.commands {
		if c.Debug {
			log.Printf("debug: executing: %s", strings.Join(command, " "))
		}
		var cmd *exec.Cmd
		if len(command) == 0 {
			cmd = exec.Command(command[0])
		} else {
			cmd = exec.Command(command[0], command[1:]...)
		}
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
