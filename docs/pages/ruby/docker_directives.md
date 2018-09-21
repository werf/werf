---
title: Docker директивы
sidebar: ruby
permalink: ruby/docker_directives.html
---

### docker.from \<image\>[, cache_version: \<cache_version\>]
Определить окружение приложения **\<image\>**).

* **\<image\>** имеет следующий формат 'REPOSITORY:TAG'.
* Опциональный параметр **\<cache_version\>** участвует в формировании сигнатуры стадии.

### docker.cmd \<cmd\>[, \<cmd\> ...]
Применить dockerfile инструкцию CMD (см. [CMD](https://docs.docker.com/engine/reference/builder/#/cmd "Docker reference")).

### docker.env \<env_name\>: \<env_value\>[, \<env_name\>: \<env_value\> ...]
Применить dockerfile инструкцию ENV (см. [ENV](https://docs.docker.com/engine/reference/builder/#/env "Docker reference")).

### docker.entrypoint \<cmd\>[, \<arg\> ...]
Применить dockerfile инструкцию ENTRYPOINT (см. [ENTRYPOINT](https://docs.docker.com/engine/reference/builder/#/entrypoint "Docker reference")).

### docker.expose \<expose\>[, \<expose\> ...]
Применить dockerfile инструкцию EXPOSE (см. [EXPOSE](https://docs.docker.com/engine/reference/builder/#/expose "Docker reference")).

### docker.label \<label_key\>: \<label_value\>[, \<label_key\>: \<label_value\> ...]
Применить dockerfile инструкцию LABEL (см. [LABEL](https://docs.docker.com/engine/reference/builder/#/label "Docker reference")).

### docker.onbuild \<cmd\>[, \<cmd\> ...]
Применить dockerfile инструкцию ONBUILD (см. [ONBUILD](https://docs.docker.com/engine/reference/builder/#/onbuild "Docker reference")).

### docker.user \<user\>
Применить dockerfile инструкцию USER (см. [USER](https://docs.docker.com/engine/reference/builder/#/user "Docker reference")).

### docker.volume \<volume\>[, \<volume\> ...]
Применить dockerfile инструкцию VOLUME (см. [VOLUME](https://docs.docker.com/engine/reference/builder/#/volume "Docker reference")).

### docker.workdir \<path\>
Применить dockerfile инструкцию WORKDIR (см. [WORKDIR](https://docs.docker.com/engine/reference/builder/#/workdir "Docker reference")).

### Примеры

#### Собрать с базовым образом "ubuntu:16.04" и несколькими dockerfile-инструкциями
```ruby
dimg do
  docker do
    from 'ubuntu:16.04'

    env EDITOR: 'vim', LANG: 'he_IL.UTF-8'
    user 'user3:stuff'
  end
end
```

#### Собрать с базовым образом "ubuntu:16.04" и несколькими dockerfile-инструкциями (строчная запись)
```ruby
dimg do
  docker.from 'ubuntu:16.04'

  docker.env EDITOR: 'vim', LANG: 'he_IL.UTF-8'
  docker.user 'user3:stuff'
end
```
