Убедитесь, что [Git](https://git-scm.com/download/mac) версии 2.18.0 или новее и [Docker](https://docs.docker.com/get-docker) установлены.

Установите [trdl](https://github.com/werf/trdl), который будет отвечать за установку и обновление `werf`:

```shell
# Добавьте ~/bin в PATH.
echo 'export PATH=$HOME/bin:$PATH' >> ~/.zprofile
export PATH="$HOME/bin:$PATH"

# Установите trdl.
curl -L "https://tuf.trdl.dev/targets/releases/0.1.3/darwin-{{ include.arch }}/bin/trdl" -o /tmp/trdl
mkdir -p ~/bin
install /tmp/trdl ~/bin/trdl
```

Добавьте репозиторий с `werf`:
```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

Для локального использования werf (в терминале) мы рекомендуем настроить автоматическую активацию утилиты. Чтобы werf была доступна во всех новых shell-сессиях, выполните следующую команду (это потребуется сделать лишь один раз):
```shell
echo 'source $(trdl use werf {{ include.version }} {{ include.channel }})' >> ~/.zshrc
```

Теперь, если вы выйдете из системы и залогинитесь в неё обратно, werf всегда будет доступна. Убедиться в этом можно следующей командой:

```shell
werf version
```

Чтобы получить werf только в текущем терминале (до того, как перезашли в систему), достаточно выполнить команду `source $(trdl use werf {{ include.version }} {{ include.channel }})`.

Для CI рекомендуется другой подход с явной активацией `werf` в начале каждого job/pipeline. Она выполняется командой:

```shell
source $(trdl use werf {{ include.version }} {{ include.channel }})
```
