---
title: Стадии
sidebar: doc_sidebar
permalink: stages.html
folder: definition
---

| Имя                               | Краткое описание 					          | Зависимость от директив                            |
| --------------------------------- | ----------------------------------- | -------------------------------------------------- |
| from                              | Выбор окружения  					          | docker.from 			   						                   |
| before_install                    | Установка софта инфраструктуры      | shell.before_install / chef.module, chef.recipe    |
| before_install_artifact           | Наложение артефактов 				        | artifact (с before: :install) 			   		         |
| git_artifact_archive              | Наложение git-артефактов            | git_artifact.local`` и git_artifact.remote 		     |
| git_artifact_pre_install_patch    | Наложение патчей git-артефактов 	  | git_artifact.local и git_artifact.remote           |
| install                           | Установка софта приложения          | shell.install / chef.module, chef.recipe           |
| git_artifact_post_install_patch   | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| after_install_artifact            | Наложение артефактов                | artifact (с after: :install)               		     |
| before_setup                      | Настройка софта инфраструктуры      | shell.before_setup / chef.module, chef.recipe      |
| before_setup_artifact             | Наложение артефактов                | artifact (с before: :setup)                		     |
| git_artifact_pre_setup_patch      | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| setup                             | Развёртывание приложения            | shell.setup / chef.module, chef.recipe             |
| chef_cookbooks                    | Установка cookbook\`ов              | -             		       						               |
| git_artifact_post_setup_patch     | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| after_setup_artifact              | Наложение артефактов                | artifact (с after: :setup)            	   		     |
| git_artifact_latest_patch         | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| docker_instructions               | Применение докерфайловых инструкций | docker.cmd, docker.env, docker.entrypoint, docker.expose, docker.label, docker.onbuild, docker.user, docker.volume, docker.workdir |
| git_artifact_artifact_patch       | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| build_artifact                    | Сборка артефакта                    | shell.build_artifact / chef.module, chef.recipe    |

### Особенности
* Существуют стадии, в формирование [cигнатур](definitions.html#сигнатура-стадии) которых используется сигнатура последующей стадии, вдобавок к зависимостям самой стадии. Такие стадии всегда будут пересобираться вместе с зависимой стадией.  
  * git_artifact_pre_install_patch зависит от install.
  * git_artifact_post_install_patch зависит от before_setup.
  * git_artifact_pre_setup_patch зависит от setup.
  * git_artifact_artifact_patch зависит от build_artifact.
* Сигнатура стадии git_artifact_post_setup_patch зависит от размера патчей git-артефактов и будет пересобрана, если их сумма превысит лимит (10 MB).

### from

### before install

### before install artifact

### git artifact archive

### Группа install

#### git artifact pre install patch

#### install

#### git artifact post install patch

### after install artifact

### before setup

### before setup artifact

### Группа setup

#### git artifact pre setup patch

#### setup

#### chef cookbooks
Стадия устанавливает cookbook`и, указанные в Berksfile проекта, в собираемый образ.

* Во время установки cookbook`ов в данной стадии, устанавливается переменная окружения DAPP_CHEF_COOKBOOKS_VENDORING=1.
* Для cookbook`ов, нужных в собираемом образе, но не нужных для сборки самого образа с помощью chef-сборщика можно использовать проверку в Berksfile, например:

```ruby
source 'https://supermarket.chef.io'

cookbook 'test', path: '.'
cookbook 'mdapp-test', path: '../mdapp-test'
cookbook 'mdapp-test2', path: '../mdapp-test2'
cookbook 'mdapp-testartifact', path: '../mdapp-testartifact'

cookbook 'apt'

if ENV['DAPP_CHEF_COOKBOOKS_VENDORING']
  cookbook 'mdapp-nginx'
  cookbook 'mdapp-init'
end
```

внутрь контейнера
DAPP_CHEF_COOKBOOKS_VENDORING означает вендоринг cookbook`ов в стадии chef_cookbooks. Эта стадия устанавливает кукбуки внутрь контейнера, это используется только chefinit`ом.

Соответственно кукбуки, добавляемые под if DAPP_CHEF_COOKBOOKS_VENDORING, нужны только для chefinit.



#### git artifact post setup patch

### after setup artifact

### git artifact latest patch

### docker instructions

### git_artifact_artifact_patch

### build_artifact
