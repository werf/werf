---
title: Секреты
permalink: advanced/helm/configuration/secrets.html
---

Для хранения в репозитории паролей, файлов сертификатов и т.п., рекомендуется использовать подсистему работы с секретами werf.

Идея заключается в том, что конфиденциальные данные должны храниться в репозитории вместе с приложением, и должны оставаться независимыми от какого-либо конкретного сервера.

werf поддерживает указание секретов следующими способами:
 - отдельный [values-файл для секретов]({{ "/advanced/helm/configuration/values.html#пользовательские-секреты" | true_relative_url }}) (`.helm/secret-values-yaml` по умолчанию или любой файл из репозитория, указанный опцией `--secret-values`).
 - секретные файлы — закодированные файлы в сыром виде без yaml, могут быть использованы в шаблонах.

## Ключ шифрования

Для шифрования и дешифрования данных необходим ключ шифрования. Есть два места откуда werf может прочитать этот ключ:
* из переменной окружения `WERF_SECRET_KEY`
* из специального файла `.werf_secret_key`, находящегося в корневой папке проекта
* из файла `~/.werf/global_secret_key` (глобальный ключ)

> Ключ шифрования должен иметь **шестнадцатеричный дамп** длиной 16, 24, или 32 байта для выбора соответственно алгоритмов AES-128, AES-192, или AES-256. Команда [werf helm secret generate-secret-key]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}) возвращает ключ шифрования, подходящий для использования алгоритма AES-128.

Вы можете быстро сгенерировать ключ, используя команду [werf helm secret generate-secret-key]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}).
### Работа с переменной окружения WERF_SECRET_KEY

Если при запуске werf доступна переменная окружения WERF_SECRET_KEY, то werf может использовать ключ шифрования из нее.

При работе локально, вы можете объявить ее с консоли. При работе с GitLab CI используйте [CI/CD Variables](https://docs.gitlab.com/ee/ci/variables/#variables) – они видны только участникам проекта с ролью master и не видны обычным разработчикам.

### Работа с файлом .werf_secret_key

Использование файла `.werf_secret_key` является более безопасным и удобным, т.к.:
* пользователям или инженерам, ответственным за запуск/релиз приложения не требуется добавлять ключ шифрования при каждом запуске;
* значения секрета из файла не будет отражено в истории команд консоли, например в файле `~/.bash_history`.

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
