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

3. Run build (will install dappfile-yml into $GOROOT/bin):

```
./go-build.sh
```
