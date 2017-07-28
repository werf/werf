---
title: Работа с секретами
sidebar: doc_sidebar
permalink: secrets_for_kube.html
folder: kube
---

## Ключ шифрования

При шифровании данных необходим ключ, который может храниться в переменной окружения `DAPP_SECRET_KEY` либо в специальном файле в корне проекта `.dapp_secret_key`. 

Использование файла `.dapp_secret_key` удобнее и безопаснее, так как:
* пользователю/релиз-инженеру не требуется добавлять ключ шифрования при каждом запуске;
* секретное значение описанное в файле не может попасть в историю cli `~/.bash_history`.

Для генерации значения используется команда `dapp kube secret key generate`. 
Для удобства вывод команды уже содержит переменную окружения и может использоваться в команде `export`.
```
$ dapp kube secret key generate
DAPP_SECRET_KEY=c85e100d4ff006b693b0555f09244fdf
```

## Шифрование данных

* Для хранения секретных значений предусмотрен файл `.helm/secret-values.yaml`, который декодируется при развёртывании и используется в helm в качестве [дополнительных значений](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md) (значения декодируются в пустые строки при отсутствии ключа шифрования).
* Помимо секретных значений в шаблонах используются файлы, хранение в незащифрованном виде которых, недопустимо в репозитории. Для таких файлов зарезервирована директория `.helm/secret`, в которой должны храниться зашифрованные файлы. Используя метод `dapp_secret_file` (генерирует dapp `_dapp_helpers.tpl` в процессе развёртывания) в шаблоне, можно получить расшифрованное содержимое файла (метод вернёт пустую строку при отсутствии ключа шифрования).

При шифровании данных используется команда `dapp kube secret generate`.
```
$ dapp kube secret generate
Enter secret: 
1000541517bccae1acce015629f4ec89996e0b4
```

Также команда поддерживает перенаправленный вывод, результат выполнения других команд.
```
$ rake magic | dapp kube secret generate
1000541517bccae1acce015629f4ec89996e0b4
```

Раздел файла с зашифрованными значениями внешней mysql-базы может быть организован следующим образом:

```
mysql:
  host: 100070c0e52ba2ff965ebd85f5fea9549392294e52aca006cf75
  user: 2ad80161428063803509eba8e9909ddcd0db0ddaada!b9ee47
  password: 80161428063803509eba8e9909ddcd0db0ddaab9ee47
  db: 406d3a4d2282ad80161428063803509eba8e9909ddcd0db0ddaab9ee47
```

При шифровании файлов необходимо указать путь до файла.

```
$ dapp kube secret generate ~/certs/tls.key
100023b2d1c0ec145681183ec721dc06db34f7ebce9f328739f0350d7f3aea988b6d0b69e9f71ed5e2ad9d79449b7a7d830ee5148a30a50bd43b7e2ecaef1c657199a483f60322cf7727ddf3928b2f51b0fbb0b1cd931489c20061a5071cf4362cb7e91c79fdbfc6d950352535eac28affd47d8ea8af64559fa39d89e815ea2b95cb07e81ddba792bf0e834cbbdc2ef843394a23f0cd44a95a38dd1583c2ae8352af140fc3fcfa6da3485bbf9bd286e2864ad45e31bc8ce4239aa05aaa82beba58c0583d3e93141ae28d87f4ffdb3d089f18b86e42e88a0b065c604f92a1478e0bbaeee46136579895b803a4be80977135979c4022b83fb1787e7b1540ddc07cd287ba5a7442f8a3ce0f5177487751c25767c28fd6eacb7f021036d978301895d6f528f06d555c926ba617669348c7873ba98372ae75ee0fdb730cabe507c576371970a27476e557b8b250f83137535f1d466eb53756986160f75ef78075dd7f63f83d72c1daf04aa026000802d4bbc2832f6d63eb231b8e16af5f44fc2cd79220715cba783a495a9d25e778ec1c2aa8013ccc164b5fc51f3a061c1eeed1228f65867c25f962639c90d2398e48ad93744cab5f8fff1f9988ccdbc5778ff39c31bdd47950759f33bf126509d3105521571252823f523fcd4a478d9bce3ddf923f8f8cbe7bff5edc0e99fe908e8b737a6de2391729e6ada3d8069819a0857ceba1eb5a16ecc81d6bcd16e497c4e60af5d218d2d2e0064c07850e5aa2a8d83e0f0a2
```

Для использования этих данных в helm template-ах необходимо сохранить их в соответствующем файле в директории `.helm/secret`.

Вызвав команду с опцией `-o OUTPUT_FILE_PATH`, можно cохранить зашифрованные данные в файл.

```
$ dapp kube secret generate ~/certs/tls.key -o .helm/secret/backend-saml/tls.key
```

Использование секрета в шаблоне может выглядеть следующим образом.

```
...
data:
  tls.key: {% raw %}{{ tuple "/backend-saml/tls.key" . | include "dapp_secret_file" | b64enc }}{% endraw %}
```

Для шифрования файла с секретными значениями необходимо пробросить путь до файла и опцию `--values`.

```
$ dapp kube secret generate secret-values.yaml --values
mysql:
  user: 2ad80161428063803509eba8e9909ddcd0db0ddaada
  password: 80161428063803509eba8e9909ddcd0db0ddaab9ee47
  db: 406d3a4d2282ad80161428063803509eba8e9909ddcd0db0ddaab9ee47
```

## Редактирование зашифрованных данных

Для редактирования существующих секретов можно воспользоваться командой `dapp kube secret edit`.
Команда позволяет работать с данными в интерактивном режиме.

## Обратное преобразование данных

Используя команду `dapp kube secret extract` можно расшифровать зашифрованные ранее значения.

```
$ dapp kube secret extract
Enter secret: 1000541517bccae1acce015629f4ec89996e0b4
42
```

Также как и при шифровании поддерживаются перенаправленный вывод и секреты из файлов.

```
$ echo "1000541517bccae1acce015629f4ec89996e0b4" | dapp kube secret extract
42
```

```
$ dapp kube secret extract .helm/secret/sense_of_life.txt
42 — ответ на «главный вопрос Жизни, Вселенной и Всего Остального».
```

Если необходимо расшифровать secret-values файл, то необходимо также указать опцию `--values`.
```
$ dapp kube secret extract .helm/secret-values.yaml --values
sense:
  of:
    life: 42
    lifes: [42, 42, 42]
```

## Регенерация существующих секретов

При запуске команды будут перегенерированы секреты (`.helm/secret/**/*`) и секретные значения (`.helm/secret-values.yaml`). При генерации используется текущий ключ и ключ (`--old-secret-key KEY`), которым были закодированы данные.
```
$ dapp kube secret regenerate --old-secret-key c85e100d4ff006b693b0555f09244fdf
```

В случае, если секретные значения хранятся в нескольких файлах, необходимо добавить пути в качестве аргументов.

```
$ dapp kube secret regenerate --old-secret-key c85e100d4ff006b693b0555f09244fdf .helm/secret-values2.yaml .helm/secret-staging.yaml
```
