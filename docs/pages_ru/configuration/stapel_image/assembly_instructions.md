---
title: Запуск инструкций сборки
sidebar: documentation
permalink: documentation/configuration/stapel_image/assembly_instructions.html
summary: |
  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vQcjW39mf0TUxI7yqNzKPq4_9ffzg2IsMxQxu1Uk1-M0V_Wq5HxZCQJ6x-iD-33u2LN25F1nbk_1Yx5/pub?w=2031&amp;h=144" data-featherlight="image">
      <img src="https://docs.google.com/drawings/d/e/2PACX-1vQcjW39mf0TUxI7yqNzKPq4_9ffzg2IsMxQxu1Uk1-M0V_Wq5HxZCQJ6x-iD-33u2LN25F1nbk_1Yx5/pub?w=1016&amp;h=72">
  </a>

  <div class="tabs">
    <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'shell')">Shell</a>
    <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'ansible')">Ansible</a>
  </div>

  <div id="shell" class="tabs__content active">
    <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="na">shell</span><span class="pi">:</span>
    <span class="na">beforeInstall</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">install</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">beforeSetup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">setup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;bash command&gt;</span>
    <span class="na">cacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeInstallCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">installCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeSetupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">setupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span></code></pre>
    </div>
  </div>

  <div id="ansible" class="tabs__content">
    <div class="language-yaml highlighter-rouge"><pre class="highlight"><code><span class="na">ansible</span><span class="pi">:</span>
    <span class="na">beforeInstall</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">install</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">beforeSetup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">setup</span><span class="pi">:</span>
    <span class="pi">-</span> <span class="s">&lt;task&gt;</span>
    <span class="na">cacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeInstallCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">installCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">beforeSetupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span>
    <span class="na">setupCacheVersion</span><span class="pi">:</span> <span class="s">&lt;arbitrary string&gt;</span></code></pre>
      </div>
  </div>

  <br/>
  <b>Running assembly instructions with git</b>

  <a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=1956&amp;h=648" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=622&amp;h=206">
  </a>
---

## Пользовательские стадии

