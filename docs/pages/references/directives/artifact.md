---
title: Проброска кода в собираемый образ - artifact
sidebar: doc_sidebar
permalink: artifact.html
folder: directive
---

Артефакт (artifact) — это набор правил для сборки образа с файловым ресурсом, который импортируется в собираемый образ.
Для добавления артефакта можно использовать поддирективу export или директиву import.

Поддержка сборщиками

- Shell - нет
- Chef - да
- Ansible - нет

## Правила использования

* Пути добавления не должны пересекаться между артефактами.
* Изменение любого параметра артефакта ведёт к смене сигнатур, пересборке связанных [стадий](stages.html#стадия) приложения.
* Приложение может содержать любое количество артефактов.

## Ansible-сборщик

@todo: описать

## Chef-сборщик

Директива `git [<url>]` позволяет определить один или несколько git-артефактов.

* Поддерживается два типа git-артефактов, local и remote.
* Необязательный параметр \<url\> соответствует адресу удалённого git-репозитория (remote).
* Для добавления git-артефакта необходимо использовать поддирективу add.
  * Принимает параметр \<cwd\> (по умолчанию используется '\\').
  * Параметры \<include_paths\>, \<exclude_paths\>, \<owner\>, \<group\>, \<to\> определяются в контексте.
  * Параметры \<branch\>, \<commit\> могут быть определены в контексте remote git-артефакта.
* В контексте директивы можно указать базовые параметры git-артефактов, которые могут быть переопределены в контексте каждого из них.
  * \<owner\>.
  * \<group\>.
  * \<branch\>.
  * \<commit\>.

### Параметры артефакта

#### Общие
* to: абсолютный путь, в который будут копироваться ресурсы.
* cwd: абсолютный путь, определяет рабочую директорию.
* include_paths: добавить только указанные относительные пути.
* exclude_paths: игнорировать указанные относительные пути.
* owner: определить пользователя.
* group: определить группу.
* Принимает параметр \<cwd\> (по умолчанию используется путь \<to\>).
* Параметры \<include_paths\>, \<exclude_paths\>, \<owner\>, \<group\>, \<to\> определяются в контексте.
* Параметр \<to\> по умолчанию соответствует \<cwd\>.
* Если собирается не scratch: в контексте обязательно использование хотя бы одной из директив **before** или **after**, где:
    * директива определяет порядок применения артефакта (до или после);
    * значение определяет стадию (install или setup).

### Примеры

##### export
```ruby
dimg do
  docker.from 'image:tag'

  artifact do
    shell.build_artifact.run 'command1', 'command2'

    export '/artifact' do
      before 'install'
      to '/app'
    end
  end

  artifact do
    shell.build_artifact.run 'command3', 'command4'

    export '/artifact/assets' do
      after 'setup'
      to '/app'
    end
  end
end
```

##### import
```ruby
dimg do
  docker.from 'image:tag'

  artifact('artifact-a') do
    shell.build_artifact.run 'command1', 'command2'
  end

  artifact('artifact-b') do
    shell.build_artifact.run 'command3', 'command4'
  end

  import('artifact-a', '/artifact') do
    before 'install'
    to '/app'
  end

  import('artifact-b', '/artifact/assets') do
    after 'setup'
    to '/app'
  end
end
```

#### Собрать scratch образ
```ruby
dimg_group do
  artifact do
    docker.from 'image:tag'
    shell.build_artifact.run 'command1', 'command2'

    export '/' do
      to '/app'
    end
  end

  dimg
end
```
