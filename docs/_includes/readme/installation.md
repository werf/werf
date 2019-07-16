
## Install Dependencies

### Docker
   
[Docker CE installation guide](https://docs.docker.com/install/).

Manage Docker as a non-root user. Create the **docker** group and add your user to the group: 
```bash
sudo groupadd docker
sudo usermod -aG docker $USER
```

### Git command line utility

[Git installation guide](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

- Minimal required version is 1.9.0.
- To optionally use [Git Submodules](https://git-scm.com/docs/gitsubmodules) minimal version is 2.14.0.
   
## Install Werf

### Method 1 (recommended): using Multiwerf

[Multiwerf](https://github.com/flant/multiwerf) is a version manager for Werf, which:
* downloads Werf binary builds;
* manages multiple versions of binaries installed on a single host, that can be used at the same time;
* automatically updates Werf binary (can be disabled).

```bash
mkdir ~/bin
cd ~/bin

## add ~/bin in PATH if not there
echo  ‘export PATH=$PATH:$HOME/bin’ >> ~/.bashrc 
exec bash

curl -L https://raw.githubusercontent.com/flant/multiwerf/master/get.sh | bash
source <(multiwerf use 1.0 beta)
```

### Method 2: download binary

The latest release can be reached via [this page](https://bintray.com/flant/werf/werf/_latestVersion)

##### MacOS

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.3-beta.6/werf-darwin-amd64-v1.0.3-beta.6 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Linux

```bash
curl -L https://dl.bintray.com/flant/werf/v1.0.3-beta.6/werf-linux-amd64-v1.0.3-beta.6 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

##### Windows

Download [werf.exe](https://dl.bintray.com/flant/werf/v1.0.3-beta.6/werf-windows-amd64-v1.0.3-beta.6.exe)

### Method 3: from source

```
go get github.com/flant/werf/cmd/werf
```
