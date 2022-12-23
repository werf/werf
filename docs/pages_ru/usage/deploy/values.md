---
title: Параметризация шаблонов
permalink: usage/deploy/values.html
---

Под данными (или **values**) понимается произвольный YAML, заполненный парами ключ-значение или массивами, которые можно использовать в [шаблонах]({{ "/usage/deploy/templates.html" | true_relative_url }}). Все данные передаваемые в chart можно условно разбить на следующие категории:

- Обычные пользовательские данные.
- Пользовательские секреты.
- Сервисные данные.

## Обычные пользовательские данные

Для хранения обычных данных используйте файл чарта `.helm/values.yaml` (необязательно). Пример структуры:

```yaml
global:
  names:
  - alpha
  - beta
  - gamma
  mysql:
    staging:
      user: mysql-staging
    production:
      user: mysql-production
    _default:
      user: mysql-dev
      password: mysql-dev
```

Данные, размещенные внутри ключа `global`, будут доступны как в текущем чарте, так и во всех [вложенных чартах]({{ "/usage/deploy/charts.html" | true_relative_url }}) (сабчарты, subcharts).

Данные, размещенные внутри произвольного ключа `SOMEKEY` будут доступны в текущем чарте и во [вложенном чарте]({{ "/usage/deploy/charts.html" | true_relative_url }}) с именем `SOMEKEY`.

Файл `.helm/values.yaml` — файл по умолчанию для хранения данных. Данные также могут передаваться следующими способами:

* С помощью параметра `--values=PATH_TO_FILE` может быть указан отдельный файл с данными (может быть указано несколько параметров, по одному для каждого файла данных).
* С помощью параметров `--set key1.key2.key3.array[0]=one`, `--set key1.key2.key3.array[1]=two` могут быть указаны непосредственно пары ключ-значение (может быть указано несколько параметров, смотри также `--set-string key=forced_string_value`).

**ЗАМЕЧАНИЕ.** Все values-файлы, включая `.helm/values.yaml` и любые другие файлы, указанные с помощью опций `--values` — все должны быть коммитнуты в git репозиторий проекта согласно гитерминизму.

### Передача values в сабчарты

Чтобы передать данные из родительского чарта в сабчарт `mysubchart` необходимо определить следующие [values]({{ "/usage/deploy/values.html" | true_relative_url }}) в родительском чарте:

```yaml
mysubchart:
  key1:
    key2:
    - key3: value
```

В сабчарте `mysubchart` эти данные можно использовать с помощью обращения к соответствующим параметрам без указания ключа `mysubchart`:

{% raw %}

```yaml
{{ .Values.key1.key2[0].key3 }}
```

{% endraw %}

Данные, определенные глобально в ключе верхнего уровня `global`, также доступны в сабчартах:

```yaml
global:
  database:
    mysql:
      user: user
      password: password
```

Обращаться к ним необходимо как обычно:

{% raw %}

```yaml
{{ .Values.global.database.mysql.user }}
```

{% endraw %}

В сабчарте `mysubchart` будут доступны только данные ключей `mysubchart` и `global`.

**ЗАМЕЧАНИЕ** Файлы `secret-values.yaml` сабчартов не будут использоваться во время процесса деплоя, несмотря на то, что данные секретов из главного чарта и данные переданные через параметр `--secret-values` будут доступны через массив `.Values` как обычно.

### Передача динамических values из родительского чарта в сабчарты

