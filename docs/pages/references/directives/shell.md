---
title: Выполнение произвольного кода
sidebar: reference
permalink: shell.html
folder: directive
---

При сборке образов зачастую хочется просто выполнить shell-команду.

Поддержка сборщиками

- Shell - нет
- Chef - да
- Ansible - да

## Ansible-сборщик

В Ansible для этого есть директива `raw`.

@todo: добавить короткий пример

## Chef-сборщик

Следующие поддирективы позволяют добавить bash-команды для выполнения на соответствующих стадиях [shell образа](definitions.html#shell-проект):

* before_install.
* before_setup.
* install.
* setup.
* build_artifact (в случае, если директива используется в артефакте).

Можно определить version, который участвует в формировании сигнатуры стадии.

* Указать базовый, для всех стадий, можно в контексте shell.
* Указать или переопределить базовый можно в контексте, соответствующем стадии.

### Примеры

#### Собрать с bash-командами на стадиях before_install и setup, указав версию для всех и переопределив для setup
```ruby
dimg do
  docker.from 'image:tag'

  shell do
    version '1'

    before_install do
      run 'command1', 'command2'
    end

    setup do
      run 'command3', 'command4'
      version '2'
    end
  end
end
```

#### Собрать с bash-командами на стадиях before_install и setup, указав версию для всех и переопределив для setup (строчная запись)
```ruby
dimg do
  docker.from 'image:tag'

  shell.version '1'
  shell.before_install.run 'command1', 'command2'
  shell.setup.run 'command3', 'command4'
  shell.setup.version '2'
end
```
