---
title: Using artifacts
sidebar: reference
permalink: reference/build/artifact_directive.html
---

```yaml
artifact: <artifact_name>
from: <image>
fromCacheVersion: <version>
fromDimg: <dimg_name>
fromDimgArtifact: <artifact_name>
git:
# local git
- as: <custom_name>
  add: <absolute_path>
  to: <absolute_path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative_path or glob>
  excludePaths:
  - <relative_path or glob>
  stageDependencies:
    install:
    - <relative_path or glob>
    beforeSetup:
    - <relative_path or glob>
    setup:
    - <relative_path or glob>
    buildArtifact:
    - <relative_path or glob>
# remote git
- url: <git_repo_url>
  branch: <branch_name>
  commit: <commit>
  tag: <tag>
  as: <custom_name>
  add: <absolute_path>
  to: <absolute_path>
  owner: <owner>
  group: <group>
  includePaths:
  - <relative_path or glob>
  excludePaths:
  - <relative_path or glob>
  stageDependencies:
    install:
    - <relative_path or glob>
    beforeSetup:
    - <relative_path or glob>
    setup:
    - <relative_path or glob>
    buildArtifact:
    - <relative_path or glob>
shell:
  beforeInstall:
  - <cmd>
  install:
  - <cmd>
  beforeSetup:
  - <cmd>
  setup:
  - <cmd>
  buildArtifact:
  - <cmd>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
  buildArtifactCacheVersion: <version>
ansible:
  beforeInstall:
  - <task>
  install:
  - <task>
  beforeSetup:
  - <task>
  setup:
  - <task>
  buildArtifact:
  - <task>
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
  buildArtifactCacheVersion: <version>
mount:
- from: build_dir
  to: <absolute_path>
- from: tmp_dir
  to: <absolute_path>
- fromPath: <absolute_path>
  to: <absolute_path>
asLayers: <false || true>
```

Размер конечного образа за счёт инструментов сборки и исходных файлов может увеличиваться в несколько раз, притом что пользователю они не требуются.

Для решения подобных проблем сообщество docker предлагает в одном шаге делать установку инструментов, сборку и удаление инструментов,

```
RUN “download-source && cmd && cmd2 && remove-source”
```

Но при таком использовании не получится использовать кэширование, а это время на постоянную установку инструментария.

dapp предлагает альтернативу в виде приложений артефактов, сборка которых осуществляется по тем же правилам, что и приложений, но с другим набором стадий.

Приложение артефакта используется для изолирования процесса сборки и инструментов сборки (среды, программного обеспечение, данных) ресурсов от образов, использующих эти ресурсы.

```ruby
dimg do
  docker.from 'ubuntu:16.04'

  # определение приложения артефакта
  artifact do
    # добавление исходных файлов и зависимости пересборки артефакта от любого изменения
    git.add do
      to('/app')
      stage_dependencies.build_artifact('*')
    end

    shell do
      # установка инструментов сборки
      install.run('apt-get install build-essentials libmysql-dev')
      # сборка
      build_artifact.run('make -C /app')
    end

    # определение артефакта, импортирование `/app/build` в `/usr/bin` приложения после стадии `setup`
    export('/app/build') do
      to('/usr/bin/app')
      after('setup')
    end
  end
end
```

В таком случае, сборка приложения будет осуществляться в образе артефакта, а в конечный образ попадёт только бинарный файл.

Разница между стадиями заключается в следующем:
* на стадии build\_artifact определяются шаги для сборки артефакта, зависимости от файлов которой можно описать в stage\_dependencies в директиве git;
* за наложение финального патча отвечает стадия g\_a\_artifact\_patch, которая будет собрана только в том случае, если потребуется пересобрать build\_artifact;
* не используется стадия docker\_instruction, так как приложение артефакта является служебным.

При отсутствии зависимостей у стадии build\_artifact артефакт закэшируется после первой сборки и не будет пересобираться.

Стоит отметить, что может быть произвольное количество как приложений артефактов, так и артефактов. Артефактом в данном случае называется директория, которая экспортируется в образ. Т.о. одно приложение артефакта может экспортировать несколько директорий в конечный образ приложения.

