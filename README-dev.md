# How to build

0. `$GOROOT` and `$GOPATH` should be set before the build.

1. Clone dapp source code to the right place in `$GOPATH`:

    ```
    mkdir -p $GOPATH/src/github.com/flant
    git clone https://github.com/flant/dapp.git $GOPATH/src/github.com/flant/dapp
    cd $GOPATH/src/github.com/flant/dapp
    ```

2. Setup development environment variables:

    ```
    source ./go-env
    ```
    
    This will make `ruby-dapp` work properly with the newly built `dappfile-yml`.

3. Download go dependencies:

    ```
    ./go-get.sh
    ```

3. Run build:

    ```
    ./go-build.sh
    ```

The `dappfile-yml` binary will be placed into your `$GOPATH/bin`.
To call it you can use `$DAPP_BIN_DAPPFILE_YML` (used by `ruby-dapp` internally) or simply `dappfile-yml`.
