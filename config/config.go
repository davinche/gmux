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

	"encoding/json"

	"github.com/davinche/gmux/command"
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
			err = os.Mkdir(configDir, 0755)
			if err != nil {
				os.Stderr.WriteString(fmt.Sprintf(
					"error: could not create gmux config directory: err=%q", err))
				os.Exit(1)
			}
		} else {
			// Unknown error occured for our configs directory
			os.Stderr.WriteString(fmt.Sprintf(
				"error: could not determine gmux config directory: err=%q", err))
			os.Exit(1)
		}

	}
}

// Config represents the top level structure of a gmux config
type Config struct {
	Name          string
	Root          string
	Windows       []*Window
	PreWindow     string `json:",omitempty"`
	StartupWindow string `json:",omitempty"`
	StartupPane   int    `json:",omitempty"`
}

// Window represents the configration for a tmux window
type Window struct {
	Name   string
	Layout string   `json:",omitempty"`
	Root   string   `json:",omitempty"`
	Panes  []string `json:",omitempty"`
}

// Config Methods -------------------------------------------------------------

// Exec runs the gmux configuration
func (c *Config) Exec(debug bool) error {
	cc := &command.Chain{Debug: debug}

	// CD to tmux config directory
	rootAbs, err := filepath.Abs(expandPath(c.Root))
	if err != nil {
		if debug {
			log.Printf("error: could not determine absolute path to config directory: err=%q\n", err)
		}
		return err
	}
	if err := os.Chdir(rootAbs); err != nil {
		if debug {
			log.Printf("error: could not change directory to config root: err=%q; dir=%q\n", err, c.Root)
		}
		return err
	}

	// Create the tmux session
	firstWindowRoot := rootAbs
	if c.Windows[0].Root != "" {
		firstWindowRoot = expandPath(c.Windows[0].Root)
	}
	cc.Add("tmux", "start-server")
	cc.Add("tmux", "new-session", "-d", "-s", c.Name, "-n", c.Windows[0].Name, "-c", firstWindowRoot)

	// Create the windows
	for idx, w := range c.Windows {
		winID := fmt.Sprintf("%s:%d", c.Name, idx)
		wRoot := rootAbs
		if w.Root != "" {
			wRoot = expandPath(w.Root)
		}
		wRoot = escapePath(wRoot)

		// First window is created automatically, so only create a new window if we're not
		// looking at the first one
		if idx != 0 {
			cc.Add("tmux", "new-window", "-t", winID, "-n", w.Name, "-c", wRoot)
		}

		// Create Panes
		for idx, p := range w.Panes {
			paneID := fmt.Sprintf("%s.%d", winID, idx)

			// Likewise, first pane is created automatically
			// so only "split window" for subsequent panes
			if idx != 0 {
				cc.Add("tmux", "split-window", "-t", winID, "-c", wRoot)
			}

			// Execute a pre_window command if one is provided
			if c.PreWindow != "" {
				cc.Add("tmux", "send-keys", "-t", paneID, c.PreWindow, "Enter")
			}

			// execute the command for a particular pane if it is provided
			if p != "" {
				cc.Add("tmux", "send-keys", "-t", paneID, p, "Enter")
			}
		}

		// Set window layout
		wLayout := "tiled"
		if w.Layout != "" {
			wLayout = w.Layout
		}
		cc.Add("tmux", "select-layout", "-t", winID, wLayout)
	}

	// Select Starting Window
	selectWindow := fmt.Sprintf("%s:0", c.Name)
	if c.StartupWindow != "" {
		selectWindow = fmt.Sprintf("%s:%s", c.Name, c.StartupWindow)
	}
	cc.Add("tmux", "select-window", "-t", selectWindow)
	cc.Add("tmux", "select-pane", "-t", fmt.Sprintf("%s.%d", selectWindow, c.StartupPane))

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

// Write the config to the configurations directory
func (c *Config) Write() error {
	filePath := getConfigFilePath(c.Name)
	fmt.Println(filePath)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0644)
}

// New returns a new gmux configuration
func New(configName string) *Config {
	config := &Config{
		Name:    configName,
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

	return config
}

// Get returns the config for a given config name
func Get(config string) (*Config, error) {
	c := &Config{}
	if !Exists(config) {
		return nil, fmt.Errorf("could not find config: %s", config)
	}

	fileBytes, err := ioutil.ReadFile(getConfigFilePath(config))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(fileBytes, c)
	if err != nil {
		return nil, err
	}

	if c.Name == "" || c.Root == "" || c.Windows == nil {
		return nil, fmt.Errorf("invalid config: missing name, root, or windows in the file")
	}
	return c, nil
}

// GetAndRun gets a projects config and executes it
func GetAndRun(config string, debug bool) error {
	c, err := Get(config)
	if err != nil {
		return err
	}
	return c.Exec(debug)
}

// List prints out the list of gmux projects
func List() error {
	files, err := ioutil.ReadDir(configDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		ext := filepath.Ext(name)
		fmt.Printf("%s\n", name[:len(name)-len(ext)])
	}
	return nil
}

// Edit uses the environment's EDITOR to edit the config
func Edit(config string) error {
	editorStr := os.Getenv("EDITOR")
	if editorStr == "" {
		return fmt.Errorf("EDITOR variable not defined in env")
	}

	if !Exists(config) {
		return fmt.Errorf("could not find config: %s", config)
	}

	editor, err := exec.LookPath(editorStr)
	if err != nil {
		return err
	}

	if err := syscall.Exec(editor,
		[]string{editorStr, getConfigFilePath(config)},
		os.Environ()); err != nil {
		return err
	}
	return nil
}

// Delete an existing gmux config
func Delete(config string) error {
	configFile := getConfigFilePath(config)
	return os.RemoveAll(configFile)
}

// Exists check if a gmux config already exists
func Exists(config string) bool {
	configFile := getConfigFilePath(config)
	_, err := os.Stat(configFile)
	return err == nil
}

// AttachToSession attempts to attach to a a currently active tmux session
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

// returns the path to the config given the config name
func getConfigFilePath(configName string) string {
	return path.Join(configDir, fmt.Sprintf("%s.json", configName))
}

func escapePath(path string) string {
	return strings.Replace(path, " ", "\\ ", -1)
}
