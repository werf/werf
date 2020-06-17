---
title: Генерируем и раздаем ассеты
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/040-assets.html
layout: guide
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/service.yaml
- .helm/templates/ingress.yaml
- werf.yaml
{% endfilesused %}


В какой-то момент в процессе разработки вам понадобятся ассеты (т.е. картинки, css, js).

Для генерации ассетов мы будем использовать webpack.
Генерировать ассеты для java-spring-maven можно, конечно, разными способами. Например, в maven есть [плагин](https://github.com/eirslett/frontend-maven-plugin), который позволяет описать сборку ассетов "не выходя" из Java. Но там есть несколько оговорок про use-case этого плагина:

*   Не предполагается использовать как замена Node для разработчиков фронтенда. Скорее для того чтобы разработчики бекенда могли быстрее включить JS-код в свою сборку.
*   Не предполагается использование на production-окружениях.

Потому хорошим и распространенным выбором будет использовать webpack отдельно.

Интуитивно понятно, что на стадии сборки нам надо будет вызвать скрипт, который генерирует файлы, т.е. что-то надо будет дописать в `werf.yaml`. Однако, не только там — ведь какое-то приложение в production должно непосредственно отдавать статические файлы. Мы не будем отдавать файлики с помощью {{Frameworkname}}. Хочется, чтобы статику раздавал nginx. А значит надо будет внести какие-то изменения и в helm чарты.

<a name="assets-scenario" />

## Сценарий сборки ассетов

webpack - гибкий в плане реализации ассетов инструмент. Настраивается его поведение в webpack.config.js и package.json.
Создадим в этом же проекте папку [assets](gitlab-java-springboot-files/02-demo-with-assets/assets/). В ней следующая структура

```
├── default.conf
├── dist
│   └── index.html
├── package.json
├── src
│   └── index.js
└── webpack.config.js
```

При настройке ассетов есть один нюанс - при подготовке ассетов мы не рекомендуем использовать какие-либо изменяемые переменные _на этапе сборки_. Потому что собранный бинарный образ должен быть независимым от конкретного окружения. А значит во время сборки у нас не может быть, например, указано домена для которого производится сборка, user-generated контента и подобных вещей.

В связи с описанным выше, все ассеты должны быть сгенерированы одинаково для всех окружений. А вот использовать их стоит в зависимости от окружения в котором запущено приложение. Для этого можно подсунуть приложению уже на этапе деплоя конфиг, который будет зависеть от окружения. Реализуется это через [configmap](https://kubernetes.io/docs/concepts/configuration/configmap/). Кстати, он же будет нам полезен, чтобы положить конфиг nginx внутрь alpine-контейнера. Обо всем этом далее.

<a name="assets-implementation" />

## Какие изменения необходимо внести

Генерация ассетов происходит в отдельном артефакте в 2-х стадиях - `install` и `setup`. На первой стадии мы выполняем `npm install`, на второй уже `npm build`. Не забываем про кеширование стадий и зависимость сборки от изменения определенных файлов в репозитории.
Так же нам потребуются изменения в шаблонах helm. Нам нужно будет описать процесс запуска контейнера с ассетами. Мы не будем их класть в контейнер с приложением. Запустим отдельный контейнер с nginx и в нем уже будем раздавать статику. Соответственно нам потребуются изменения в deployment. Так же потребуется добавить configmap и описать в нем файл с переменными для JS и конфиг nginx. Так же, чтобы трафик попадал по нужному адресу - Потребуются правки в ingress и service.

### Изменения в сборке

В конфиги сборке (werf.yaml) потребуется описать артефакт для сборки и импорт результатов сборки в контейнер с nginx. Делается это аналогично сборке на Java, разумеется используются команды сборки специфичные для webpack/nodejs. Из интересеного - мы добавляем не весь репозиторий для сборки, а только содержимое assets.

```yaml
git:
  - add: /assets
    to: /app
    excludePaths:

    stageDependencies:
      install:
      - package.json
      - webpack.config.js
      setup:
      - "**/*"
```

[werf.yaml](gitlab-java-springboot-files/02-demo-with-assets/werf.yaml:42-53)

Для артефакта сборки ассетов настроили запуск `npm install` в случае изменения `package.json` и `webpack.config.js`. А так же запуск `npm run build` при любом изменении файла в репозитории.

Также стоит исключить assets из сборки Java:

```yaml
git:
- add: /
  to: /app
  excludePaths:
  - assets
```

[werf.yaml](gitlab-java-springboot-files/02-demo-with-assets/werf.yaml:7-10)

Получившийся общий werf.yaml можно посмотреть [в репозитории]([werf.yaml](02-demo-with-assets/werf.yaml)).

### Изменения в деплое

Статику логично раздавать nginx-ом. Значит нам нужно запустить nginx в том же поде с приложением но в отдельном контейнере.

В deployment допишем что контейнер будет называться frontend, будет слушать на 80 порту

```yaml
      containers:
      - name: frontend
{{ tuple "frontend" . | include "werf_container_image" | indent 8 }}
        ports:
        - name: http-front
          containerPort: 80
          protocol: TCP
```

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:23-29)

Разумеется, нам нужно положить nginx-конфиг для раздачи ассетов внутрь контейнера. Поскольку в нем могут (в итоге, например для доступа к s3) использоваться переменные - добавляем именно configmap, а не подкладываем файл на этапе сборки образа.
Здесь же добавим js-файл, содердащий переменные, к которым обращается js во время выполнения. Подкладываем его на этапе деплоя именно для того, чтобы иметь гибкость в работе с ассетами в условиях одинаковых исходных образов для разных окружений.

```yaml
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx/conf.d/default.conf
          subPath: default.conf
        - name: env-js
          mountPath: /app/dist/env.js
          subPath: env.js
```

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:30-36)

Не забываем указать что файлы мы эти берем из определенного конфигмапа:

```yaml
      volumes:
      - name: nginx-config
        configMap:
          name: {{ .Chart.Name }}-configmap
      - name: env-js
        configMap:
          name: {{ .Chart.Name }}-configmap
```

[deployment.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/10-deployment.yaml:52-58)


Здесь мы добавили подключение configmap к deployment. В самом configmap пропишем конфиг nginx - default.conf и файл с переменными для js - env.js. Вот так, например выглядит конфиг с переменной для JS:

```yaml
...
  env.js: |
    module.exports = {
        url: {{ pluck .Values.global.env .Values.app.url | first | default .Values.app.url._default |quote }},
        environment: {{ .Values.global.env |quote }}
    };

```

[Полный файл](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/01-cm.yaml)

Еще раз обращу внимание на env.js - мы присваиваем url значение в зависимости от окружения указанное в values.yaml. Как раз та "изменчивость" js что нам нужна.

Так же, чтобы kubernetes мог направить трафик в nginx с ассетами нужно дописать port 80 в service, а затем мы направим внешний трафик предназначенный для ассетов в этот сервис.

```yaml
...
  - name: http-front
    port: 80
    targetPort: 80
```

[service.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/20-service.yaml:10-12)

### Изменения в роутинге

В ingress направим все что связано с ассетами на порт 80, чтобы все запросы к, в нашем случае example.com/static/ попали в нужный контейнер на 80-ый порт, где у нас будет отвечать nginx, прописанный выше.

```yaml
...
      - path: /static/
        backend:
          serviceName: {{ .Chart.Name | quote }}
          servicePort: 80
```

[ingress.yaml](gitlab-java-springboot-files/02-demo-with-assets/.helm/templates/90-ingress.yaml:17-20)