Если вы хотите передать values, доступные только в родительском чарте, в сабчарты, то вам поможет директива `export-values`, которая имитирует (с небольшими отличиями) поведение [import-values](https://helm.sh/docs/topics/charts/#importing-child-values-via-dependencies), только вместо передачи values из сабчарта в родительский чарт она делает обратное: передает values в сабчарт из родительского чарта. Пример использования:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: backend
    version: 1.0.0
    export-values:
    - parent: werf.image.backend
      child: backend.image
```

Таким образом мы передадим в сабчарт всё, что доступно в `.Values.werf.image.backend` родительского чарта. В нашем случае это будет строка с репозиторием, именем и тегом образа `backend`, описанного в `werf.yaml`, который может выглядеть так: `example.org/backend/<имя_тега>`. В сабчарте эта строка станет доступна через `.Values.backend.image`:

{% raw %}

```yaml
# .helm/charts/backend/app.yaml
...
spec:
  template:
    spec:
      containers:
      - name: backend
        image: {{ .Values.backend.image }}  # Ожидаемый результат: `image: example.org/backend:<имя_тега>`
```

{% endraw %}

В отличие от YAML-якорей `export-values` будет работать с динамически выставляемыми values (сервисные данные werf), с переданными через командную строку values (`--set` и пр.) и с секретными values.

Также доступна альтернативная укороченная форма `export-values`, которая работает только для словарей (maps):

```yaml
    export-values:
    - "someMap"
```

Это будет эквивалентно следующей полной форме `export-values`:

```yaml
    export-values:
    - parent: exports.somemap
      child: .
```

Так в корень values сабчарта будут экспортированы все ключи, найденные в словаре `.Values.exports.somemap`.


### Параметры set

Имеется возможность переопределить значения values и передать новые values через параметры командной строки:

- `--set KEY=VALUE`;
- `--set-string KEY=VALUE`;
- `--set-file=PATH`;
- `--set-docker-config-json-value=true|false`.

**ЗАМЕЧАНИЕ.** Все файлы, указанные опцией `--set-file` должны быть коммитнуты в git репозиторий проекта согласно гитерминизму.

#### set-docker-config-json-value

При использовании параметра `--set-docker-config-json-value` werf выставит специальный value `.Values.dockerconfigjson` взяв текущий docker config из окружения где запущен werf (поддерживается переменная окружения `DOCKER_CONFIG`).

В данном значении `.Values.dockerconfigjson` будет содержимое конфига docker закодированное в base64 и пригодное для использования например при создании секретов для доступа к container registry:

{% raw %}

```
{{- if .Values.dockerconfigjson -}}
apiVersion: v1
kind: Secret
metadata:
  name: regsecret
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: {{ .Values.dockerconfigjson }}
{{- end -}}
```

{% endraw %}

**ВАЖНО!** Конфигурация docker текущего окружения, где запущен werf, может содержать доступы к registry, созданные с помощью короткоживущих токенов (например `CI_JOB_TOKEN` в GitLab). В таком случае использование `.Values.dockerconfigjson` для `imagePullSecrets` недопустимо, т.к. registry перестанет быть доступным из кластера Kubernetes как только завершится срок действия токена.

## Пользовательские секреты

Секреты, предназначенные для хранения конфиденциальных данных (паролей, сертификатов и других чувствительных к утечке данных), удобны для хранения прямо в репозитории проекта.

Для хранения секретов может использоваться дефолтный файл чарта `.helm/secret-values.yaml` (необязательно) или любое количество файлов с произвольным именем (`--secret-values`). Пример структуры:

```yaml
global:
  mysql:
    production:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
    staging:
      password: 100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123
```

Каждое значение в файле секретов (например, `100024fe29e45bf00665d3399f7545f4af63f09cc39790c239e16b1d597842161123`), представляет собой зашифрованные с помощью werf данные. Структура хранения секретов, такая же как и при хранении обычных данных, например, в `values.yaml`.

Файл `.helm/secret-values.yaml` — файл для хранения данных секретов по умолчанию. Данные также могут передаваться с помощью параметра `--secret-values=PATH_TO_FILE`, с помощью которого может быть указан отдельный файл с данными секретов (может быть указано несколько параметров, по одному для каждого файла данных секретов).

**ЗАМЕЧАНИЕ.** Все secret-values-файлы, включая `.helm/secret-values.yaml` и любые другие файлы, указанные с помощью опций `--secret-values` — все должны быть коммитнуты в git репозиторий проекта согласно гитерминизму.

## Сервисные данные

Сервисные данные генерируются werf автоматически для передачи дополнительной информации при рендеринге шаблонов чарта.

Пример структуры и значений сервисных данных werf:

```yaml
werf:
  name: myapp
  namespace: myapp-production
  env: production
  repo: registry.domain.com/apps/myapp
  image:
    assets: registry.domain.com/apps/myapp:a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
    rails: registry.domain.com/apps/myapp:e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
  tag:
    assets: a243949601ddc3d4133c4d5269ba23ed58cb8b18bf2b64047f35abd2-1598024377816
    rails: e760e9311f938e3d92681e93da3a81e176aa7f7e684ee06d092ec199-1598269478292
  commit:
    date:
      human: 2022-01-21 18:51:39 +0300 +0300
      unix: 1642780299
    hash: 1b28e6843a963c5bdb3579f6fc93317cc028051c

global:
  werf:
    name: myapp
    version: v1.2.7
```

Существуют следующие сервисные значения:

- Имя проекта из файла конфигурации `werf.yaml`: `.Values.werf.name`.
- Используемая версия werf: `.Values.werf.version`.
- Развертывание будет осуществлено в namespace `.Values.werf.namespace`.
- Название окружения CI/CD системы, используемое во время деплоя: `.Values.werf.env`.
- Адрес container registry репозитория, используемый во время деплоя: `.Values.werf.repo`.
- Полное имя и тег Docker-образа для каждого описанного в файле конфигурации `werf.yaml` образа: `.Values.werf.image.NAME`.
- Только теги собранных Docker-образов. Предназначены в первую очередь для использования совместно с `.Values.werf.repo`, для проброса полного имени и тега образов по-отдельности.
- Информация о коммите, с которого werf был запущен: `.Values.werf.commit.hash`, `.Values.werf.commit.date.human`, `.Values.werf.commit.date.unix`.

### Сервисные данные в сабчартах

Если вы используете [сабчарты]({{ "/usage/deploy/charts.html" | true_relative_url }}) и хотите использовать неглобальные сервисные данные (`.Values.werf`) в сабчарте, то эти сервисные данные потребуется явно экспортировать в сабчарт из родительского чарта:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: rails
    version: 1.0.0
    export-values:
    - parent: werf
      child: werf
```

Теперь сервисные данные, изначально доступные только на `.Values.werf` в родительском чарте, стали доступными по тому же пути (`.Values.werf`) и в сабчарте "rails". Обращайтесь к сервисным данным из сабчарта таким образом:

{% raw %}

```yaml
# .helm/charts/rails/app.yaml
...
spec:
  template:
    spec:
      containers:
      - name: rails
        image: {{ .Values.werf.image.rails }}  # Ожидаемый результат: `image: registry.domain.com/apps/myapp/rails:e760e931...`
```

{% endraw %}

Путь, по которому сервисные данные будут доступны после экспорта, можно изменить:

```yaml
    export-values:
    - parent: werf
      child: definitely.not.werf  # Сервисные данные станут доступными на `.Values.definitely.not.werf` в сабчарте.
```

Также можно экспортировать сервисные значения в сабчарт по отдельности:

```yaml
# .helm/Chart.yaml
apiVersion: v2
dependencies:
  - name: postgresql
    version: "10.9.4"
    repository: "https://charts.bitnami.com/bitnami"
    export-values:
    - parent: werf.repo
      child: image.repository
    - parent: werf.tag.my-postgresql
      child: image.tag
```

Больше информации про `export-values` можно найти [здесь]({{ "usage/deploy/values.html#передача-динамических-values-из-родительского-чарта-в-сабчарты" | true_relative_url }}).

## Итоговое объединение данных

<!-- This section could be in internals -->

Во время процесса деплоя werf объединяет все данные, включая секреты и сервисные данные, в единую структуру, которая передается на вход этапа рендеринга шаблонов (смотри подробнее [как использовать данные в шаблонах](#использование-данных-в-шаблонах)). Данные объединяются в следующем порядке (более свежее значение переопределяет предыдущее):

1. Данные из файла `.helm/values.yaml`.
2. Данные из параметров запуска `--values=PATH_TO_FILE`, в порядке указания параметров.
3. Данные секретов из файла `.helm/secret-values.yaml`.
4. Данные секретов из параметров запуска `--secret-values=PATH_TO_FILE`, в порядке указания параметров.
5. Данные из параметров `--set*`, указанные при вызове утилиты werf.
6. Сервисные данные.

## Использование данных в шаблонах

Для доступа к данным в шаблонах чарта используется следующий синтаксис:

{% raw %}

```yaml
{{ .Values.key.key.arraykey[INDEX].key }}
```

{% endraw %}

Объект `.Values` содержит [итоговый набор объединенных значений](#итоговое-объединение-данных).

## Секретные values и файлы

Для хранения в репозитории паролей, файлов сертификатов и т.п., рекомендуется использовать подсистему работы с секретами werf.

Идея заключается в том, что конфиденциальные данные должны храниться в репозитории вместе с приложением, и должны оставаться независимыми от какого-либо конкретного сервера.

werf поддерживает указание секретов следующими способами:
 - отдельный [values-файл для секретов]({{ "/usage/deploy/values.html#пользовательские-секреты" | true_relative_url }}) (`.helm/secret-values.yaml` по умолчанию или любой файл из репозитория, указанный опцией `--secret-values`).
 - секретные файлы — закодированные файлы в сыром виде без yaml, могут быть использованы в шаблонах.

## Ключ шифрования

Для шифрования и дешифрования данных необходим ключ шифрования. Есть три места откуда werf может прочитать этот ключ:
* из переменной окружения `WERF_SECRET_KEY`
* из специального файла `.werf_secret_key`, находящегося в корневой папке проекта
* из файла `~/.werf/global_secret_key` (глобальный ключ)

> Ключ шифрования должен иметь **шестнадцатеричный дамп** длиной 16, 24, или 32 байта для выбора соответственно алгоритмов AES-128, AES-192, или AES-256. Команда [werf helm secret generate-secret-key]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}) возвращает ключ шифрования, подходящий для использования алгоритма AES-128.

Вы можете быстро сгенерировать ключ, используя команду [werf helm secret generate-secret-key]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}).