***Пользовательские стадии*** — это [_стадии_]({{ site.baseurl }}/documentation/reference/stages_and_images.html) со сборочными инструкциями из [конфигурации]({{ site.baseurl}}/documentation/configuration/introduction.html#что-такое-конфигурация-werf). Другими словами, — это стадии, которые конфигурирует пользователь (существуют также служебные стадии, которые пользователь конфигурировать не может). В настоящее время существует два вида сборочных инструкций: _shell_ и _ansible_.

В werf существуют четыре _пользовательские стадии_, которые выполняются последовательно, в следующем порядке: _beforeInstall_, _install_, _beforeSetup_ и _setup_. В результате выполнения инструкций какой-либо стадии создается один Docker-слой. Т.е. по одному слою на каждую стадию, в независимости от количества инструкций в ней.

## Причины использовать стадии

### Своя концепция структуры сборки (opinionated software)

Шаблон и механизм работы _пользовательских стадий_ основан на анализе сборки реальных приложений. В результате анализа мы пришли к выводу, что для того, чтобы качественно улучшить сборку большинства приложений, достаточно разбить инструкции сборки на 4 группы (эти группы и есть _пользовательские стадии_), подчиняющиеся определенным правилам. Такая группировка уменьшает количество слоев и ускоряет сборку.

### Четкая структура сборочного процесса

Наличие _пользовательских стадий_ определяет структуру процесса сборки и, таким образом, устанавливает некоторые рамки для разработчика. Несмотря на дополнительное ограничение по сравнению с неструктурированными инструкциями Docker-файла, это наоборот дает выигрыш в скорости, т.к. разработчик знает, какие инструкции на каком этапе должны быть.

### Запуск инструкций сборки при изменениях в git-репозитории

werf может использовать как локальные так и удаленные git-репозитории при сборке. Любой _пользовательской стадии_ можно определить зависимость от изменения конкретных файлов или папок в одном или нескольких git-репозиториях. Это дает возможность принудительно пересобирать _пользовательскую стадию_ если в локальном или удаленном репозитории (или репозиториях) изменяются какие-либо файлы.

### Больше инструментов сборки: shell, ansible, ...

_Shell_ — знакомый и хорошо известный инструмент сборки. 
_Ansible_ — более молодой инструмент, требующий чуть больше времени на изучение.

Если вам нужен быстрый результат с минимальными затратами времени и как можно быстрее, то использования _shell_ может быть вполне достаточно — все работает аналогично директиве `RUN` в Dockerfile.

В случае с _Ansible_ применяется декларативный подход и подразумевается идемпотентность операций. Такой подход дает более предсказуемый результат, особенно в случае проектов большого жизненного цикла.

Архитектура werf позволяет добавлять в будущем поддержку и других инструментов сборки.

## Использование пользовательских стадий

werf позволяет определять до четырех _пользовательских стадий_ с инструкциями сборки. На содержание самих инструкций сборки werf не накладывает каких-либо ограничений, т.е. вы можете указывать все те же инструкции, которые указывали в Dockerfile в директиве `RUN`. Однако важно не просто перенести инструкции из Dockerfile, а правильно разбить их на _пользовательские стадии_. Мы предлагаем такое разбиение исходя из опыта работы с реальными приложениями, и вся суть тут в том, что большинство сборок приложений проходят следующие этапы:
- установка системных пакетов
- установка системных зависимостей
- установка зависимостей приложения
- настройка системных пакетов
- настройка приложения

Какая может быть наилучшая стратегия выполнения этих этапов? Может показаться очевидным, что лучше всего выполнять эти этапы последовательно, кешируя промежуточные результаты. 
Либо, как вариант, не смешивать инструкции этапов, из-за разных файловых зависимостей. 
Шаблон _пользовательских стадий_ предлагает следующую стратегию:
- использовать стадию _beforeInstall_ для инсталляции системных пакетов;
- использовать стадию _install_ для инсталляции системных зависимостей и зависимостей приложения;
- использовать стадию _beforeSetup_ для настройки системных параметров и установки приложения;
- использовать стадию _setup_ для настройки приложения.

### beforeInstall

Данная стадия предназначена для выполнения инструкций перед установкой приложения. 
Сюда следует относить установку системных приложений которые редко изменяются, но установка которых занимает много времени. 
Примером таких приложений могут быть языковые пакеты, инструменты сборки, такие как composer, java, gradle и т. д. 
Также сюда правильно относить инструкции настройки системы, которые меняются редко. 
Например, языковые настройки, настройки часового пояса, добавление пользователей и групп.

Поскольку эти компоненты меняются редко, они будут кэшироваться в рамках стадии _beforeInstall_ на длительный период.

### install

Данная стадия предназначена для установки приложения и его зависимостей, а также выполнения базовых настроек.

На данной стадии появляется доступ к исходному коду (директива git), и возможна установка зависимостей на основе manifest-файлов с использованием таких инструментов как composer, gradle, npm и т.д. Поскольку сборка стадии зависит от manifest-файла, для достижения наилучшего результата важно связать изменение в manifest-файлах репозитория с данной стадией. Например, если в проекте используется composer, то установление зависимости от файла composer.lock позволит пересобирать стадию _beforeInstall_, в случае изменения файла composer.lock.

### beforeSetup

Данная стадия предназначена для подготовки приложения перед настройкой.

На данной стадии рекомендуется выполнять разного рода компиляцию и обработку. 
Например, компиляция jar-файлов, бинарных файлов, файлов библиотек, создание ассетов web-приложений, минификация, шифрование и т.п. 
Перечисленные операции, как правило, зависят от изменений в исходном коде, и на данной стадии также важно определить достаточные зависимости изменений в репозитории. Логично, что зависимость данной стадии от изменений в репозитории будет покрывать уже большую область файлов в репозитории, чем на предыдущей стадии, и, соответственно, ее пересборка будет выполняться чаще.

При правильно определенных зависимостях, изменение в коде приложения должно вызывать пересборку стадии _beforeSetup_, в случае если не изменялся manifest-файл. А в случае если manifest-файл изменился, должна уже вызываться пересборка стадии _install_ и следующих за ней стадий.

### setup

Данная стадия предназначена для настройки приложения.

Обычно на данной стадии выполняется копирование файлов конфигурации (например, в папку `/etc`), создание файлов текущей версии приложения и т.д. 
Такого рода операции не должны быть затратными по времени, т.к. они скорее всего будут выполняться при каждом новом коммите в репозитории.

### Пользовательская стратегия

Несмотря на изложенную четкую стратегию шаблона _пользовательских стадий_ и назначения каждой стадии, по сути нет никаких ограничений. Предложенные назначения каждой стадии являются лишь рекомендацией, которые основаны на нашем анализе работы реальных приложений. Вы можете использовать только одну пользовательскую стадию, либо определить свою стратегию группировки инструкций, чтобы получить преимущества кэширования и зависимостей от изменений в git-репозиториях с учетом особенностей сборки вашего приложения.

## Синтаксис

Пользовательские стадии и инструкции сборки определяются внутри двух взаимоисключающих директив вида сборочных инструкций — `shell` и `ansible`. Вы можете собирать образ используя внутри всех стадий либо сборочные инструкции ***shell***, либо сборочные инструкции ***ansible***.

Внутри директив вида сборочных инструкций можно указывать четыре директивы сборочных инструкций каждой _пользовательской стадии_, соответственно:
- `beforeInstall`
- `install`
- `beforeSetup`
- `setup`

Внутри директив вида сборочных инструкций также можно указывать директивы версий кэша (***cacheVersion***), которые по сути являются частью сигнатуры каждой _пользовательской стадии_. Более подробно об этом читай в [соответствующем разделе](#зависимость-от-значения-cacheversion) section.

## Shell

Синтаксис описания _пользовательских стадий_ при использовании сборочных инструкций _shell_:

```yaml
shell:
  beforeInstall:
  - <bash_command 1>
  - <bash_command 2>
  ...
  - <bash_command N>
  install:
  - bash command
  ...
  beforeSetup:
  - bash command
  ...
  setup:
  - bash command
  ...
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```

Сборочные инструкции _shell_ — это массив bash-команд для соответствующей _пользовательской стадии_. Все команды одной стадии выполняются как одна инструкция `RUN` в Dockerfile, т.е. в результате создается один слой на каждую _пользовательскую стадию_.

werf при сборке использует собственный исполняемый файл bash и вам не нужно отдельно добавлять его в образ (или [базовый образ]({{ site.baseurl }}/documentation/configuration/stapel_image/base_image.html)) при сборке. Все команды одной стадии объединяются с помощью выражения `&&` bash и кодируются алгоритмом base64 перед передачей в _сборочный контейнер_. _Сборочный контейнер_ пользовательской стадии декодирует команды и запускает их.

Пример описания стадии _beforeInstall_ содержащей команды `apt-get update` и `apt-get install`:

```yaml
beforeInstall:
- apt-get update
- apt-get install -y build-essential g++ libcurl4
```

werf выполнит команды стадии следующим образом::
- на хост-машине сгенерируется временный скрипт:

    ```shell
    #!/.werf/stapel/embedded/bin/bash -e

    apt-get update
    apt-get install -y build-essential g++ libcurl4
    ```

- скрипт смонтируется в _сборочный контейнер_ как `/.werf/shell/script.sh`
- скрипт выполнится.

> Исполняемый файл `bash` находится внутри Docker-тома _stapel_. Подробнее про эту концепцию можно узнать в этой [статье](https://habr.com/company/flant/blog/352432/) (упоминаемый в статье `dappdeps` был переименован в `stapel`, но принцип сохранился)

## Ansible

Синтаксис описания _пользовательских стадий_ при использовании сборочных инструкций _ansible_:

```yaml
ansible:
  beforeInstall:
  - <ansible task 1>
  - <ansible task 2>
  ...
  - <ansible task N>
  install:
  - ansible task
  ...
  beforeSetup:
  - ansible task
  ...
  setup:
  - ansible task
  ...
  cacheVersion: <version>
  beforeInstallCacheVersion: <version>
  installCacheVersion: <version>
  beforeSetupCacheVersion: <version>
  setupCacheVersion: <version>
```

### Ansible config and stage playbook

Сборочные инструкции _ansible_ —  это массив Ansible-заданий для соответствующей _пользовательской стадии_. Для запуска этих заданий с помощью `ansible-playbook` werf монтирует следующую структуру папок в _сборочный контейнер_:

```shell
/.werf/ansible-workdir
├── ansible.cfg
├── hosts
└── playbook.yml
```

`ansible.cfg` содержит настройки для Ansible:
- использование локального транспорта (transport = local)
- подключение callback плагина werf для удобного логирования (stdout_callback = werf)
- включение режима цвета (force_color = 1)
- установка использования `sudo` для повышения привилегий (чтобы не было необходимости использовать `become` в Ansible-заданиях)

`hosts` — inventory-файл, содержит только localhost и некоторые `ansible_*` параметры.

`playbook.yml` — playbook, содержащий все задания соответствующей _пользовательской стадии_. Пример `werf.yaml` с описанием стадии _install_:

```yaml
ansible:
  install:
  - debug: msg='Start install'
  - file: path=/etc mode=0777
  - copy:
      src: /bin/sh
      dest: /bin/sh.orig
  - apk:
      name: curl
      update_cache: yes
  ...
```

В приведенном примере, werf сгенерирует следующий `playbook.yml` для стадии _install_:
```yaml
- hosts: all
  gather_facts: 'no'
  tasks:
  - debug: msg='Start install'  \
  - file: path=/etc mode=0777   |
  - copy:                        > эти строки будут скопированы из werf.yaml
      src: /bin/sh              |
      dest: /bin/sh.orig        |
  - apk:                        |
      name: curl                |
      update_cache: yes         /
  ...
```

werf выполняет playbook _пользовательской стадии_ в сборочном контейнере стадии с помощью команды `playbook-ansible`:

```shell
$ export ANSIBLE_CONFIG="/.werf/ansible-workdir/ansible.cfg"
$ ansible-playbook /.werf/ansible-workdir/playbook.yml
```

> Исполняемые файлы и библиотеки `ansible` и `python` находятся внутри Docker-тома _stapel_. Подробнее про эту концепцию можно узнать в этой [статье](https://habr.com/company/flant/blog/352432/) (упоминаемый в статье `dappdeps` был переименован в `stapel`, но принцип сохранился)

### Поддерживаемые модули

Одной из концепций, которую использует werf, является идемпотентность сборки. Это значит что если "ничего не изменилось", то werf при повторном и последующих запусках сборки должен создавать бинарно идентичные образы. В werf эта задача решается с помощью подсчета _сигнатур стадий_.

Многие модули Ansible не являются идемпотентными, т.е. они могут давать разный результат запусков при неизменных входных параметрах. Это, конечно, не дает возможность корректно высчитывать _сигнатуру_ стадии, чтобы определять реальную необходимость её пересборки из-за изменений. Это привело к тому, что список поддерживаемых модулей был ограничен.

На текущий момент, список поддерживаемых модулей Ansible следующий:

- [Commands modules](https://docs.ansible.com/ansible/2.5/modules/list_of_commands_modules.html): command, shell, raw, script.
- [Crypto modules](https://docs.ansible.com/ansible/2.5/modules/list_of_crypto_modules.html): openssl_certificate и другие.
- [Files modules](https://docs.ansible.com/ansible/2.5/modules/list_of_files_modules.html): acl, archive, copy, stat, tempfile и другие.
- [Net Tools Modules](https://docs.ansible.com/ansible/2.5/modules/list_of_net_tools_modules.html): get_url, slurp, uri.
- [Packaging/Language modules](https://docs.ansible.com/ansible/2.5/modules/list_of_packaging_modules.html#language): composer, gem, npm, pip и другие.
- [Packaging/OS modules](https://docs.ansible.com/ansible/2.5/modules/list_of_packaging_modules.html#os): apt, apk, yum и другие.
- [System modules](https://docs.ansible.com/ansible/2.5/modules/list_of_system_modules.html): user, group, getent, locale_gen, timezone, cron и другие.
- [Utilities modules](https://docs.ansible.com/ansible/2.5/modules/list_of_utilities_modules.html): assert, debug, set_fact, wait_for.

При указании в _конфигурации сборки_ модуля отсутствующего в приведенном списке, сборка прервется с ошибкой. Не стесняйтесь [сообщать](https://github.com/flant/werf/issues/new) нам, если вы считаете что какой-либо модуль должен быть включен в список поддерживаемых.

### Копирование файлов

Предпочтительный способ копирования файлов в образ — использование [_git mapping_]({{ site.baseurl }}/documentation/configuration/stapel_image/git_directive.html). 
werf не может определять изменения в копируемых файлах при использовании модуля `copy`. 
Единственный вариант копирования внешнего файла в образ на текущий момент — использовать метод `.Files.Get` Go-шаблона. 
Данный метод возвращает содержимое файла как строку, что дает возможность использовать содержимое как часть _пользовательской стадии_. 
Таким образом, при изменении содержимого файла изменится сигнатура соответствующей стадии, что приведет к пересборке всей стадии.

Пример копирования файла `nginx.conf` в образ:

{% raw %}
```yaml
ansible:
  install:
  - copy:
      content: |
{{ .Files.Get "/conf/etc/nginx.conf" | indent 8}}
      dest: /etc/nginx/nginx.conf
```
{% endraw %}

werf применит Go-шаблонизатор и в результате получится подобный `playbook.yml`:

```yaml
- hosts: all
  gather_facts: no
  tasks:
    install:
    - copy:
        content: |
          http {
            sendfile on;
            tcp_nopush on;
            tcp_nodelay on;
            ...
```

### Шаблоны Jinja

В Ansible реализована поддержка шаблонов [Jinja](https://docs.ansible.com/ansible/2.5/user_guide/playbooks_templating.html) в playbook'ах. Однако, у Go-шаблонов и Jinja-шаблонов одинаковый разделитель: {% raw %}`{{` и `}}`{% endraw %}. Чтобы использовать Jinja-шаблоны в конфигурации werf, их нужно экранировать. Для этого есть два варианта: экранировать только {% raw %}`{{`{% endraw %}, либо экранировать все выражение шаблона Jinja.

Например, у вас есть следующая задача Ansible:

{% raw %}
```yaml
- copy:
    src: {{item}}
    dest: /etc/nginx
    with_files:
    - /app/conf/etc/nginx.conf
    - /app/conf/etc/server.conf
```
{% endraw %}

{% raw %}
Тогда, выражение Jinja-шаблона `{{item}}` должно быть экранировано:
{% endraw %}

{% raw %}
```yaml
# экранируем только {{
src: {{"{{"}} item }}
```
либо
```yaml
# экранируем все выражение
src: {{`{{item}}`}}
```
{% endraw %}

### Проблемы с Ansible

- Live-вывод реализован только для модулей `raw` и `command`. Остальные модули отображают вывод каналов `stdout` и `stderr` после выполнения, что приводит к задержкам, скачкообразному выводу.
- Большой вывод в `stderr` может подвесить выполнение Ansible-задачи ([issue #784](https://github.com/flant/werf/issues/784)).
- Модуль `apt` подвисает на некоторых версиях Debian и Ubuntu. Проявляется также на наследуемых образах([issue #645](https://github.com/flant/werf/issues/645)).

## Зависимости пользовательских стадий

Одна из особенностей werf — возможность определять зависимости при которых происходит пересборка _стадии_.
Как указано в [справочнике]({{ site.baseurl }}/documentation/reference/stages_and_images.html), сборка _стадий_ выполняется последовательно, одна за другой, и для каждой _стадии_ высчитывается _сигнатура стадии_. У _сигнатур_ есть ряд зависимостей, при изменении которых _сигнатура стадии_ меняется, что служит для werf сигналом для пересборки стадии с измененной _сигнатурой_. Поскольку каждая следующая _стадия_ имеет зависимость, в том числе и от предыдущей _стадии_ согласно _конвейеру стадий_, при изменении сигнатуры какой-либо _стадии_, произойдет пересборка и _стадии_ с измененной сигнатурой и всех последующих _стадий_.

_Сигнатура пользовательских стадий_ и соответственно пересборка _пользовательских стадий_ зависит от изменений:
- в инструкциях сборки
- в директивах семейства _cacheVersion_
- в git-репозитории (или git-репозиториях)
- в файлах, импортируемых из [артефактов]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html)

Первые три описанных варианта зависимостей, рассматриваются подробно далее.

## Зависимость от изменений в инструкциях сборки

_Сигнатура пользовательской стадии_ зависит от итогового текста инструкций, т.е. после применения шаблонизатора. Любые изменения в тексте инструкций с учетом применения шаблонизатора Go или Jinja (в случае Ansible) в _пользовательской стадии_ приводят к пересборке _стадии_. Например, вы используете следующие _shell-инструкции_ :

```yaml
shell:
  beforeInstall:
  - echo "Commands on the Before Install stage"
  install:
  - echo "Commands on the Install stage"
  beforeSetup:
  - echo "Commands on the Before Setup stage"
  setup:
  - echo "Commands on the Setup stage"
```

При первой сборке этого образа будут выполнены инструкции всех четырех _пользовательских стадий_. 
В данной конфигурации нет _git mapping_, так что последующие сборки не приведут к повторному выполнению инструкций — _сигнатура пользовательских стадий_ не изменилась, сборочный кэш содержит актуальную информацию.

Изменим инструкцию сборки для стадии _install_:

```yaml
shell:
  beforeInstall:
  - echo "Commands on the Before Install stage"
  install:
  - echo "Commands on the Install stage"
  - echo "Installing ..."
  beforeSetup:
  - echo "Commands on the Before Setup stage"
  setup:
  - echo "Commands on the Setup stage"
```

Сигнатура стадии _install_ изменилась, и запуск werf для сборки приведет к выполнению всех инструкций стадии _install_ и инструкций последующих _стадий_, т.е. _beforeSetup_ и _setup_.

Сигнатура стадии может меняться также из-за использования переменных окружения и Go-шаблонов. 
Если не уделять этому достаточное внимание при написание конфигурации, можно столкнуться с неожиданными пересборки стадий. Например:

{% raw %}
```yaml
shell:
  beforeInstall:
  - echo "Commands on the Before Install stage for {{ env "CI_COMMIT_SHA” }}"
  install:
  - echo "Commands on the Install stage"
  ...
```
{% endraw %}

Первая сборка высчитает сигнатуру стадии _beforeInstall_ на основе команды наример такой (хэш коммита конечно будет другой):
```shell
echo "Commands on the Before Install stage for 0a8463e2ed7e7f1aa015f55a8e8730752206311b"
```

После очередного коммита, при сборке сигнатура стадии _beforeInstall_ уже будет другой (с других хэшем коммита), например:

```shell
echo "Commands on the Before Install stage for 36e907f8b6a639bd99b4ea812dae7a290e84df27"
```

Соответственно, используя переменную `CI_COMMIT_SHA` сигнатура стадии _beforeInstall_ будет меняться после каждого коммита, что будет приводить к пересборке.

## Зависимость от изменений в git-репозитории

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=1956&amp;h=648" data-featherlight="image">
    <img src="https://docs.google.com/drawings/d/e/2PACX-1vRv56S-dpoTSzLC_24ifLqJHQoHdmJ30l1HuAS4dgqBgUzZdNQyA1balT-FwK16pBbbXqlLE3JznYDk/pub?w=622&amp;h=206">
  </a>

Как описывалось в статье про [_git mapping_]({{ site.baseurl}}/documentation/configuration/stapel_image/git_directive.html), существуют специальные стадии _gitArchive_ и _gitLatestPatch_. 
Стадия _gitArchive_ выполняется после пользовательской стадии _beforeInstall_, а стадия _gitLatestPatch_ после пользовательской стадии _setup_, если в локальном git-репозитории есть изменения. 
Таким образом, чтобы выполнить сборку с последней версией исходного кода, можно либо пересобрать стадию _gitArchive_ с помощью [специального коммита]({{site.baseurl}}/documentation/configuration/stapel_image/git_directive.html#сброс-стадии-gitarchive), либо пересобрать стадию _beforeInstall_, изменив значение директивы _cacheVersion_ либо изменив сами инструкции стадии _beforeInstall_.

Пользовательские стадии _install_, _beforeSetup_ и _setup_ также могут зависеть от изменений в git-репозитории. В этом случае (если такая зависимость определена) git-патч применяется перед выполнением _пользовательской стадии_, чтобы сборочные инструкции выполнялись с последней версией кода приложения.

> Во время процесса сборки, исходный код обновляется **только в рамках одной стадии**, последующие стадии, зависящие последовательно друг от друга, будут использовать также обновленную версию файлов. 
> Первая сборка добавляет файлы из git-репозиотрия на стадии _gitArchive_. Все последующие сборки обновляют файлы  на стадии _gitCache_, _gitLatestPatch_ или на одной из следующих пользовательских стадий: _install_, _beforeSetup_, _setup_.
<br />
<br />
Пример этого этапа (фаза подсчета сигнатур, _calculating signatures_):
![git files actualized on specific stage]({{ site.baseurl }}/images/build/git_mapping_updated_on_stage.png)

Зависимость _пользовательской стадии_ от изменений в git-репозитории указывается с помощью параметра `git.stageDependencies`. Синтаксис:

```yaml
git:
- ...
  stageDependencies:
    install:
    - <mask 1>
    ...
    - <mask N>
    beforeSetup:
    - <mask>
    ...
    setup:
    - <mask>
```

У параметра `git.stageDependencies` возможно указывать 3 ключа: `install`, `beforeSetup` и `setup`. 
Значение каждого ключа — массив масок файлов, относящихся к соответствующей стадии. Соответствующая _пользовательская стадия_ пересобирается, если в git-репозитории происходят изменения подпадающие под указанную маску.

Для каждой _пользовательской стадии_ werf создает список подпадающих под маску файлов и вычисляет контрольную сумму каждого файла с учетом его аттрибутов и содержимого. Эти контрольные суммы являются частью _сигнатуры стадии_, поэтому любое изменение файлов в репозитории, подпадающее под маску, приводит к изменениям _сигнатуры стадии_. К этим изменениям относятся: изменение атрибутов файла, изменение содержимого файла, добавление или удаление подпадающего под маску файла и т.п..

При применении маски указанной в `git.stageDependencies` учитываются значения параметров `git.includePaths` и `git.excludePaths` (смотри подробнее про них в соответствующем [разделе]({{site.baseurl}}/documentation/configuration/stapel_image/git_directive.html#использование-фильтров). werf считает подпадающими под маску только файлы удовлетворяющие фильтру `includePaths` и подпадающие под маску `stageDependencies`. Аналогично, werf считает подпадающими под маску только файлы не удовлетворяющие фильтру `excludePaths` и не подпадающие под маску `stageDependencies`.

Правила описания маски в параметре `stageDependencies` аналогичны описанию параметров `includePaths` и `excludePaths`. Маска определяет шаблон для файлов и путей, и может содержать следующие шаблоны:

- `*` — Удовлетворяет любому файлу. Шаблон включает `.` и исключает `/`.
- `**` — Удовлетворяет директории со всем ее содержимым, рекурсивно.
- `?` — Удовлетворяет любому однму символу в имени файла (аналогично regexp-шаблону `/.{1}/`).
- `[set]` — Удовлетворяет любому символу из указанного набора символов. Аналогично использованию в regexp-шаблонах, включая указание диапазонов типа `[^a-z]`.
- `\` — Экранирует следующий символ.


Маска, которая начинается с шаблона `*` или `**`, должна быть взята в одинарные или двойные кавычки в `werf.yaml`:

```
# * в начале маски, используем двойные кавычки
- "*.rb"
# одинарные также работают
- '**/*'
# нет * в начале, можно не использовать кавычки
- src/**/*.js
```

Факт изменения файлов в git-репозитории werf определяет подсчитывая их контрольные суммы. Для _пользовательской стадии_ и для каждой маски применяется следующий алгоритм:

- werf создает список всех файлов согласно пути определенному в параметре `add`, и применяет фильтры `excludePaths` и `includePaths`;
- К каждому файлу с учетом его пути применяется маска, согласно правилам применения шаблонов;
- Если под маску подпадает папка, то все содержимое папки считается подпадающей под маску рекурсивно;
- У получившегося списка файлов werf подсчитывает контрольные суммы с учетом аттрибутов файлов и их содержимого.

Контрольные суммы подсчитываются вначале сборочного процесса, перед запуском какой-либо стадии..

Пример:

```yaml
image: app
git:
- add: /src
  to: /app
  stageDependencies:
    beforeSetup:
    - "*"
shell:
  install:
  - echo "install stage"
  beforeSetup:
  - echo "beforeSetup stage"
  setup:
  - echo "setup stage"
```

В приведенном файле конфигурации `werf.yaml` указан _git mapping_, согласно которому содержимое папки `/src` локального git-репозитория копируется в папку `/app` собираемого образа. 
Во время первой сборки, файлы кэшируются в стадии _gitArchive_ и выполняются сборочные инструкции стадий _install_, _beforeSetup_ и _setup_.

Сборка после следующего коммита, в котором будут только изменения файлов за пределами папки `/src` не приведет к выполнению инструкций каких-либо стадий. Если же коммит будет содержать изменение внутри папки `/src`, контрольные суммы файлов подпадающих под маску изменятся, werf применит git-патч и пересоберет все пользовательские стадии начиная со стадии _beforeSetup_, а именно — _beforeSetup_ и _setup_. Применение git-патча будет выполнено один раз, на стадии _beforeSetup_.

## Зависимость от значения CacheVersion

Существуют ситуации, когда необходимо принудительно пересобрать все или какую-то конкретную _пользовательскую стадию_. Этого можно достичь изменяя параметры `cacheVersion` или `<user stage name>CacheVersion`.

Сигнатура пользовательской стадии _install_ зависит от значения параметра `installCacheVersion`. Чтобы пересобрать пользовательскую стадию _install_ (и все последующие стадии), можно изменить значение параметра `installCacheVersion`.

> Обратите внимание, что параметры `cacheVersion` и `beforeInstallCacheVersion` имеют одинаковый эффект, — при изменении этих параметров возникает пересборка стадии  _beforeInstall_ и всех последующих стадий.

### Пример: Общий образ для нескольких приложений

Вы можете определить образ, содержащий общие системные пакеты в отдельном файле `werf.yaml`. Изменение параметра `cacheVersion` может быть использовано для пересборки этого образа, чтобы обновить версии системных пакетов.

```yaml
image: ~
from: ubuntu:latest
shell:
  beforeInstallCacheVersion: 2
  beforeInstall:
  - apt update
  - apt install ...
```

Этот образ может быть использован как базовый образ для нескольких приложений (например, если образ с hub.docker.com не удовлетворяет вашим требованиям).

### Пример использования внешних зависимостей

Параметры _CacheVersion_ можно использовать совместно с [шаблонами Go]({{ site.baseurl }}/documentation/configuration/introduction.html#go-templates), чтобы определить зависимость _пользовательской стадии_ от файлов, не находящихся в git-репозитории.

{% raw %}
```yaml
image: ~
from: ubuntu:latest
shell:
  installCacheVersion: {{.Files.Get "some-library-latest.tar.gz" | sha256sum}}
  install:
  - tar zxf some-library-latest.tar.gz
  - <build application>
```
{% endraw %}

Если использовать, например, скрипт загрузки файла `some-library-latest.tar.gz` и запускать werf для сборки уже после скачивания файла, то пересборка пользовательской стадии _install_ (и всех последующих) будет происходить в случае если скачан новый (измененный) файл.
