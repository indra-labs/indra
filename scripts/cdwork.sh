#!/usr/bin/zsh
# populate ~/.workpath thus: `cd path/to/repo/root; pwd>~/.workpath
export INDRAROOT=$(cat ~/.workpath)
export PATH=$INDRAROOT/scripts:$PATH
# put the path of the root of the repository in ./scripts/path
cd $INDRAROOT
zsh