> **ВНИМАНИЕ! Не сохраняйте файл `.werf_secret_key` в git-репозитории. Если вы это сделаете, то потеряете весь смысл шифрования, т.к. любой пользователь с доступом к git-репозиторию, сможет получить ключ шифрования. Поэтому, файл `.werf_secret_key` должен находиться в исключениях, т.е. в файле `.gitignore`!**

## Ротация ключа шифрования

werf поддерживает специальную процедуру смены ключа шифрования с помощью команды [`werf helm secret rotate-secret-key`]({{ "reference/cli/werf_helm_secret_rotate_secret_key.html" | true_relative_url }}).

## Secret values

Файлы с секретными переменными предназначены для хранения секретных данных в виде — `ключ: секрет`. **По умолчанию** werf использует для этого файл `.helm/secret-values.yaml`, но пользователь может указать любое число подобных файлов с помощью параметров запуска.

Файл с секретными переменными может выглядеть следующим образом:
```yaml
mysql:
  host: 10005968c24e593b9821eadd5ea1801eb6c9535bd2ba0f9bcfbcd647fddede9da0bf6e13de83eb80ebe3cad4
  user: 100016edd63bb1523366dc5fd971a23edae3e59885153ecb5ed89c3d31150349a4ff786760c886e5c0293990
  password: 10000ef541683fab215132687a63074796b3892d68000a33a4a3ddc673c3f4de81990ca654fca0130f17
  db: 1000db50be293432129acb741de54209a33bf479ae2e0f53462b5053c30da7584e31a589f5206cfa4a8e249d20
```

