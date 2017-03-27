---
title: Артефакт (#TODO)
sidebar: doc_sidebar
permalink: artifact_for_advanced_build.html
folder: advanced_build
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

* Problem: ok, no packages in image. I don’t need build tools too. Let’s go to artifacts.
* Example 3: you may want to pack all static failes into scratch volume to mount it into nginx container.
* directives: artifact.

### Тезисы / проблемы / вопросы

* Ресурсы, которые требуются при сборке не нужны конечному пользователю и занимают лишнее место в образе.
  * компонент приложения из исходных файлов: инструменты сборки и исходные файлы далее не требуются.
  * скомпилированные статические файлы веб-приложения (css, js): компиляторы и исходные файлы далее не требуются.
* Подготовка ресурсов занимает значительное время, а пересборка происходит при изменении несвязанных данных.
  * компиляция должна происходить только при изменении исходных файлов.
* Часть ресурсов необходимо собирать в среде отличающейся от среды приложения.

### Предметная область

* Приложение артефакта используется для изолирования процесса сборки и инструментов сборки (среды, программного обеспечение, данных) ресурсов от образов, использующих эти ресурсы.
* Сборка приложения артефакта происходит по тем же правилам, что и сборка приложения, но с другим набором стадий.
* Артефакт — это набор ресурсов, который импортируется в приложение из приложения артефакта.
* Приложение может иметь произвольное количество артефактов.

### Раскрытие темы

Предположим, необходимо собрать приложение, в котором конечный пользователь ожидает:

* код проекта;
* собранную утилиту из проекта, код которой располагается в директории `system/service`;
* последнюю версию phantomjs.

```ruby
dimg do
  docker.from ‘ubuntu:16.04’

  # добавление кода проекта и зависимость сборки утилиты от изменений в её исходных файлах
  git.add do
    to(‘/app’)
    stage_dependencies.setup('system/service')
  end

  # добавление исходных файлов phantomjs и зависимость сборки от изменений в директории src
  git('https://github.com/ariya/phantomjs.git').add do
    to('/phantomjs')
    stage_dependencies.setup('src')
  end

  # установка пакетов для сборки phantomjs
  shell.install.run ‘apt-get install build-essential g++ flex \
                                     bison gperf ruby perl \
                                     libsqlite3-dev \
                                     libfontconfig1-dev \
                                     libicu-dev libfreetype6 libssl-dev \
                                     libpng-dev libjpeg-dev python \
                                     libx11-dev libxext-dev’

  # компиляция
  shell.setup.run 'python phantomjs/build.py'

  # сборка утилиты приложения
  shell.setup.run 'python app/system/service/build.py'
end
```

В данном случае, имеем следующие проблемы:

* помимо скомпилированного бинарного файла phantomjs конечный пользователь получает исходные файлы и инструменты сборки;
* изменения в файлах утилиты проекта или в файлах репозитория phantomjs приводят к пересборке как phantomjs, так и утилиты.

Решением может быть использование артефакта:

```ruby
dimg do
 docker.from ‘ubuntu:16.04’

 artifact do
   # добавление исходных файлов phantomjs и зависимость сборки от изменений в директории src
   git('https://github.com/ariya/phantomjs.git').add do
     to('/phantomjs')
     stage_dependencies.build_artifact('src')
   end

   # установка пакетов для сборки phantomjs
   shell.install.run ‘apt-get install build-essential g++ flex \
                                      bison gperf ruby perl \
                                      libsqlite3-dev \
                                      libfontconfig1-dev \
                                      libicu-dev libfreetype6 libssl-dev \
                                      libpng-dev libjpeg-dev python \
                                      libx11-dev libxext-dev’

   # компиляция
   shell.build_artifact.run 'python phantomjs/build.py'

   # добавление директории с бинарным файлов в образ приложения после стадии setup
   export('/phantomjs/bin') do
     to('/phantomjs/bin')
     after('setup')
   end
 end

 # добавление кода проекта и зависимость сборки утилиты от изменений в её исходных файлах
 git.add do
   to(‘/app’)
   stage_dependencies.setup('system/service')
 end

 # сборка утилиты приложения
 shell.setup.run 'python app/system/service/build.py'
end
```

Таким образом:

* артефакт изолирует процесс и инструменты сборки phantomjs;
* сборка артефакта происходит при изменениях в репозитории и не зависит от сборки утилиты.
