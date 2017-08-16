#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# See which shells
has_zsh=$(command -v zsh > /dev/null && echo 1 || echo 0)
shells=$([ $has_zsh -eq 1 ] && echo "bash zsh" || echo "bash")

prompt() {
    read -p "$1 ([y]/n)" -n 1
    echo
    [[ $REPLY =~ ^[Nn]$ ]]
}

insert_source() {
    line="source \"$1\""
    exists=$(grep -F "$line" "$2")
    if [ -z "$exists" ]; then
        echo "Updating $dest:"
        echo -n "  - $line ... "
        echo $line >> $dest
        echo OK
    else
        echo "  already exists ..."
    fi
}

for shell in $shells; do
    src=~/.gmux.${shell}
    echo -n "Generate $src ... "
    echo "PROG=gmux source \"$DIR/autocomplete/${shell}_autocomplete\"" > $src
    echo OK
done

echo

for shell in $shells; do
    prompt "Install $shell completion?"
    answer=$?
    src=~/.gmux.${shell}
    [ $shell = zsh ] && dest=${ZDOTDIR:-~}/.zshrc || dest=~/.bashrc

    if [ $answer -eq 1 ]; then
        insert_source $src $dest
    fi
done
