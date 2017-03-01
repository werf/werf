---
title: Стадии
sidebar: doc_sidebar
permalink: stages.html
folder: definition
---

### Стадия
Стадия (stage) — это сгруппированный набор инструкций для сборки docker образа.

* Стадии заданы статически.
  * Предопределены имена стадий.
  * Предопределен порядок следования стадий.
  * См. [конвеер стадий](#конвеер-стадий).
* В результате сборки стадии создается отдельный docker образ.
* Имя docker образа стадии формируется по шаблону: dimgstage-\<[имя проекта](#имя-проекта)\>:\<[сигнатура стадии](#сигнатура-стадии)\>.
* Собранный образ dimg представляет собой связанный список docker образов стадий.
* Стадия может быть пропущена, если для нее не указано инструкций.
  * Для такой стадии не будет существовать отдельный docker образ.
  * См. [сигнатура стадии](#сигнатура-стадии).

### Пользовательская стадия
Пользовательская стадия — это стадия, инструкции для сборки которой задаются пользователем dapp.

Инструкции задаются через [Dappfile](#Dappfile) или chef-рецепты — зависит от используемого сборщика: [shell сборщик](#shell-сборщик) или [chef сборщик](#chef-сборщик).

### Конвеер стадий
Конвеер стадий — это статически определенная последовательность стадий для сборки определенного типа образов. Существуют следующие конвееры стадий:

* конвеер стадий dimg;
* конвеер стадий артефакта;
* конвеер стадий scratch dimg.

### Конвеер стадий dimg
Конвеер стадий dimg — это стадии, использующиеся для сборки стандартных образов. Последовательность и имена стадий:

* from
* before install
* before-install artifact
* git-artifact archive
* git-artifact pre install patch
* install
* git-artifact post install patch
* after-install artifact
* before setup
* before-setup artifact
* git-artifact pre setup patch
* setup
* git-artifact post setup patch
* after-setup artifact
* git-artifact latest patch
* docker instructions

[Пользовательскими стадиями](#пользовательская-стадия) являются:
* before install
* install
* before setup
* setup

Конвеер стадий dimg используется как в [shell dimg](#shell-dimg), так и в [chef dimg](#chef-dimg).

### Конвеер стадий артефакта
Конвеер стадий артефакта — это стадии, использующиеся для сборки образов артефактов. Последовательность и имена стадий:

* from
* before install
* before-install artifact
* git-artifact archive
* git-artifact pre install patch
* install
* git-artifact post install patch
* after-install artifact
* before setup
* before-setup artifact
* git-artifact pre setup patch 
* setup
* after-setup artifact
* git-artifact artifact patch
* build artifact

[Пользовательскими стадиями](#пользовательская-стадия) являются:
* before install
* install
* before setup
* setup
* build artifact

### Конвеер стадий scratch dimg
Конвеер стадий scratch dimg — состоит из одной стадии [import artifacts](#import-artifacts).

Пользовательских стадий в данном конвеере нет.

### Сигнатура стадии
Сигнатура стадии (stage signature) — это контрольная сумма правил сборки, зависимостей стадии и сигнатуры предыдущей стадии, если она существует.

* Изменение сигнатуры стадии ведет к её пересборке, а также последующих стадий.
* При отсутствии правил и зависимостей, стадия игнорируется, используется сигнатура предыдущей стадии.

### Назначение стадий

| Имя                               | Краткое описание 					          | Зависимость от директив                            |
| --------------------------------- | ----------------------------------- | -------------------------------------------------- |
| from                              | Выбор базового образа  					    | docker.from 			   						                   |
| before install                    | Установка софта инфраструктуры      | shell.before install / chef.dimod, chef.recipe     |
| before install artifact           | Наложение артефактов 				        | artifact (с before: :install) 			   		         |
| git artifact archive              | Наложение git-артефактов            | git_artifact.local`` и git_artifact.remote 		     |
| git artifact pre install patch    | Наложение патчей git-артефактов 	  | git_artifact.local и git_artifact.remote           |
| install                           | Установка софта приложения          | shell.install / chef.dimod, chef.recipe            |
| git artifact post install patch   | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| after install artifact            | Наложение артефактов                | artifact (с after: :install)               		     |
| before setup                      | Настройка софта инфраструктуры      | shell.before_setup / chef.dimod, chef.recipe       |
| before setup artifact             | Наложение артефактов                | artifact (с before: :setup)                		     |
| git artifact pre setup patch      | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| setup                             | Развёртывание приложения            | shell.setup / chef.dimod, chef.recipe              |
| chef cookbooks                    | Установка cookbook\`ов              | -             		       						               |
| git artifact post setup patch     | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| after setup artifact              | Наложение артефактов                | artifact (с after: :setup)            	   		     |
| git artifact latest patch         | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| docker instructions               | Применение докерфайловых инструкций | docker.cmd, docker.env, docker.entrypoint, docker.expose, docker.label, docker.onbuild, docker.user, docker.volume, docker.workdir |
| git artifact artifact patch       | Наложение патчей git-артефактов     | git_artifact.local и git_artifact.remote           |
| build artifact                    | Сборка артефакта                    | shell.build_artifact / chef.dimod, chef.recipe     |
| import artifacts                  | Установка артефактов при сборке [scratch dimg](base.html#scratch-dimg) | |

### Особенности
* Существуют стадии, в формировании [cигнатур](definitions.html#сигнатура-стадии) которых используется сигнатура последующей стадии, вдобавок к зависимостям самой стадии. Такие стадии всегда будут пересобираться вместе с зависимой стадией.  
  * git artifact pre install patch зависит от install.
  * git artifact post install patch зависит от before setup.
  * git artifact pre setup patch зависит от setup.
  * git artifact artifact patch зависит от build artifact.
* Сигнатура стадии git artifact post setup patch зависит от размера патчей git-артефактов и будет пересобрана, если их сумма превысит лимит (10 MB).

#### from

Данная стадия производит скачивание указанного базового образа (фактически docker pull) и фиксирует его в кэше dapp.

* Стадия используется только при указании базового образа директивой docker.from с аргументом в формате <image:tag>.
* Стадия не будет использоваться, если docker.from не указан — будет собран [scratch dimg](base.html#scratch-dimg)

#### import artifacts

Данная стадия включается только при сборке [scratch dimg](base.html#scratch-dimg) и является единственной стадией при сборке scratch dimg.

Сборка scratch dimg предполагает создание образа только путем импорта в итоговый образ файловых ресурсов описанных пользователем артефактов.

Порядок сборки: собирается каждый из описанных артефактов, отрабатывает стадия import artifacts, добавляя все описанные артефакты в итоговый образ (фактически с помощью docker import). При этом сборка каждого из артефактов идет изолированно и проходит через все стандартные стадии сборки артефактов.

#### chef cookbooks
Стадия устанавливает cookbook`и, указанные в Berksfile проекта, в собираемый образ.

* Во время установки cookbook`ов в данной стадии, устанавливается переменная окружения DAPP_CHEF_COOKBOOKS_VENDORING=1.
* Для cookbook`ов, нужных в собираемом образе, но не нужных для сборки самого образа с помощью chef-сборщика можно использовать проверку в Berksfile, например:

```ruby
source 'https://supermarket.chef.io'

cookbook 'test', path: '.'
cookbook 'dimod-test', path: '../dimod-test'
cookbook 'dimod-test2', path: '../dimod-test2'
cookbook 'dimod-testartifact', path: '../dimod-testartifact'

cookbook 'apt'

if ENV['DAPP_CHEF_COOKBOOKS_VENDORING']
  cookbook 'dimod-nginx'
  cookbook 'dimod-init'
end
```
