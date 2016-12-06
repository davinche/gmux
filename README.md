# GMUX
Another Tmux Manager... but in Golang

## Installation

Mac: `brew install davinche/tools/gmux`

## Editor

Like tmuxinator, gmux uses your shell's EDITOR to edit gmux configurations.  
A default editor can specified via `export EDITOR=youreditor`

## Completion

An install script is provided to enable gmux completion. When installed via homebrew, the script is located under `/usr/local/Cellar/gmux/{VERSION}/install_completion.sh`

Running it will generate a shell completion file under your `$HOME` directory by the name `.gmux.bash` and/or `.gmux.zsh`.  
Source these files files in your `.bashrc` or `.zshrc` to enable completion.

The installer will provide the option to add these sources to your shellrc automatically.

## Usage

Once installed, usage information can be viewed via `gmux --help`.

## About

Gmux is heavily inspired by [tmuxinator][tmuxinator]. For the time being, use Tmuxinator if you want a more featureful Tmux manager. Currently Gmux only offers a basic subset of tmuxinator's capabilities.

[tmuxinator]: https://github.com/tmuxinator/tmuxinator
