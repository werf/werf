
## Install Dependencies

1. [Git command line utility](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

   Minimal required version is 1.9.0.

   To optionally use [Git Submodules](https://git-scm.com/docs/gitsubmodules) minimal version is 2.14.0.

## Install Werf

### Way 1 (recommended): using Multiwerf

[Multiwerf](https://github.com/flant/multiwerf) is a version manager for Werf, which:
* downloads werf binary builds;
* manages multiple versions of binaries installed on a single host, that can be used at the same time;
* automatically updates werf binary (can be disabled).

```
mkdir ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
source <(multiwerf use 1.0 beta)
```

### Way 2: download binary

The latest release can be reached via [this page](https://bintray.com/flant/werf/werf/_latestVersion)

##### MacOS

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.1-beta.4/werf-darwin-amd64-v1.0.1-beta.4 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Linux

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.1-beta.4/werf-linux-amd64-v1.0.1-beta.4 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Windows

Download [werf.exe](https://dl.bintray.com/flant/werf/v1.0.1-beta.4/werf-windows-amd64-v1.0.1-beta.4.exe)

### Way 3: from source

```
go get github.com/flant/werf/cmd/werf
```
