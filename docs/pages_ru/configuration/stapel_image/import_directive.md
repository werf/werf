---
title: Импорт из артефактов и образов
sidebar: documentation
permalink: documentation/configuration/stapel_image/import_directive.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vSHlip8uqKZ7Wh00abw6kuh0_3raMr-g1LcLjgRDgztHVIHbY2V-_qp7zZ0GPeN46LKoqb-yMhfaG-l/pub?w=2031&amp;h=144" data-featherlight="image">
  <img src="https://docs.google.com/drawings/d/e/2PACX-1vSHlip8uqKZ7Wh00abw6kuh0_3raMr-g1LcLjgRDgztHVIHbY2V-_qp7zZ0GPeN46LKoqb-yMhfaG-l/pub?w=1016&amp;h=72">
  </a>

  <div class="language-yaml highlighter-rouge"><div class="highlight"><pre class="highlight"><code><span class="na">import</span><span class="pi">:</span>
  <span class="pi">-</span> <span class="na">artifact</span><span class="pi">:</span> <span class="s">&lt;artifact name&gt;</span>
    <span class="na">image</span><span class="pi">:</span> <span class="s">&lt;image name&gt;</span>
    <span class="na">before</span><span class="pi">:</span> <span class="s">&lt;install || setup&gt;</span>
    <span class="na">after</span><span class="pi">:</span> <span class="s">&lt;install || setup&gt;</span>
    <span class="na">add</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="na">to</span><span class="pi">:</span> <span class="s">&lt;absolute path&gt;</span>
    <span class="na">owner</span><span class="pi">:</span> <span class="s">&lt;owner&gt;</span>
    <span class="na">group</span><span class="pi">:</span> <span class="s">&lt;group&gt;</span>
    <span class="na">includePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
    <span class="na">excludePaths</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;relative path or glob&gt;</span>
  </code></pre></div></div>
---

Из-за используемых инструментов сборки, либо просто из-за исходных файлов, размер конечного образа может увеличиваться в несколько раз. Зачастую эти файлы не нужны в конечном образе. Для решения таких проблем, сообщество Docker предлагает выполнять установки инструментов, сборку и удаление ненужных файлов за один шаг.

Условный пример:
```
RUN “download-source && cmd && cmd2 && remove-source”
```

> Аналогичный пример может быть реализован и в Werf. Для этого достаточно описать инструкции в одной _пользовательской стадии_. Пример при использовании _shell-сборщика_ для стадии _install_ (аналогичен и для _ansible-сборщика_):
```yaml
shell:
  install:
  - "download-source"
  - "cmd"
  - "cmd2"
  - "remove-source"
```

Однако, при использовании такого метода кэширование работать не будет, и установка инструментов сборки будет выполняться каждый раз.

Другой способ--- использование multi-stage сборки, которая поддерживается начиная с версии 17.05 Docker.

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

Werf предлагает такой-же подход, но с использованием [_образов_]({{ site.baseurl }}/documentation/configuration/introduction.html#image-config-section) и  [_артефактов_]({{ site.baseurl }}/documentation/configuration/introduction.html#artifact-config-section).

> Почему Werf не использует multi-stage сборку?
* Исторически, возможность _импорта_ появилась значительно раньше чем в Docker появилась multi-stage сборка.
* Werf дает больше гибкости при работе со вспомогательными образами

Импорт _ресурсов_ из _образов_ и _артефактов_ должен быть описан в директиве `import` в конфигурации [_образа_]({{ site.baseurl }}/documentation/configuration/introduction.html#image-config-section) или [_артефакта_]({{ site.baseurl }}/documentation/configuration/introduction.html#artifact-config-section)) куда импортируются файлы. `import` — массив записей, каждая из которых должна содержать следующие параметры:

- `image: <image name>` или `artifact: <artifact name>`: _исходный образ_, имя образа из которого вы хотите копировать файлы или папки.
- `add: <absolute path>`: _исходный путь_, абсолютный путь к файлу или папке в _исходном образе_ для копирования.
- `to: <absolute path>`: _путь назначения_, абсолютный путь в _образе назначения_ (куда импортируются файлы или папки). В случае отсутствия считается равным значению указанному в параметре `add`.
- `before: <install || setup>` or `after: <install || setup>`: _стадия_ сборки _образа назначения_ для импорта. В настоящий момент возможен импорт только на стадиях _install_ или _setup_.

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
Это поведение аналогично используемому при добавлении кода из git-репозиториев, и вы можете подробнее почитать об этом в [соответствующем разделе]({{ site.baseurl }}/documentation/configuration/stapel_image/git_directive.html).

> Обратите внимание, что путь импортируемых ресурсов и путь указанный в _git mappings_ не должны пересекаться

Подробнее об использовании _артефактов_ можно узнать в [отдельной статье]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html).
