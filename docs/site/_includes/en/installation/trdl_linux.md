##### 1. trdl [installation](https://github.com/werf/trdl)

```shell
export PATH=$PATH:$HOME/bin
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

mkdir -p $HOME/bin
curl -LO https://tuf.trdl.dev/targets/releases/0.1.3/linux-{{ include.arch }}/bin/trdl
install ./trdl $HOME/bin/trdl
```

##### 2. Adding official werf TUF-repository into trdl

```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

##### 3. Using werf in the current shell

This will create `werf` shell function which calls to the werf binary which trdl has been prepared for your session:

```shell
source $(trdl use werf {{ include.version }} {{ include.channel }})
werf version
...
```

##### 4. Optional: activate werf on terminal startup

```shell
echo '. $(trdl use werf {{ include.version }} {{ include.channel }})' >> ~/.bashrc
```
