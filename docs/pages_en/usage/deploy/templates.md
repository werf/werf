---
title: Templates
permalink: usage/deploy/templates.html
---

## Templating

For templating, werf uses Helm-based templating implemented using the [Go text/template](https://pkg.go.dev/text/template) package.

Any Helm pattern can be used with werf. While werf offers a number of extra features on top of what Helm templating does, using these extra features is completely optional.

## Template files

Template files are located in the `templates` directory of the chart.  

The `templates/*.yaml` files are used to generate final, deployment-ready Kubernetes manifests. Each of these files can be used to generate multiple Kubernetes resource manifests. For this, insert a `---` separator between the manifests.

The `templates/_*.tpl` files only contain named templates for using in other files. Kubernetes manifests cannot be generated using the `*.tpl' files alone.

## Actions

Actions is the key element of the templating process. Actions can only return strings. The action must be wrapped in double curly braces:

{% raw %}

```
{{ print "hello" }}
```

{% endraw %}

Output:

```
hello
```

## Variables

Variables are used to store or refer to data of any type.

This is how you can declare a variable and assign a value to it:

{% raw %}

```
{{ $myvar := "hello" }}
```

{% endraw %}

This is how you can assign a new value to an existing variable:

{% raw %}

```
{{ $myvar = "helloworld" }}
```

{% endraw %}

Here's an example of how to use a variable:

{% raw %}

```
{{ $myvar }}
```

{% endraw %}

Output:

```
helloworld
```

Here's how to use predefined variables:

{% raw %}

```
{{ $.Values.werf.env }}
```

{% endraw %}

You can also substitute the data without first declaring a variable:

{% raw %}

```
labels:
  app: {{ "myapp" }}
```

{% endraw %}

Output:

```yaml
labels:
  app: myapp
```

You can also store function or pipeline results in variables:

{% raw %}

```
{{ $myvar := 1 | add 1 1 }}
{{ $myvar }} 
```

{% endraw %}

Output:

```
3
```

## Variable scope

Scope limits the visibility of variables. By default, the scope is limited to the template file.

The scope can change for some blocks and functions. For example, the `if` statement creates a different scope, and the variables declared in the `if` statement will not be accessible outside it:

{% raw %}

```
{{ if true }}
  {{ $myvar := "hello" }}
{{ end }}

{{ $myvar }}
```

{% endraw %}

Output:

```
Error: ... undefined variable "$myvar"
```

To get around this limitation, declare the variable outside the statement and assign a value to it inside the statement:

{% raw %}

```
{{ $myvar := "" }}
{{ if true }}
  {{ $myvar = "hello" }}
{{ end }}

{{ $myvar }}
```

{% endraw %}

Output:

```
hello
```

## Data types

Available data types:

| Data type                                                           | Example                                              |
| -------------------------------------------------------------------- | --------------------------------------------------- |
| Boolean                                                              | {% raw %}`{{ true }}`{% endraw %}                   |
| String                                                               | {% raw %}`{{ "hello" }}`{% endraw %}                |
| Integer                                                              | {% raw %}`{{ 1 }}`{% endraw %}                      |
| Floating-point number                                                | {% raw %}`{{ 1.1 }}`{% endraw %}                    |
| List with elements of any type (ordered)                             | {% raw %}`{{ list 1 2 3 }}`{% endraw %}             |
| Dictionary with string keys and values of any type (unordered)       | {% raw %}`{{ dict "key1" 1 "key2" 2 }}`{% endraw %} |
| Special objects                                                      | {% raw %}`{{ $.Files }}`{% endraw %}                |
| Null                                                                 | {% raw %}`{{ nil }}`{% endraw %}                    |

## Functions

werf has an extensive library of functions for use in templates. Most of them are Helm functions.

Functions can only be used in actions. Functions *can* have arguments and *can* return data of any type. For example, the add function below takes three numeric arguments and returns a number:

{% raw %}

```
{{ add 3 2 1 }}
```

{% endraw %}

Output:

```
6
```

Note that **the action result is always converted to a string** regardless of the data type returned by the function.

A function may have arguments of the following types:

- regular values: `1`

- calls of other functions: `add 1 1`

- pipes: `1 | add 1`

- combination of the above types: `1 | add (add 1 1)`

Put the argument in parentheses `()` if it is a call to another function or pipeline:

{% raw %}

```
{{ add 3 (add 1 1) (1 | add 1) }}
```

{% endraw %}

To ignore the result returned by the function, simply assign it to the `$_` variable:

{% raw %}

```
{{ $_ := set $myDict "mykey" "myvalue"}}
```

{% endraw %}

## Pipelines

Pipelines allow you to pass the result of the first function as the last argument to the second function, and the result of the second function as the last argument to the third function, and so on:

{% raw %}

```
{{ now | unixEpoch | quote }}
```

{% endraw %}

Here, the result of the `now` function (gets the current date) is passed as an argument to the `unixEpoch` function (converts the date to Unix time). The resulting value is then passed to the `quote` function (adds quotation marks).

Output:

```
"1671466310"
```

The use of pipelines is optional; you can rewrite them as follows:

{% raw %}

```
{{ quote (unixEpoch (now)) }}
```

{% endraw %}

... however, we recommend using the pipelines.

## Logic gates and comparisons

The following logic gates are available:

| Operation | Function                       | Example                                    |
| --------- | ------------------------------ | ------------------------------------------ |
| Not       | `not <arg>`                    | {% raw %}`{{ not false }}`{% endraw %}     |
| And       | `and <arg> <arg> [<arg>, ...]` | {% raw %}`{{ and true true }}`{% endraw %} |
| Or        | `or <arg> <arg> [<arg>, ...]`  | {% raw %}`{{ or false true }}`{% endraw %} |

The following comparison operators are available:

| comparison              | Function                       | Example                                          |
| ----------------------- | ------------------------------ | ------------------------------------------------ |
| Equal                   | `eq <arg> <arg> [<arg>, ...]`  | {% raw %}`{{ eq "hello" "hello" }}`{% endraw %}  |
| Not equal               | `neq <arg> <arg> [<arg>, ...]` | {% raw %}`{{ neq "hello" "world" }}`{% endraw %} |
| Less than               | `lt <arg> <arg>`               | {% raw %}`{{ lt 1 2 }}`{% endraw %}              |
| Greater than            | `gt <arg> <arg>`               | {% raw %}`{{ gt 2 1 }}`{% endraw %}              |
| Less than or equal      | `le <arg> <arg>`               | {% raw %}`{{ le 1 2 }}`{% endraw %}              |
| Greater than or equal   | `ge <arg> <arg>`               | {% raw %}`{{ ge 2 1 }}`{% endraw %}              |

Example of combining various operators

{% raw %}

```
{{ and (eq true true) (neq true false) (not (empty "hello")) }}
```

{% endraw %}

## Conditionals

The `if/else` conditionals allows to perform templating only if specific conditions are met/not met, for example:

{% raw %}

```
{{ if $.Values.app.enabled }}
...
{{ end }}
```

{% endraw %}

A condition is considered *failed* if the result of its calculation is either of:

* boolean `false`;

* zero `0`;

* an empty string `""`;

* an empty list `[]`;

* an empty dictionary `{}`;

* null: `nil`.

In all other cases the condition is considered satisfied. A condition may include data, a variable, a function, or a pipeline.

Example:

{% raw %}

```
{{ if eq $appName "backend" }}
app: mybackend
{{ else if eq $appName "frontend" }}
app: myfrontend
{{ else }}
app: {{ $appName }}
{{ end }}
```

{% endraw %}

Simple conditionals can be implemented not only with `if/else`, but also with the `ternary` function. For example, the following `ternary` expression:

{% raw %}

```
{{ ternary "mybackend" $appName (eq $appName "backend") }}
```

{% endraw %}

... is similar to the `if/else' construction below:

{% raw %}

```
{{ if eq $appName "backend" }}
app: mybackend
{{ else }}
app: {{ $appName }}
{{ end }}
```

{% endraw %}

## Cycles

### Циклы по спискам

Циклы `range` позволяют перебирать элементы списка и выполнять нужную шаблонизацию на каждой итерации:

{% raw %}

```
{{ range $urls }}
{{ . }}
{{ end }}
```

{% endraw %}

Результат:

```
https://example.org
https://sub.example.org
```

Относительный контекст `.` всегда указывает на элемент списка, соответствующий текущей итерации, хотя указатель можно сохранить и в произвольную переменную:

{% raw %}

```
{{ range $elem := $urls }}
{{ $elem }}
{{ end }}
```

{% endraw %}

Результат будет таким же:

```
https://example.org
https://sub.example.org
```

Получить индекс элемента в списке можно следующим образом:

{% raw %}

```
{{ range $i, $elem := $urls }}
{{ $elem }} имеет индекс {{ $i }}
{{ end }}
```

{% endraw %}

Результат:

```
https://example.org имеет индекс 0
https://sub.example.org имеет индекс 1
```

### Циклы по словарям

Циклы `range` позволяют перебирать ключи и значения словарей и выполнять нужную шаблонизацию на каждой итерации:

```yaml
# values.yaml:
apps:
  backend:
    image: openjdk
  frontend:
    image: node
```

{% raw %}

```
# templates/app.yaml:
{{ range $.Values.apps }}
{{ .image }}
{{ end }}
```

{% endraw %}

Результат:

```
openjdk
node
```

Относительный контекст `.` всегда указывает на значение элемента словаря, соответствующего текущей итерации, при этом указатель можно сохранить и в произвольную переменную:

{% raw %}

```
{{ range $app := $.Values.apps }}
{{ $app.image }}
{{ end }}
```

{% endraw %}

Результат будет таким же:

```
openjdk
node
```

Получить ключ элемента словаря можно так:

{% raw %}

```
{{ range $appName, $app := $.Values.apps }}
{{ $appName }}: {{ $app.image }}
{{ end }}
```

{% endraw %}

Результат:

```yaml
backend: openjdk
frontend: node
```

### Контроль выполнения цикла

Специальное действие `continue` позволяет пропустить текущую итерацию цикла. В качестве примера пропустим итерацию для элемента `https://example.org`:

{% raw %}

```
{{ range $url := $urls }}
{{ if eq $url "https://example.org" }}{{ continue }}{{ end }}
{{ $url }}
{{ end }}
```

{% endraw %}

Специальное действие `break` позволяет не только пропустить текущую итерацию, но и прервать весь цикл:

{% raw %}

```
{{ range $url := $urls }}
{{ if eq $url "https://example.org" }}{{ break }}{{ end }}
{{ $url }}
{{ end }}
```

{% endraw %}

## Контекст

### Корневой контекст ($)

Корневой контекст — словарь, на который ссылается переменная `$`. Через него доступны values и некоторые специальные объекты. Корневой контекст имеет глобальную видимость в пределах файла-шаблона (исключение — блок `define` и некоторые функции).

Пример использования:

{% raw %}

```
{{ $.Values.mykey }}
```

{% endraw %}

Результат:

```
myvalue
```

К корневому контексту можно добавлять произвольные ключи/значения, которые также станут доступны из любого места файла-шаблона:

{% raw %}

```
{{ $_ := set $ "mykey" "myvalue"}}
{{ $.mykey }}
```

{% endraw %}

Результат:

```
myvalue
```

Корневой контекст остаётся неизменным даже в блоках, изменяющих относительный контекст (исключение — `define`):

{% raw %}

```
{{ with $.Values.backend }}
- command: {{ .command }}
  image: {{ $.Values.werf.image.backend }}
{{ end }}
```

{% endraw %}

Некоторые функции вроде `tpl` или `include` могут терять корневой контекст. Для сохранения доступа к корневому контексту многим из них можно передать корневой контекст аргументом:

{% raw %}

```
{{ tpl "{{ .Values.mykey }}" $ }}
```

{% endraw %}

Результат:

```
myvalue
```

### Относительный контекст (.)

Относительный контекст — данные любого типа, на которые ссылается переменная `.`. По умолчанию относительный контекст указывает на корневой контекст. 

Некоторые блоки и функции могут менять относительный контекст. В примере ниже в первой строке относительный контекст указывает на корневой контекст `$`, а во второй строке — уже на `$.Values.containers`:

{% raw %}

```
{{ range .Values.containers }}
{{ . }}
{{ end }}
```

{% endraw %}

Для смены относительного контекста можно использовать блок `with`:

{% raw %}

```
{{ with $.Values.app }}
image: {{ .image }}
{{ end }}
```

{% endraw %}

## Переиспользование шаблонов

### Именованные шаблоны

Для переиспользования шаблонов объявите *именованные шаблоны* в блоках `define` в файлах `templates/_*.tpl`:

{% raw %}

```
# templates/_helpers.tpl:
{{ define "labels" }}
app: myapp
team: alpha
{{ end }}
```

{% endraw %}

Далее подставляйте именованные шаблоны в файлы `templates/*.(yaml|tpl)` функцией `include`:

{% raw %}

```
# templates/deployment.yaml:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  selector:
    matchLabels: {{ include "labels" nil | nindent 6 }}
  template:
    metadata:
      labels: {{ include "labels" nil | nindent 8 }}
```

{% endraw %}

Результат:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  selector:
    matchLabels:
      app: myapp
      team: alpha
  template:
    metadata:
      labels:
        app: myapp
        team: alpha
```

Имя именованного шаблона для функции `include` может быть динамическим:

{% raw %}

```
{{ include (printf "%s.labels" $prefix) nil }}
```

{% endraw %}

**Именованные шаблоны обладают глобальной видимостью** — единожды объявленный в родительском или любом дочернем чарте именованный шаблон становится доступен сразу во всех чартах — и в родительском, и в дочерних. Убедитесь, что в подключенных родительском и дочерних чартах нет именованных шаблонов с одинаковыми именами.

### Параметризация именованных шаблонов

Функция `include`, подставляющая именованные шаблоны, принимает один произвольный аргумент. Его можно использовать для параметризации именованного шаблона, где этот аргумент станет относительным контекстом `.`:

{% raw %}

```
{{ include "labels" "myapp" }}
```

{% endraw %}

{% raw %}

```
{{ define "labels" }}
app: {{ . }}
{{ end }}
```

{% endraw %}

Результат:

```yaml
app: myapp
```

Для передачи сразу нескольких аргументов используйте список с несколькими значениями:

{% raw %}

```
{{ include "labels" (list "myapp" "alpha") }}
```

{% endraw %}

{% raw %}

```
{{ define "labels" }}
app: {{ index . 0 }}
team: {{ index . 1 }}
{{ end }}
```

{% endraw %}

... или словарь:

{% raw %}

```
{{ include "labels" (dict "app" "myapp" "team" "alpha") }}
```

{% endraw %}

{% raw %}

```
{{ define "labels" }}
app: {{ .app }}
team: {{ .team }}
{{ end }}
```

{% endraw %}

Необязательные позиционные аргументы можно реализовать так:

{% raw %}

```
{{ include "labels" (list "myapp") }}
{{ include "labels" (list "myapp" "alpha") }}
```

{% endraw %}

{% raw %}

```
{{ define "labels" }}
app: {{ index . 0 }}
{{ if gt (len .) 1 }}
team: {{ index . 1 }}
{{ end }}
{{ end }}
```

{% endraw %}

А необязательные непозиционные аргументы — так:

{% raw %}

```
{{ include "labels" (dict "app" "myapp") }}
{{ include "labels" (dict "team" "alpha" "app" "myapp") }}
```

{% endraw %}

{% raw %}

```
{{ define "labels" }}
app: {{ .app }}
{{ if hasKey . "team" }}
team: {{ .team }}
{{ end }}
{{ end }}
```

{% endraw %}

Именованному шаблону, не требующему параметризации, просто передайте `nil`:

{% raw %}

```
{{ include "labels" nil }}
```

{% endraw %}

### Результат выполнения include

Функция `include`, подставляющая именованный шаблон, **всегда возвращает только текст**. Для возврата структурированных данных нужно *десериализовать* результат выполнения `include` с помощью функции `fromYaml`:

{% raw %}

```
{{ define "commonLabels" }}
app: myapp
{{ end }}
```

{% endraw %}

{% raw %}

```
{{ $labels := include "commonLabels" nil | fromYaml }}
{{ $labels.app }}
```

{% endraw %}

Результат:

```
myapp
```

> Обратите внимание, что `fromYaml` не работает для списков. Используйте `fromYamlArray`.

Для явной сериализации данных можно воспользоваться функциями `toYaml` и `toJson`, для десериализации — функциями `fromYaml/fromYamlArray` и `fromJson/fromJsonArray`.

### Контекст именованных шаблонов

Объявленные в `templates/_*.tpl` именованные шаблоны теряют доступ к корневому и относительному контекстам файла, в который они включаются функцией `include`. Исправить это можно, передав корневой и/или относительный контекст в виде аргументов `include`:

{% raw %}

```
{{ include "labels" $ }}
{{ include "labels" . }}
{{ include "labels" (list $ .) }}
{{ include "labels" (list $ . "myapp") }}
```

{% endraw %}

### include в include

В блоках `define` тоже можно использовать функцию `include` для включения именованных шаблонов:

{% raw %}

```
{{ define "doSomething" }}
{{ include "doSomethingElse" . }}
{{ end }}
```

{% endraw %}

Через `include` можно вызвать даже тот именованный шаблон, из которого происходит вызов, т. е. вызвать его рекурсивно:

{% raw %}

```
{{ define "doRecursively" }}
{{ if ... }}
{{ include "doRecursively" . }}
{{ end }}
{{ end }}
```

{% endraw %}

## Шаблонизация с tpl

Функция `tpl` позволяет выполнить шаблонизацию любой строки и тут же получить результат. Она принимает один аргумент, который должен быть корневым контекстом.

Пример шаблонизации values:

{% raw %}

```yaml
# values.yaml:
appName: "myapp"
deploymentName: "{{ .Values.appName }}-deployment"
```

{% endraw %}

{% raw %}

```
# templates/app.yaml:
{{ tpl $.Values.deploymentName $ }}
```

{% endraw %}

Результат:

```
myapp-deployment
```

Пример шаблонизации произвольных файлов, которые сами по себе не поддерживают Helm-шаблонизацию:

{% raw %}

```
{{ tpl ($.Files.Get "nginx.conf") $ }} 
```

{% endraw %}

Для передачи дополнительных аргументов в функцию `tpl` можно добавить аргументы как новые ключи корневого контекста:

{% raw %}

```
{{ $_ := set $ "myarg" "myvalue"}}
{{ tpl "{{ $.myarg }}" $ }}
```

{% endraw %}

## Контроль отступов

Используйте функцию `nindent` для выставления отступов:

{% raw %}

```
       containers: {{ .Values.app.containers | nindent 6 }}
```

{% endraw %}

Результат:

```yaml
      containers:
      - name: backend
        image: openjdk
```

Пример комбинации с другими данными:

{% raw %}

```
       containers:
       {{ .Values.app.containers | nindent 6 }}
       - name: frontend
         image: node
```

{% endraw %}

Результат:

```yaml
      containers:
      - name: backend
        image: openjdk
      - name: frontend
        image: node
```

Используйте `-` после {% raw %}`{{`{% endraw %} и/или до {% raw %}`}}`{% endraw %} для удаления лишних пробелов до и/или после результата выполнения действия, например:

{% raw %}

```
  {{- "hello" -}} {{ "world" }}
```

{% endraw %}

Результат:

```
helloworld
```

## Комментарии

Поддерживаются два типа комментариев — комментарии шаблонизации {% raw %}`{{ /* */ }}`{% endraw %} и комментарии манифестов `#`.

### Комментарии шаблонизации

Комментарии шаблонизации скрываются при формировании манифестов:

{% raw %}

```
{{ /* Этот комментарий пропадёт */ }}
app: myApp
```

{% endraw %}

Комментарии могут быть многострочными:

{% raw %}

```
{{ /*
Hello
World
/* }}
```

{% endraw %}

Шаблоны в них игнорируются:

{% raw %}

```
{{ /*
{{ print "Эта шаблонизация игнорируется" }}
/* }}
```

{% endraw %}

### Комментарии манифестов

Комментарии манифестов сохраняются при формировании манифестов:

```yaml
# Этот комментарий сохранится
app: myApp
```

Комментарии могут быть только однострочнными:

```yaml
# Для многострочных комментариев используйте
# несколько однострочных комментариев подряд
```

Шаблоны в них выполняются:

{% raw %}

```
# {{ print "Эта шаблонизация выполняется" }}
```

{% endraw %}

## Отладка

Используйте `werf render`, чтобы полностью сформировать и отобразить конечные Kubernetes-манифесты. Укажите опцию `--debug`, чтобы увидеть манифесты, даже если они не являются корректным YAML.

Отобразить содержимое переменной:

{% raw %}

```
output: {{ $appName | toYaml }}
```

{% endraw %}

Отобразить содержимое переменной-списка или словаря:

{% raw %}

```
output: {{ $dictOrList | toYaml | nindent 2 }}
```

{% endraw %}

Отобразить тип данных у переменной:

{% raw %}

```
output: {{ kindOf $myvar }}
```

{% endraw %}

Отобразить произвольную строку, остановив дальнейшее формирование шаблонов:

{% raw %}

```
{{ fail (printf "Тип данных: %s" (kindOf $myvar)) }}
```

{% endraw %}
