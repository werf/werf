---
title: Гитерминизм
sidebar: documentation
permalink: documentation/advanced/configuration/giterminism.html
change_canonical: true
---

Начиная с версии v1.2 werf вводит так называемый _режим гитерминизма_ по умолчанию (примечание: название происходит от совмещения слов `git` и `determinism`, что можно понимать как "детерминированный гитом").

В данном режиме все конфигурационные файлы читаются из текущего коммита локального гит-репозитория. Запрещено иметь некоммитнутые изменения в этих файлах — werf упадёт с ошибкой при наличии таких файлов. Также запрещено использование некоторых директив конфигурации и шаблонизации (например `mount` и `env`). С помощью дополнительного [файла конфигурации `werf-giterminism.yaml`](#werf-giterminismyaml) можно явно снимать ограничения, вводимые данным режимом.

## `werf-giterminism.yaml`

Файл конфигурации `werf-giterminism.yaml` описывает какие ограничения режима гитерминизма должны быть сняты для текущей конфигурации (разрешить использование каких-либо некоммитнутых файлов, директив конфигурации, переменных окружения и т.д.).

**Важно**. Данный файл вступает в действие только если он коммитнут в репозиторий проекта.

**Важно**. Отключать режим детерминизма не рекомендовано, т.к. это повышает вероятность написания конфигурации, которая приведёт к невоспроизводимым сборкам и выкатам приложения. Важно минимизировать снятие ограничений детерминированного режима через `werf-giterminism.yaml` для построения конфигурации соответствующей подходу GitOps, надёжной и легко воспроизводимой.

{% raw %}
```yaml
giterminismConfigVersion: 1
config:  # giterminism configuration for werf.yaml
  allowUncommitted: true
  allowUncommittedTemplates:
    - /**/*/
    - .werf/template.tmpl
  goTemplateRendering:
    allowEnvVariables:                        # {{ env "VARIABLE_X" }}
      - /CI_*/
      - VARIABLE_X
    allowUncommittedFiles:                    # {{ .Files.Get|Glob|... "PATH1" }}
      - /**/*/
      - .werf/nginx.conf
  stapel:
    allowFromLatest: true
    git:
      allowBranch: true
    mount:
      allowBuildDir: true                     # from: build_dir
      allowFromPaths:                         # fromPath: PATH
        - PATH1
        - PATH2
  dockerfile:
    allowUncommitted:
      - /**/*/
      - myapp/Dockerfile
    allowUncommittedDockerignoreFiles:
      - /**/*/
      - myapp/.dockerignore
    allowContextAddFiles:
      - aaa
      - bbb
helm: # giterminism configuration for helm
  allowUncommittedFiles:
    - /templates/**/*/
    - values.yaml
    - Chart.yaml
```
{% endraw %}

### Функция `.Files.Get`

В режиме детерминизма функция `.Files.Get` доступная при чтении конфига `werf.yaml будет читать файлы только из текущего коммита гит-репозитория.

Для указания списка файлов, которые необходимо читать из текущей рабочей директории проекта, а не из текущего коммита гит-репозитория, используйте директиву [`config.goTemplateRendering.allowUncommittedFiles`](#werf-giterminismyaml) конфигурационного файла `werf-giterminism.yaml` (поддерживаются glob-ы).

### Функции go-шаблонов для доступа к переменным окружения

Функции [`{{ env }}` и `{{ expandenv }}`]({{ "documentation/advanced/configuration/supported_go_templates.html" | true_relative_url }}) доступны для использования только при включении директивы [`config.goTemplateRendering.allowEnvVariables`](#werf-giterminismyaml) конфигурационного файла `werf-giterminism.yaml` (поддерживаются glob-ы).

### Директива mount

[Директива `mount`]({{ "documentation/reference/werf_yaml.html" | true_relative_url }}) для сборщика образов stapel доступна для использования только при включении директив [`config.stapel.mount`](#werf-giterminismyaml) конфигурационного файла `werf-giterminism.yaml` (конкретная директива выбирается исходя из типа требуемого mount-а).

## Сборщик Dockerfile

Werf использует контекст для Dockerfile и сам `Dockerfile` и `.dockerignore` только из текущего коммита локального гит-репозитория.

Есть единственный способ явным образом добавить файлы вне гит-репозитория в контекст сборки Dockerfile: с помощью [директивы `contextAddFile`]({{ "documentation/reference/werf_yaml.html" | true_relative_url}}):

```
context: app
dockerfile: Dockerfile
contextAddFile:
 - myfile
 - dir/a.out
```

В данной конфигурации werf создаст контекст для Dockerfile, состоящий из:
 - директории `app` из текущего коммита локального гит-репозитория (эта директория указана директивой `context`);
 - файлов `myfile` и `dir/a.out` из директории `app` в текущей рабочей директории проекта (данные файлы могут быть не коммитнуты в гит и могут быть untracked).

Для использования директивы `contextAddFile` конфигурационного файла `werf.yaml` должна быть также явно включена директива [`config.dockerfile.allowContextAddFiles`](#werf-giterminismyaml) конфигурационного файла `werf-giterminism.yaml`:

```yaml
# werf-giterminism.yaml configuration file
giterminismConfigVersion: 1
config:
  dockerfile:
    allowContextAddFiles:
      - myfile
      - dir/a.out
```
