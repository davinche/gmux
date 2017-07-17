# GMux
Another Tmux Manager... but in Golang

## Installation

### Go

~~~
go get github.com/davinche/gmux
~~~

### Mac

~~~
brew install davinche/tools/gmux
~~~

## Usage

### Editor

Like tmuxinator, gmux uses your shell's **EDITOR** to edit configurations.  
A default editor can specified via `export EDITOR=youreditor`

### Shell Completion

An install script is provided to enable gmux completion. When installed via homebrew, the script is located under `/usr/local/Cellar/gmux/{VERSION}/install_completion.sh`

Running it will generate a shell completion file under your `$HOME` directory by the name `.gmux.bash` and/or `.gmux.zsh`.  
Source these files files in your `.bashrc` or `.zshrc` to enable completion.

The installer will provide the option to add these sources to your .shellrc automatically.

## Help for CLI Usage

Once installed, usage information can be viewed via `gmux --help`.

## Configuration

To create a new profile, just run the following command.

~~~
gmux new <name>
~~~

This will create a new file in `$HOME/.gmux/<name>`.

### Example:

```json

{
  "Name": "SecretProject",
  "Root": "~/SecretProject",
  "PreWindow": "nvm use default",
  "StartupWindow": "editor",
  "StartupPane": 0,
  "Windows": [
    {
      "Name": "editor",
      "Layout": "tiled",
      "Panes": [
        "nvim .",
        "pm2 start server.js --watch"
      ]
    },
    {
        "Name": "tests",
        "Root": "~/SecretProject/tests",
        "Panes": [
            "jest --watchAll",
            ""
        ]
    }
  ]
}
```

### Options

#### Root Level ####

| Name          | Type      | Description                                        |
|:--------------|:----------|:---------------------------------------------------|
| Name          | string    | The name of your tmux session                      |
| Root          | string    | The working directory for your tmux session        |
| PreWindow     | string    | A command you want run at the start of each window |
| StartupWindow | string    | The window to focus on after session creation      |
| StartupPane   | number    | The pane to focus on (starts from 0)               |
| Windows       | []Windows | An array of configurations for each window         |


#### Window Object ####

| Name   | Type     | Desc                                          |
|:-------|:---------|:----------------------------------------------|
| Name   | string   | The name of the window                        |
| Root   | string   | The working directory for your window         |
| Layout | string   | The way you want the panes to be laid out     |
| Panes  | []string | List of commands you want to run in each pane |


## About

Gmux is heavily inspired by [tmuxinator][tmuxinator]. For the time being, use Tmuxinator if you want a more featureful Tmux manager. Currently Gmux only offers a basic subset of tmuxinator's capabilities.

[tmuxinator]: https://github.com/tmuxinator/tmuxinator


## License:

MIT
