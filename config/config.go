package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/davinche/gmux/command"

	yaml "gopkg.in/yaml.v2"
)

// User's Home Directory
var userDir string

// GMux Config Directory ($HOME/.gmux)
var configDir string

func init() {
	if cUser, err := user.Current(); err == nil {
		userDir = cUser.HomeDir
	} else if homeDir := os.Getenv("HOME"); userDir == "" {
		userDir = homeDir
	} else {
		log.Fatalf("error: could not determine user home")
	}

	if !strings.HasSuffix(userDir, "/") {
		userDir += "/"
	}

	// Create our config directory inside the user's home directory
	configDir = path.Join(userDir, ".gmux/")
	fInfo, err := os.Stat(configDir)

	// make sure it's not a file..
	if err == nil && !fInfo.IsDir() {
		os.Stderr.WriteString(fmt.Sprintf("error: %s is not a directory", configDir))
		os.Exit(1)
	}

	if err != nil {
		// Create our configs directory if it doesn't exist
		if os.IsNotExist(err) {
			err = os.Mkdir(configDir, os.ModeDir)
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf(
					"error: could not create gmux config directory: err=%q", err))
				os.Exit(1)
			}
		}

		// Unknown error occured for our configs directory
		os.Stderr.WriteString(fmt.Sprintf(
			"error: could not determine gmux config directory: err=%q", err))
		os.Exit(1)
	}
}

// Config represents the top level structure of a gmux config
type Config struct {
	Name    string
	Root    string
	Windows []*Window
}

// Window represents the configration for a tmux window
type Window struct {
	Name   string
	Layout string   `yaml:",omitempty"`
	Root   string   `yaml:",omitempty"`
	Panes  []string `yaml:",omitempty"`
}

// Config Methods -------------------------------------------------------------
// Exec runs the tmux configuration
func (c *Config) Exec(debug bool) error {
	cc := &command.Chain{Debug: debug}

	// CD to tmux project directory
	rootAbs, err := filepath.Abs(expandPath(c.Root))
	if err != nil {
		if debug {
			log.Printf("error: could not determine absolute path to project directory: err=%q\n", err)
		}
		return err
	}
	if err := os.Chdir(rootAbs); err != nil {
		if debug {
			log.Printf("error: could not change directory to project root: err=%q; dir=%q\n", err, c.Root)
		}
		return err
	}

	// Create the tmux session
	cc.Add("tmux", "start-server")
	cc.Add("tmux", "new-session", "-d", "-s", c.Name, "-n", c.Windows[0].Name)

	// Create the windows
	for idx, w := range c.Windows {
		winID := fmt.Sprintf("%s:%d", c.Name, idx)
		wRoot := rootAbs
		if w.Root != "" {
			wRoot = expandPath(w.Root)
		}
		quotedWRoot := fmt.Sprintf("%q", wRoot)

		// First window is created automatically, so only create a new window if we're not
		// looking at the first one
		if idx != 0 {
			cc.Add("tmux", "new-window", "-t", winID, "-n", w.Name, "-c", quotedWRoot)
		}

		// Set window layout
		wLayout := "tiled"
		if w.Layout != "" {
			wLayout = w.Layout
		}
		cc.Add("tmux", "select-layout", "-t", winID, wLayout)

		// Create Panes
		for idx, p := range w.Panes {
			paneID := fmt.Sprintf("%s.%d", winID, idx)

			// Likewise, first pane is created automatically
			// so only "split window" for subsequent panes
			if idx != 0 {
				cc.Add("tmux", "split-window", "-t", winID, "-c", quotedWRoot)
			}

			// execute the command for a particular pane if it is provided
			if p != "" {
				cc.Add("tmux", "send-keys", "-t", paneID, p, "Enter")
			}
		}

	}

	// select the first window and first pane
	firstWindow := fmt.Sprintf("%s:0", c.Name)
	cc.Add("tmux", "select-window", "-t", firstWindow)
	cc.Add("tmux", "select-pane", "-t", fmt.Sprintf("%s.0", firstWindow))

	// Run our tmux script
	if err := cc.Run(); err != nil {
		return err
	}

	if err := AttachToSession(c.Name); err != nil {
		if debug {
			log.Printf("error: could not attach to session: %q\n", err)
		}
		return err
	}
	return nil
}

// New returns a new gmux configuration
func New(project string) (*Config, error) {
	config := &Config{
		Name:    project,
		Root:    "~/",
		Windows: make([]*Window, 3),
	}

	config.Windows[0] = &Window{
		Name:   "editor",
		Layout: "main-vertical",
		Panes: []string{
			"vim",
			"guard",
		},
	}

	config.Windows[1] = &Window{
		Name: "server",
		Panes: []string{
			"bundle exec rails s",
		},
	}

	config.Windows[2] = &Window{
		Name: "logs",
		Panes: []string{
			"tail -f log/development.log",
		},
	}

	return config, nil
}

// Get returns the config for a given project name
func Get(project string) (*Config, error) {
	c := &Config{}
	configPath := path.Join(configDir, fmt.Sprintf("%s.yml", project))
	file, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	if file.IsDir() {
		return nil, fmt.Errorf("invalid config: file is a directory")
	}

	fileBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(fileBytes, c)
	if err != nil {
		return nil, err
	}

	if c.Name == "" || c.Root == "" || c.Windows == nil {
		return nil, fmt.Errorf("invalid config: missing name, root, or windows in the file")
	}
	return c, nil
}

// GetAndRun gets a projects config and executes it
func GetAndRun(project string, debug bool) error {
	c, err := Get(project)
	if err != nil {
		return err
	}
	return c.Exec(debug)
}

func AttachToSession(name string) error {
	// Replace current context with tmux attach session
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return err
	}
	args := []string{"tmux"}

	// Attach to the session if we're not already in tmux.
	// Otherwise, switch from our current session to the new one
	if os.Getenv("TMUX") == "" {
		args = append(args, "-u", "attach-session", "-t", name)
	} else {
		args = append(args, "-u", "switch-client", "-t", name)
	}

	// Replace our program context with tmux
	if sysErr := syscall.Exec(tmux, args, os.Environ()); sysErr != nil {
		return err
	}
	return nil
}

// perform any path expansions the shell would normally do for us
func expandPath(p string) string {
	newP := p
	if strings.HasPrefix(newP, "~/") {
		p = strings.Replace(p, "~/", userDir, 1)
	}
	return p
}
