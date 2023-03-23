---
title: Templates
permalink: usage/deploy/templates.html
---

## Templating

The templating mechanism in werf is the same as in Helm. It uses the [Go text/template](https://pkg.go.dev/text/template) template engine, enhanced with the [Sprig](https://masterminds.github.io/sprig/) and Helm feature set.

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
# ...
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

### Cycling through lists

The `range` cycles allow you to cycle through the list items and do the necessary templating at each iteration:

{% raw %}

```
{{ range $urls }}
{{ . }}
{{ end }}
```

{% endraw %}

Output:

```
https://example.org
https://sub.example.org
```

The `.` relative context always points to the list element that corresponds to the current iteration; the pointer can also be assigned to an arbitrary variable:

{% raw %}

```
{{ range $elem := $urls }}
{{ $elem }}
{{ end }}
```

{% endraw %}

The output is the same:

```
https://example.org
https://sub.example.org
```

Here's how you can get the index of an element in the list:

{% raw %}

```
{{ range $i, $elem := $urls }}
{{ $elem }} has an index of {{ $i }}
{{ end }}
```

{% endraw %}

Output:

```
https://example.org has an index of 0
https://sub.example.org has an index of 1
```

### Cycling through dictionaries

The `range` cycles allow you to cycle through the dictionary keys and values and do the necessary templating at each iteration:

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

Output:

```
openjdk
node
```

The `.` relative context always points to the value of the dictionary element that corresponds to the current iteration; the pointer can also be assigned to an arbitrary variable:

{% raw %}

```
{{ range $app := $.Values.apps }}
{{ $app.image }}
{{ end }}
```

{% endraw %}

The output is the same:

```
openjdk
node
```

Here's how you can get the key of the dictionary element:

{% raw %}

```
{{ range $appName, $app := $.Values.apps }}
{{ $appName }}: {{ $app.image }}
{{ end }}
```

{% endraw %}

Output:

```yaml
backend: openjdk
frontend: node
```

### Cycle control

The `continue` statement allows you to skip the current cycle iteration. To give you an example, let's skip the iteration for the `https://example.org` element:

{% raw %}

```
{{ range $url := $urls }}
{{ if eq $url "https://example.org" }}{{ continue }}{{ end }}
{{ $url }}
{{ end }}
```

{% endraw %}

In contrast, the `break` statement lets you both skip the current iteration and terminate the whole cycle:

{% raw %}

```
{{ range $url := $urls }}
{{ if eq $url "https://example.org" }}{{ break }}{{ end }}
{{ $url }}
{{ end }}
```

{% endraw %}

## Context

### Root context ($)

The root context is the dictionary to which the `$` variable refers. You can use it to access values and some special objects. The root context has global visibility within the template file (except for the `define` block and some functions).

Example of use:

{% raw %}

```
{{ $.Values.mykey }}
```

{% endraw %}

Output:

```
myvalue
```

You can add custom keys/values to the root context. They will also be available throughout the template file:

{% raw %}

```
{{ $_ := set $ "mykey" "myvalue"}}
{{ $.mykey }}
```

{% endraw %}

Output:

```
myvalue
```

The root context remains intact even in blocks that change the relative context (except for `define`):

{% raw %}

```
{{ with $.Values.backend }}
- command: {{ .command }}
  image: {{ $.Values.werf.image.backend }}
{{ end }}
```

{% endraw %}

Functions like `tpl` or `include` can lose the root context. You can pass the root context as an argument to them to restore access to it:

{% raw %}

```
{{ tpl "{{ .Values.mykey }}" $ }}
```

{% endraw %}

Output:

```
myvalue
```

### Relative context (.)

The relative context is any type of data referenced by the `.` variable. By default, the relative context points to the root context.

Some blocks and functions can modify the relative context. In the example below, in the first line, the relative context points to the root context `$`, while in the second line, it points to `$.Values.containers`:

{% raw %}

```
{{ range .Values.containers }}
{{ . }}
{{ end }}
```

{% endraw %}

Use the `with` block to modify the relative context:

{% raw %}

```
{{ with $.Values.app }}
image: {{ .image }}
{{ end }}
```

{% endraw %}

## Reusing templates

### Named templates

To reuse templating, declare *named templates* in the `define` blocks in the `templates/_*.tpl` files:

{% raw %}

```
# templates/_helpers.tpl:
{{ define "labels" }}
app: myapp
team: alpha
{{ end }}
```

{% endraw %}

Next, insert the named templates into the `templates/*.(yaml|tpl)` files using the `include` function:

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

Output:

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

The name of the named template to use in the `include` function may be dynamic:

{% raw %}

```
{{ include (printf "%s.labels" $prefix) nil }}
```

{% endraw %}

**Named templates are globally visible** - once declared in a parent or any child chart, a named template becomes available in all charts at once: in both parent and child charts. Make sure there are no named templates with the same name in the parent and child charts.

### Parameterizing named templates

The `include` function that inserts named templates takes a single optional argument. This argument can be used to parameterize a named template, where that argument becomes the `.` relative context:

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

Output:

```yaml
app: myapp
```

To pass several arguments at once, use a list containing multiple arguments:

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

...or a dictionary:

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

Optional positional arguments can be handled as follows:

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

Optional non-positional arguments can be handled as follows:

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

Pass `nil` to a named template that does not require parametrizing:

{% raw %}

```
{{ include "labels" nil }}
```

{% endraw %}

### The result of running include

The `include` function that inserts a named template **returns text data**. To return structured data, you need to *deserialize* the result of `include` using the `fromYaml` function:

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

Output:

```
myapp
```

> Note that `fromYaml` does not support lists. For lists, use the dedicated `fromYamlArray` function.

You can use the `toYaml` and `toJson` functions for data serialization, and the `fromYaml/fromYamlArray` and `fromJson/fromJsonArray` functions for deserialization.

### Named template context

The named templates declared in `templates/_*.tpl` cannot use the root and relative contexts of the file into which they are included by the `include` function. You can fix this by passing the root and/or relative context as `include` arguments:

{% raw %}

```
{{ include "labels" $ }}
{{ include "labels" . }}
{{ include "labels" (list $ .) }}
{{ include "labels" (list $ . "myapp") }}
```

{% endraw %}

### include in include

You can also use the `include` function in the `define` blocks to include named templates:

{% raw %}

```
{{ define "doSomething" }}
{{ include "doSomethingElse" . }}
{{ end }}
```

{% endraw %}

You can even call the `include` function to include a named template from this very template, i.e., recursively:

{% raw %}

```
{{ define "doRecursively" }}
{{ if ... }}
{{ include "doRecursively" . }}
{{ end }}
{{ end }}
```

{% endraw %}

## tpl templating

The `tpl` function allows you to process any line in real time. It takes one argument (the root context).

In the example below, we use the `tpl` function to retrieve the deployment name from the values.yaml file:

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

Output:

```
myapp-deployment
```

And here's how you can process arbitrary files that don't support Helm templating:

{% raw %}

```
{{ tpl ($.Files.Get "nginx.conf") $ }} 
```

{% endraw %}

You can add arguments as new root context keys to the `tpl` function to pass additional arguments:

{% raw %}

```
{{ $_ := set $ "myarg" "myvalue"}}
{{ tpl "{{ $.myarg }}" $ }}
```

{% endraw %}

## Indentation control

Use the `nindent` function to set the indentation:

{% raw %}

```
       containers: {{ .Values.app.containers | nindent 6 }}
```

{% endraw %}

Output:

```yaml
      containers:
      - name: backend
        image: openjdk
```

And here's how you can mix it with other data:

{% raw %}

```
       containers:
       {{ .Values.app.containers | nindent 6 }}
       - name: frontend
         image: node
```

{% endraw %}

Output:

```yaml
      containers:
      - name: backend
        image: openjdk
      - name: frontend
        image: node
```

Use `-` after {% raw %}`{{`{% endraw %} and/or before {% raw %}`}}`{% endraw %} to remove extra spaces before and/or after the action result, for example:

{% raw %}

```
  {{- "hello" -}} {{ "world" }}
```

{% endraw %}

Output:

```
helloworld
```

## Comments

werf supports two types of comments — template comments {% raw %}`{{ /* */ }}`{% endraw %}} and manifest comments `#`.

