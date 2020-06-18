---
title: Подключение базы данных
sidebar: applications-guide
permalink: documentation/guides/applications-guide/gitlab-java-springboot/080-database.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/templates/job.yaml
- .helm/templates/_envs.tpl
- .helm/requirements.yaml
- .helm/values.yaml
- .helm/secret-values.yaml
- config/database.yml
- .gitlab-ci.yml
{% endfilesused %}


Тут будут единые примеры, которые во всех статьях используются. Ибо для сборки и деплоя как-то не сильно разно получается, и хельм одинаковый.

<a name="database-app-generic" />

## Общий подход

TODO: куда-то делась

<a name="database-app-connection" />

## Как подключить БД

Разумеется, нашему приложению может потребоваться не только подключение к redis, но и к другим БД. Но от предыдущего пункта настройка принципиально не отличается. 
Описываем подключение application.properties, прописывая вместо реальных хостов переменные окружения. Правим helm-чарты, добавляя туда нужные переменные окружения и их значения.
Сгенерируем скелет приложения в start.spring.io. В зависимости добавим web и mysql. Скачаем, разархивируем. Все как и раньше. Либо можно просто добавить в pom.xml, который использовался для redis:

```
                <dependency>
                        <groupId>org.springframework.boot</groupId>
                        <artifactId>spring-boot-starter-data-jpa</artifactId>
                </dependency>
...
                <dependency>
                        <groupId>mysql</groupId>
                        <artifactId>mysql-connector-java</artifactId>
                        <scope>runtime</scope>
                </dependency>

```

werf.yaml подойдет от предыдущего шага с redis или тот что использовался в главе оптимизация сборки - они одинаковые. Без ассетов - для более простого восприятия.

Для наглядности результата поправим код приложения, как это описано в https://spring.io/guides/gs/accessing-data-mysql/. Только имя пакета оставим demo. Для единообразия.

Добавим в файл application.properties данные для подключения к mysql. Берем их, как обычно, из переменных окружения, которые передадим в контейнер при деплое.

```
spring.datasource.url=jdbc:mysql://${MYSQL_HOST:localhost}:3306/${MYSQL_DATABASE}
spring.datasource.username=${MYSQL_USER}
spring.datasource.password=${MYSQL_PASSWORD}
```

[application.properties](gitlab-java-springboot-files/03-demo-db/src/main/resources/application.properties:4-6)

Пропишем их в [values.yaml](gitlab-java-springboot-files/03-demo-db/.helm/values.yaml) и [secret-values.yaml](gitlab-java-springboot-files/03-demo-db/.helm/values.yaml), как было описано в прошлых главах. Для secret-values использовались следующие значения:

```yaml
mysql:
  password:
    _default: ThePassword
  rootpassword:
    _default: root
```

Также нам понадобится чарт с mysql. Опишем его сами в [statefulset](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) в helm-чарте. Причины почему стоит использовать именно statefulset для mysql в основном заключаются в требовниях к его стабильности и сохранности данных.
Внутрь контейнера с mysql так же передадим параметры для его работы через переменные окружения часть из которых прописана в values.yaml и шифрованная часть в secret-values.yaml.

```yaml
       env:
       - name: MYSQL_DATABASE
         value: {{ pluck .Values.global.env .Values.infr.mysql.db | first | default .Values.infr.mysql.db._default }}
       - name: MYSQL_PASSWORD
         value: {{ pluck .Values.global.env .Values.infr.mysql.password | first | default .Values.infr.mysql.password._default }}
       - name: MYSQL_ROOT_PASSWORD
         value: root
       - name: MYSQL_USER
         value: {{ pluck .Values.global.env .Values.infr.mysql.user | first | default .Values.infr.mysql.user._default }}
```

[mysql.yaml](gitlab-java-springboot-files/03-demo-db/.helm/templates/20-mysql.yaml:23-30)

Тут возникает еще нюанс - запуск нашего нового приложения полностью зависит от наличия mysql, и, если приложение не сможет достучаться до mysql - оно упадет с ошибкой. Для того чтобы этого избежать добавим init-container для нашего приложения, чтобы оно не запускалось раньше, чем появится коннекта к mysql (statefulset запустится.)

добавится следующая секция в deployment.yaml:

```yaml
    spec:
     initContainers:
     - name: wait-mysql
       image: alpine:3.9
       command:
       - /bin/sh
       - -c
       - while ! getent ahostsv4 $MYSQL_HOST; do echo waiting for mysql; sleep 2; done
```

[deployment.yaml](gitlab-java-springboot-files/03-demo-db/.helm/templates/10-deployment.yaml:16-22)

Все готово - собираем приложение аналогично тому что описано ранее. Делаем запрос на него:
`curl example.com/demo/all`

Получим следующий результат, что говорит о том, что приложение и mysql успешно общаются:

```shell
$ curl example.com/demo/all
[]
$ curl example.com/demo/add -d name=First -d email=someemail@someemailprovider.com
Saved
$ curl example.com/demo/add -d name=Second -d email=some2email@someemailprovider.com     
Saved
$ curl example.com/demo/all |jq
[
 {
   "id": 1,
   "name": "First",
   "email": "someemail@someemailprovider.com"
 },
 {
   "id": 2,
   "name": "Second",
   "email": "some2email@someemailprovider.com"
 }
]
```

<a name="database-migrations" />

## Выполнение миграций

Обычной практикой при использовании kubernetes является выполнение миграций в БД с помощью Job в кубернетес. Это единоразовое выполнение команды в контейнере с определенными параметрами.
Однако в нашем случае spring сам выполняет миграции в соответствии с правилами, которые описаны в нашем коде. Подробно о том как это делается и настраивается прописано в [документации](https://docs.spring.io/spring-boot/docs/2.1.1.RELEASE/reference/html/howto-database-initialization.html). Там рассматривается несколько инструментов для выполнения миграций и работы с бд вообще - Spring Batch Database, Flyway и Liquibase. С точки зрения helm-чартов или сборки совсем ничего не поменятся. Все доступы у нас уже прописаны.

<a name="database-fixtures" />

## Накатка фикстур

Аналогично предыдущему пункту про выполнение миграций, накатка фикстур производится фреймворком самостоятельно используя вышеназванные инструменты для миграций. Для накатывания фикстур в зависимости от стенда (dev-выкатываем, production - никогда не накатываем фикстуры) мы опять же можем передавать переменную окружения в приложение. Например, для Flyway это будет SPRING_FLYWAY_LOCATIONS, для Spring Batch Database нужно присвоить Java-переменной spring.batch.initialize-schema значение переменной из environment. В ней мы в зависимости от окружения проставляем либо always либо never. Реализации разные, но подход один - обязательно нужно принимать эту переменную извне. Документация та же что и в предыдущем пункте.


<div>
    <a href="090-unittesting.html" class="nav-btn">Далее: Юнит-тесты и Линтеры</a>
</div>
