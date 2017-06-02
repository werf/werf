---
title: Работа с секретами
sidebar: doc_sidebar
permalink: secrets_for_deploy.html
folder: deploy
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

## Шифрование значений

Для хранения секретных значений предусмотрен файл `.helm/secret-values.yaml`, который декодируется при развёртывании и используется в helm в качестве [дополнительных значений](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md) (при отсутствии ключа шифрования файл используется, но не декодируется)

При шифрованнии данных используется команда `dapp kube secret generate`. Завершить ввод необходимо нажатием `Ctrl + D`.
```
$ dapp kube secret generate
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

## Шифрование файлов

Помимо секретных значений в шаблонах используются файлы, хранение в незащифрованном виде которых, недопустимо в репозитории. Для таких файлов зарезервирована директория `.helm/secret`, в которой должны храниться зашифрованные файлы.  

Используя метод `dapp_secret_file` (генерирует dapp `_dapp_helpers.tpl` в процессе развёртывания) в шаблоне, можно получить расшифрованное содержимое файла. 
При отсутствии ключа шифрования метод вернёт зашифрованные данные.

Для шифрования файлов предусмотрен метод `dapp kube secret file encrypt`. В качестве аргумента принимается путь до файла.

```
$ dapp kube secret file encrypt ~/certs/tls.key
100023b2d1c0ec145681183ec721dc06db34f7ebce9f328739f0350d7f3aea988b6d0b69e9f71ed5e2ad9d79449b7a7d830ee5148a30a50bd43b7e2ecaef1c657199a483f60322cf7727ddf3928b2f51b0fbb0b1cd931489c20061a5071cf4362cb7e91c79fdbfc6d950352535eac28affd47d8ea8af64559fa39d89e815ea2b95cb07e81ddba792bf0e834cbbdc2ef843394a23f0cd44a95a38dd1583c2ae8352af140fc3fcfa6da3485bbf9bd286e2864ad45e31bc8ce4239aa05aaa82beba58c0583d3e93141ae28d87f4ffdb3d089f18b86e42e88a0b065c604f92a1478e0bbaeee46136579895b803a4be80977135979c4022b83fb1787e7b1540ddc07cd287ba5a7442f8a3ce0f5177487751c25767c28fd6eacb7f021036d978301895d6f528f06d555c926ba617669348c7873ba98372ae75ee0fdb730cabe507c576371970a27476e557b8b250f83137535f1d466eb53756986160f75ef78075dd7f63f83d72c1daf04aa026000802d4bbc2832f6d63eb231b8e16af5f44fc2cd79220715cba783a495a9d25e778ec1c2aa8013ccc164b5fc51f3a061c1eeed1228f65867c25f962639c90d2398e48ad93744cab5f8fff1f9988ccdbc5778ff39c31bdd47950759f33bf126509d3105521571252823f523fcd4a478d9bce3ddf923f8f8cbe7bff5edc0e99fe908e8b737a6de2391729e6ada3d8069819a0857ceba1eb5a16ecc81d6bcd16e497c4e60af5d218d2d2e0064c07850e5aa2a8d83e0f0a2
```

Полученное значение записывается в файл в директории `.helm/secret`.
Альтернативным решением может быть использование опции `-o OUTPUT_FILE_PATH`.

```
$ dapp kube secret file encrypt ~/certs/tls.key -o .helm/secret/backend-saml/tls.key
```

Использование секрета в шаблоне может выглядить следующим образом.

```
...
data:
  tls.key: {% raw %}{{ tuple "/backend-saml/tls.key" . | include "dapp_secret_file" | b64enc }}{% endraw %}
```
