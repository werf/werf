---
title: Импорт из артефактов и образов
permalink: advanced/building_images_with_stapel/import_directive.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
directive_summary: import
---

Из-за используемых инструментов сборки, либо просто из-за исходных файлов, размер конечного образа может увеличиваться в несколько раз. Зачастую эти файлы не нужны в конечном образе. Для решения таких проблем, сообщество Docker предлагает выполнять установки инструментов, сборку и удаление ненужных файлов за один шаг.

Условный пример:
```
RUN “download-source && cmd && cmd2 && remove-source”
```

> Аналогичный пример может быть реализован и в werf. Для этого достаточно описать инструкции в одной _пользовательской стадии_. Пример при использовании _shell-сборщика_ для стадии _install_ (аналогичен и для _ansible-сборщика_):
```yaml
shell:
  install:
  - "download-source"
  - "cmd"
  - "cmd2"
  - "remove-source"
```

Однако при использовании такого метода кэширование работать не будет, и установка инструментов сборки будет выполняться каждый раз.

Другой способ — использование multi-stage сборки, которая поддерживается начиная с версии 17.05 Docker.

```
FROM node:latest AS storefront
WORKDIR /usr/src/atsea/app/react-app
COPY react-app .
RUN npm install
RUN npm run build

FROM maven:latest AS appserver
WORKDIR /usr/src/atsea
COPY pom.xml .
RUN mvn -B -f pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:resolve
COPY . .
RUN mvn -B -s /usr/share/maven/ref/settings-docker.xml package -DskipTests

FROM java:8-jdk-alpine
RUN adduser -Dh /home/gordon gordon
WORKDIR /static
COPY --from=storefront /usr/src/atsea/app/react-app/build/ .
WORKDIR /app
COPY --from=appserver /usr/src/atsea/target/AtSea-0.0.1-SNAPSHOT.jar .
ENTRYPOINT ["java", "-jar", "/app/AtSea-0.0.1-SNAPSHOT.jar"]
CMD ["--spring.profiles.active=postgres"]
```

Смысл такого подхода в следующем — описать несколько вспомогательных образов и выборочно копировать артефакты из одного образа в другой, оставляя все то, что не нужно в конечном образе.

werf предлагает такой-же подход, но с использованием [_образов_]({{ "reference/werf_yaml.html#секция-image" | true_relative_url }}) и _артефактов_.

> Почему werf не использует multi-stage сборку?
* Исторически, возможность _импорта_ появилась значительно раньше чем в Docker появилась multi-stage сборка.
* werf дает больше гибкости при работе со вспомогательными образами

Импорт _ресурсов_ из _образов_ и _артефактов_ должен быть описан в директиве `import` в конфигурации [_образа_]({{ "reference/werf_yaml.html#секция-image" | true_relative_url }}) или _артефакта_ куда импортируются файлы. `import` — массив записей, каждая из которых должна содержать следующие параметры:

- `image: <image name>` или `artifact: <artifact name>`: _исходный образ_, имя образа из которого вы хотите копировать файлы или папки.
- `stage: <stage name>`: _стадия исходного образа_, определённая стадия _исходного образа_ из которого вы хотите копировать файлы или папки.
- `add: <absolute path>`: _исходный путь_, абсолютный путь к файлу или папке в _исходном образе_ для копирования.
- `to: <absolute path>`: _путь назначения_, абсолютный путь в _образе назначения_ (куда импортируются файлы или папки). В случае отсутствия считается равным значению указанному в параметре `add`.
- `before: <install || setup>` or `after: <install || setup>`: _стадия образа назначения_ для импорта. В настоящий момент возможен импорт только до/после стадии _install_ или _setup_.

Пример:
```yaml
import:
- artifact: application-assets
  add: /app/public/assets
  to: /var/www/site/assets
  after: install
- image: frontend
  add: /app/assets
  after: setup
```

Так же как и при конфигурации _git mappings_ поддерживаются маски включения и исключения файлов и папок. 
Для указания маски включения файлов используется параметр `include_paths: []`, а для исключения `exclude_paths: []`. Маски указываются относительно пути источника (параметр `add`). 
Вы также можете указывать владельца и группу для импортируемых ресурсов с помощью параметров `owner: <owner>` и `group: <group>` соответственно. 
Это поведение аналогично используемому при добавлении кода из git-репозиториев, и вы можете подробнее почитать об этом в [соответствующем разделе]({{ "advanced/building_images_with_stapel/git_directive.html" | true_relative_url }}).

> Обратите внимание, что путь импортируемых ресурсов и путь указанный в _git mappings_ не должны пересекаться

Подробнее об использовании _артефактов_ можно узнать в [отдельной статье]({{ "advanced/building_images_with_stapel/artifacts.html" | true_relative_url }}).
