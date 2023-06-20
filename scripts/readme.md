# scripts

populate ~/.workpath thus: `cd path/to/repo/root; pwd>~/.workpath`

add this to your `~/.bashrc` or `~/.zshrc`:

    export PATH=$(cat ~/.workpath)/scripts:$HOME/sdk/go1.19.10/bin:$PATH 
    export GOBIN=$HOME/.local/bin

`source` the `rc` file or open a new terminal session and type `cdwork.sh` and you will have a number of useful commands
that handle paths and special build parameters to make the code locations work without hard coding them anywhere.