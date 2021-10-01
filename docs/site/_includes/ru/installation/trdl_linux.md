Убедитесь, что Git версии 2.18.0 или новее и [Docker](https://docs.docker.com/get-docker) установлены.

Чтобы обычный пользователь мог запустить `werf`, пользователь должен иметь доступ к Docker-сервису. Cоздайте пользовательскую группу `docker`. Обратите внимание, что Docker при установке сам создает эту группу, и указанная ниже команда нужна для подстраховки. 

```shell
sudo groupadd docker
```

Если после выполнения появилось сообщение `groupadd: group 'docker' already exists`, значит такая группа уже есть. 

Добавьте своего текущего пользователя в группу `docker`.

```shell
sudo usermod -aG docker $USER
newgrp docker
```

Установите [trdl](https://github.com/werf/trdl), который будет отвечать за установку и обновление `werf`:

```shell
# Добавьте ~/bin в PATH.
echo 'export PATH=$HOME/bin:$PATH' >> ~/.bash_profile
export PATH="$HOME/bin:$PATH"

# Установите trdl.
curl -L "https://tuf.trdl.dev/targets/releases/0.1.3/linux-{{ include.arch }}/bin/trdl" -o /tmp/trdl
mkdir -p ~/bin
install /tmp/trdl ~/bin/trdl
```

Добавьте `werf` репозиторий:
```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

Активируйте `werf` и добавьте автоматическую активацию для новых shell-сессий:

```shell
source $(trdl use werf {{ include.version }} {{ include.channel }})
echo 'source $(trdl use werf {{ include.version }} {{ include.channel }})' >> ~/.bashrc
```

Для CI рекомендуем активировать `werf` явно в начале каждого job/pipeline, выполняя:

```shell
source $(trdl use werf {{ include.version }} {{ include.channel }})
```

Убедитесь, что `werf` теперь доступен в командной строке (начните новую shell-сессию, если вы предпочли автоматическую активацию):
```shell
werf version
```