Для управления файлами с секретными переменными используйте следующие команды:
- [`werf helm secret values edit`]({{ "reference/cli/werf_helm_secret_values_edit.html" | true_relative_url }})
- [`werf helm secret values encrypt`]({{ "reference/cli/werf_helm_secret_values_encrypt.html" | true_relative_url }})
- [`werf helm secret values decrypt`]({{ "reference/cli/werf_helm_secret_values_decrypt.html" | true_relative_url }})

### Использование в шаблонах чарта

Значения секретных переменных расшифровываются в процессе деплоя и используются в Helm в качестве [дополнительных значений](https://helm.sh/docs/chart_template_guide/values_files/). Таким образом, использование секретов не отличается от использования данных в обычном случае:

{% raw %}
```yaml
...
env:
- name: MYSQL_USER
  value: {{ .Values.mysql.user }}
- name: MYSQL_PASSWORD
  value: {{ .Values.mysql.password }}
```
{% endraw %}

## Секретные файлы

Помимо использования секретов в переменных, в шаблонах также используются файлы, которые нельзя хранить незашифрованными в репозитории. Для размещения таких файлов выделен каталог `.helm/secret`, в котором должны храниться файлы с зашифрованным содержимым.

Чтобы использовать файлы содержащие секретную информацию в шаблонах Helm, вы должны сохранить их в соответствующем виде в каталоге `.helm/secret`.

Для управления файлами, содержащими секретную информацию, используйте следующие команды:
 - [`werf helm secret file edit`]({{ "reference/cli/werf_helm_secret_file_edit.html" | true_relative_url }})
 - [`werf helm secret file encrypt`]({{ "reference/cli/werf_helm_secret_file_encrypt.html" | true_relative_url }})
 - [`werf helm secret file decrypt`]({{ "reference/cli/werf_helm_secret_file_decrypt.html" | true_relative_url }})

> **ПРИМЕЧАНИЕ** werf расшифрует содержимое всех файлов в `.helm/secret` перед рендерингом шаблонов helm-чарта. Необходимо удостовериться что директория `.helm/secret` содержит корректные зашифрованные файлы.

### Использование в шаблонах чарта

<!-- Move to reference -->

Функция `werf_secret_file` позволяет использовать расшифрованное содержимое секретного файла в шаблоне. Обязательный аргумент функции путь к секретному файлу, относительно папки `.helm/secret`.

Пример использования секрета `.helm/secret/backend-saml/tls.key` в шаблоне:

{% raw %}
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myproject-backend-saml
type: kubernetes.io/tls
data:
  tls.crt: {{ werf_secret_file "backend-saml/stage/tls.crt" | b64enc }}
  tls.key: {{ werf_secret_file "backend-saml/stage/tls.key" | b64enc }}
```
{% endraw %}
