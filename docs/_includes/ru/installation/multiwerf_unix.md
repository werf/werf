```shell
# add ~/bin into PATH
export PATH=$PATH:$HOME/bin
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

# install multiwerf into ~/bin directory
mkdir -p ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/werf/multiwerf/master/get.sh | bash
```

##### Использование werf в текущей сессии shell

Следующий вызов создаст shell-функцию `werf`, которая вызывает бинарный файл той версии werf, которую multiwerf скачал и активировал:

```shell
source $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)
werf version
...
```

##### Опционально: автоматически активировать werf при запуске терминала

```shell
echo '. $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)' >> ~/.bashrc
```