Таким образом, с использованием артефактов можно независимо собирать неограниченное количество компонентов, притом также решая следующие проблемы:
* Пересборка происходит при изменении несвязанных данных и подготовка ресурсов занимает значительное время, а приложение можно разделить на несвязанные компоненты.
* Ресурсы необходимо собирать в среде отличающейся от среды приложения.

Артефакт (artifact) — это набор правил для сборки образа с файловым ресурсом (образа артефакта), который затем используется в одном или нескольких образах приложений. Образ артефакта не используется и не остается после окончания процесса сборки образов приложений, он нужен для изолирования процесса сборки ресурсов, или инструментов сборки (среды, программного обеспечение, данных) от процесса сборки образов приложений, использующих эти ресурсы или инструменты.

Количество артефактов, описываемых в одном dappfile строго не ограничено

Основное различие при использовании разного синтаксиса dappfile в части описания артефактов - в случае с YAML синтаксисом доступен только импорт артефакта, в случае же с ruby синтаксисом, доступен как вариант описания импорта артефакта в образ приложения так и описание экспорта артефакта из образа артефакта.

## Правила использования

* пути добавления не должны пересекаться между артефактами
* изменение любого параметра артефакта ведёт к смене сигнатур, пересборке связанных стадий приложения
* приложение может содержать любое количество артефактов
* артефакты обязаны иметь имя
* имена образов приложений и образов артефактов должны различаться


## Синтаксис

Описание образа артефакта выполняется с помощью директивы `artifact: <name>`, где `<name>` - обязательное имя артефакта. Имя артефакта используется для указания артефакта по имени в описании образа приложения, при импорте файлов из артефакта в образ приложения. При использовании YAML синтаксиса доступен только импорт из артефакта (отсутствует директива `export`).

Указание базового образа при описании образа приложения или образа артефакта выполняется с помощью обязательной директивы `from: <DOCKER_IMAGE>`, где `<DOCKER_IMAGE>` - имя образа в формате `image:tag` (tag может отсутствовать, тогда используется latest). Как и при описании образов приложений, описание нескольких образов артефактов выполняется линейно - друг за другом, отделяются строкой состоящей из последовательности `---` (согласно YAML спецификации).

Использование артефакта в образе приложения описывается с помощью директивы `import:`, которая представляет собой массив следующих элементов:
* `artifact: ARTIFACT_NAME`, где `ARTIFACT_NAME` - имя артефакта, из которого необходимо скопировать файлы в образ приложения;
* `add: SOURCE_DIRECTORY_TO_IMPORT`, где `SOURCE_DIRECTORY_TO_IMPORT` - путь в образе артефакта к файлу или папке, которые должны быть скопированы в образ приложения;
* `to: DESTINATION_DIRECTORY`, где `DESTINATION_DIRECTORY` - путь в образе приложения, куда должны быть скопированы файлы из образа артефакта. Необязательный элемент, в случае отсутствия принимается равным значению `SOURCE_DIRECTORY_TO_IMPORT`;
* `before: STAGE | after: STAGE`, где `STAGE` - стадия `install` или `setup`, соответственно до (`before`) или после (`after`) которой необходимо импортировать файлы артефакта в образ.

Сборка образов артефактов отличается отсутствием стадии `latest_patch`, т.о. при первой сборке образа артефакта используются текущие состояния git-репозиториев и при последующих сборках образы артефактов не пересобираются (при условии, что отсутствуют зависимости пользовательских стадий от файлов, описанных с помощью директивы `stageDependencies`, о чем см. ниже).

Unlike with applications, building artifacts does not include the `latest_patch` stage.
When dapp first builds an artifact image, it uses the current state of the git repository.
On next builds it does not rebuild the image from scratch.

### Пример использования артефакта
```
import:
- artifact: application-assets
  add: /app/public/assets
  after: install
- artifact: application-assets
  add: /vendor
  to: /app/vendor
  before: setup
```

### Пример описания образа приложения с использованием артефакта

В следующем примере используется описание образа приложения и образа артефакта. В артефакте создается файл, который импортируется в образ `application`, и на стадии setup выводится его содержимое.

```
artifact: assets
from: alpine
shell:
  install:
  - mkdir /tmp/asset-files
  - echo 'Artifact file for import' >/tmp/asset-files/asset
---
dimg: application
from: ubuntu:16.04
import:
  - artifact: assets
    add: /tmp/asset-files
    to: /app
    after: install
shell:
  setup:
    - cat /app/asset
```
