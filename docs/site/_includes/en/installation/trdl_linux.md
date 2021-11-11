Make sure you have Git 2.18.0 or newer and [Docker](https://docs.docker.com/get-docker) installed.

To run `werf` as a regular user, you should have access to the Docker daemon. First, make sure the `docker` group exists:

```shell
sudo groupadd docker
```

Add your current user to the `docker` group:

```
sudo usermod -aG docker $USER
newgrp docker
```

Setup [trdl](https://github.com/werf/trdl) which will manage `werf` installation and updates:
```shell
# Add ~/bin to the PATH.
echo 'export PATH=$HOME/bin:$PATH' >> ~/.bash_profile
export PATH="$HOME/bin:$PATH"

# Install trdl.
curl -L "https://tuf.trdl.dev/targets/releases/0.1.3/linux-{{ include.arch }}/bin/trdl" -o /tmp/trdl
mkdir -p ~/bin
install /tmp/trdl ~/bin/trdl
```

Add `werf` repo:
```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

To use werf locally in your terminal, we recommend to enable its automatic activation. To make werf available in all new shell sessions, you need to execute this command (just once):
```shell
echo 'command -v trdl &>/dev/null && source $(trdl use werf {{ include.version }} {{ include.channel }})' >> ~/.bashrc
```

Now, if you log out and log in to the system again, werf will be always available. You can make sure of that by executing:

```shell
werf version
```

To get werf running in your current terminal only (before any logout/login is done), you can simply execute the `source $(trdl use werf {{ include.version }} {{ include.channel }})` command.

In CI, you need a different approaching with activating `werf` explicitly in the beginning of each job/pipeline by executing:
```shell
source $(trdl use werf {{ include.version }} {{ include.channel }})
```