### Template comments

The template comments are stripped off during manifest generation:

{% raw %}

```
{{ /* This comment will be stripped off */ }}
app: myApp
```

{% endraw %}

Comments can be multi-line:

{% raw %}

```
{{ /*
Hello
World
/* }}
```

{% endraw %}

Template actions are ignored in such comments

{% raw %}

```
{{ /*
{{ print "This template action will be ignored" }}
/* }}
```

{% endraw %}

### Manifest comments

The manifest comments are retained during manifest generation:

```yaml
# This comment will stay in place
app: myApp
```

Only single-line comments of this type are supported:

```yaml
# For multi-line comments, use several
# single-line comments in a row
```

The template actions encountered in them are carried out:

{% raw %}

```
# {{ print "This template action will be carried out" }}
```

{% endraw %}

## Debugging

Use `werf render` to render and display ready-to-use Kubernetes manifests. The `--debug` option displays manifests even if they are not valid YAML.

Here's how you can display the variable contents:

{% raw %}

```
output: {{ $appName | toYaml }}
```

{% endraw %}

Display the contents of a list or dictionary variable:

{% raw %}

```
output: {{ $dictOrList | toYaml | nindent 2 }}
```

{% endraw %}

Display the variable's data type:

{% raw %}

```
output: {{ kindOf $myvar }}
```

{% endraw %}

Display some string and stop template rendering:

{% raw %}

```
{{ fail (printf "Data type: %s" (kindOf $myvar)) }}
```

{% endraw %}
