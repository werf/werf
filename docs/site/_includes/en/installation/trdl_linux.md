Make sure you have Git 2.18.0 or newer, gpg and [Docker](https://docs.docker.com/get-docker) installed. To run `werf` as a regular user you should have [access to the Docker daemon](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user).

[Install trdl](https://github.com/werf/trdl/releases/) to `~/bin/trdl`, which will manage `werf` installation and updates. Add `~/bin` to your $PATH.

Add `werf` repo to `trdl`:
```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```
 
To use `werf` on a workstation we recommend setting up `werf` _automatic activation_. For this the activation command should be executed for each new shell session. Often this is achieved by adding the activation command to `~/.bashrc` for Bash or `~/.zshrc` for Zsh, but this depends on the OS/distro/your own configuration. This is the `werf` activation command for the current shell-session:
```shell
source "$(trdl use werf {{ include.version }} {{ include.channel }})"
```

To use `werf` in CI prefer activating `werf` manually instead. For this execute the activation command in the beginning of your CI job, before calling the `werf` binary.

After activation `werf` should be available in the shell-session from which it was activated:
```shell
werf version
```
