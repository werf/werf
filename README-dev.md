# How to build

0. $GOROOT and $GOPATH should be set at that moment.

1. Clone dapp source code and place into the right place in $GOPATH:

```
mkdir -p $GOPATH/src/github.com/flant
git clone https://github.com/flant/dapp.git $GOPATH/src/github.com/flant/dapp
cd $GOPATH/src/github.com/flant/dapp
```

2. Setup development environment variables:

```
source ./go-env
```

This will make ruby-dapp work properly with newly builded dappfile-yml.

3. Download go dependencies:

```
./go-get.sh
```

3. Run build:

```
./go-build.sh
```

dappfile-yml binary will be placed into your $GOPATH/bin. To call you can use `$DAPP_BIN_DAPPFILE_YML` (used by ruby-dapp internally) or simply `dappfile-yml`.
