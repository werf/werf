---
title: Директивы артефактов и git-артефактов
sidebar: doc_sidebar
permalink: artifact_directives.html
folder: directive
---

### Особенности

* Пути добавления не должны пересекаться между артефактами.
* Изменение любого параметра артефакта ведёт к смене сигнатур, пересборке связанных [стадий](definitions.html#стадии) приложения.
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

### git \<type_or_url\>
Директива позволяет определить один или несколько git-артефактов.

* Поддерживается два типа git-артефактов, local и remote, которые задаются параметром \<type_or_url\>.
  * Строка 'local'.
  * Url адрес удалённого git-репозитория.
* Для добавления git-артефакта необходимо использовать поддирективу add.
  * Принимает параметр \<cwd\>.
  * Параметры \<include_paths\>, \<exclude_paths\>, \<owner\>, \<group\>, \<to\> определяются в контексте.
  * Параметры \<branch\>, \<commit\> могут быть определены в контексте remote git-артефакта.
* В контексте директивы можно указать базовые параметры git-артефактов, которые могут быть переопределены в контексте каждого из них.
  * \<owner\>.
  * \<group\>.
  * \<branch\>.
  * \<commit\>.

#### Примеры

##### Собрать с несколькими git-артефактами
```ruby
dimg do
  docker.from 'image:tag'

  git 'local' do
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

### artifact
Директива позволяет определить один или несколько [артефактов](definitions.html#артефакт).

* Для добавления артефакта необходимо использовать поддирективу export.
  * Принимает параметр \<cwd\>.
  * Параметры \<include_paths\>, \<exclude_paths\>, \<owner\>, \<group\>, \<to\> определяются в контексте.
  * Если собирается не scratch: в контексте обязательно использование хотя бы одной из директив **before** или **after**, где:
    * директива определяет порядок применения артефакта (до или после);
    * значение определяет стадию (install или setup).

#### Примеры

##### Собрать с артефактами
```ruby
dimg do
  docker.from 'image:tag'
  
  artifact do
    shell.build_artifact.run 'command1', 'command2'

    export '/' do
      exclude_paths 'assets'
      
      before 'install'
      to '/app'
    end
  end
  
  artifact do
    shell.build_artifact.run 'command3', 'command4'

    export '/' do
      include_paths 'assets'
    
      after 'setup'
      to '/app'
    end
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

### artifact\_depends\_on \<glob\>\[,\<glob\>, \<glob\>, ...]
Список файлов зависимостей для стадии [build_artifact](shell_directives.html#shell-build_artifact-<cmd>-<cmd>-cache_version-<cache_version>) [артефакта](definitions.html#артефакт).

* При изменении содержимого указанных файлов, произойдет пересборка стадии build_artifact.
* Учитывается лишь содержимое файлов и порядок в котором они указаны (имена файлов не учитываются).
* Поддерживаются glob-паттерны.
* Директории игнорируются.