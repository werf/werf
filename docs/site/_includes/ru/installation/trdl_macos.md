##### 1. Установка [trdl](https://github.com/werf/trdl)

```shell
export PATH=$PATH:$HOME/bin
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bash_profile

mkdir -p $HOME/bin
curl -LO https://tuf.trdl.dev/targets/releases/0.1.3/darwin-{{ include.arch }}/bin/trdl
install ./trdl $HOME/bin/trdl
```

##### 2. Добавление официального TUF-репозитория werf в trdl

```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

##### 3. Использование werf в текущей сессии shell

Следующий вызов добавит в текущий PATH путь до той версии werf, которую trdl скачал и активировал:

```shell
source $(trdl use werf {{ include.version }} {{ include.channel }})
werf version
...
```

##### 4. Опционально: автоматически активировать werf при запуске терминала

```shell
echo '. $(trdl use werf {{ include.version }} {{ include.channel }})' >> ~/.bash_profile
```
