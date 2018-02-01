---
title: Директивы артефактов и git-артефактов
sidebar: doc_sidebar
permalink: artifact_directives.html
folder: directive
---

### Особенности

* Пути добавления не должны пересекаться между артефактами.
* Изменение любого параметра артефакта ведёт к смене сигнатур, пересборке связанных [стадий](stages.html#стадия) приложения.
* Приложение может содержать любое количество артефактов.

### Параметры артефакта

#### Общие
* to: абсолютный путь, в который будут копироваться ресурсы.
* cwd: абсолютный путь, определяет рабочую директорию.
* include_paths: добавить только указанные относительные пути.
* exclude_paths: игнорировать указанные относительные пути.
* owner: определить пользователя.
* group: определить группу.

#### Дополнительные для remote git-артефакта
* branch: определить branch.
* commit: определить commit.

### git \[\<url\>\]
Директива позволяет определить один или несколько git-артефактов.

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

#### git.add.stage_dependencies
Директива позволяет определить для стадий install, before_setup, setup и build_artifact зависимости от файлов git-артефакта.

* При изменении содержимого указанных файлов, произойдет пересборка зависимой стадии.
* Учитывается содержимое и имена файлов.
* Поддерживаются glob-паттерны.
* Пути в \<glob\> указываются относительно cwd git-артефакта.
* Директории игнорируются.
* \<glob\> чувствителен к регистру.  

##### Примеры

###### Собрать с несколькими git-артефактами
```ruby
dimg do
  docker.from 'image:tag'

  git do
    add '/' do
      exclude_paths 'assets'
      to '/app'
    end

    add '/assets' do
      to '/web/site.narod.ru/assets_with_strange_name'
    end
  end

  git 'https://site.com/com/project.git' do
    owner 'user4'
    group 'stuff'

    add '/' do
      to '/project'
    end
  end
end
```

###### Определить зависимости для нескольких git-артефактов
```ruby
dimg do
  docker.from 'image:tag'

  git do
    add '/' do
      to '/app'
      
      stage_dependencies do
        install 'flag'
      end
    end

    add '/assets' do
      to '/web/site.narod.ru/assets_with_strange_name'
      stage_dependencies.setup '*.less'
    end
  end
end
```

### artifact
Директива позволяет определить один или несколько [артефактов](definitions.html#артефакт).

* Для добавления артефакта можно использовать поддирективу export или директиву import.
  * Принимает параметр \<cwd\> (по умолчанию используется путь \<to\>).
  * Параметры \<include_paths\>, \<exclude_paths\>, \<owner\>, \<group\>, \<to\> определяются в контексте.
  * Параметр \<to\> по умолчанию соответствует \<cwd\>.
  * Если собирается не scratch: в контексте обязательно использование хотя бы одной из директив **before** или **after**, где:
    * директива определяет порядок применения артефакта (до или после);
    * значение определяет стадию (install или setup).

#### Примеры

##### Собрать с артефактами

###### export
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

###### import
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

##### Собрать scratch образ
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