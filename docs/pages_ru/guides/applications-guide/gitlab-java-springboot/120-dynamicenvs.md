---
title: Динамические окружения
sidebar: applications-guide
permalink: documentation/guides/applications-guide/template/120-dynamicenvs.html
layout: guide
---

{% filesused title="Файлы, упомянутые в главе" %}
- .gitlab-ci.yaml
{% endfilesused %}

Во время разработки и эксплуатации приложения может потребоваться использовать не только условные stage и production окружения. Зачастую, удобно что-то разрабатывать в изолированном от других задач стенде.
Поскольку у нас уже готовы описание сборки, helm-чарты, достаточно прописать запуск таких стендов (review-стенды или feature-стенды) из веток. Для этого добавим в .gitlab-ci.yaml кнопки для [start](gitlab-java-springboot-files/05-demo-complete/.gitlab-ci.yaml:62-71):

```yaml
Deploy to Review:
  extends: .base_deploy
  stage: deploy
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    url: review-${CI_COMMIT_REF_SLUG}.example.com
    on_stop: Stop Review
  only:
    - /^feature-*/
  when: manual
```

и [стоп](gitlab-java-springboot-files/05-demo-complete/.gitlab-ci.yaml:73-82) этих стендов:
```yaml
Stop Review:
  stage: deploy
  script:
    - werf dismiss --env $CI_ENVIRONMENT_SLUG --namespace ${CI_ENVIRONMENT_SLUG} --with-namespace
  environment:
    name: review/${CI_COMMIT_REF_SLUG}
    action: stop
  only:
    - /^feature-*/
  when: manual
```

здесь мы видим новую для нас команду [werf dismiss](https://werf.io/documentation/cli/main/dismiss.html). Она удалит приложение из kubernetes, helm-релиз так же будет удален вместе с namespace в который выкатывалось приложение.

В .gitlab-ci, как уже упоминалось ранее удобно передавать environment url - его видно на вкладке environments. Можно не вспоминая у какой задачи какая ветка визуально её найти и нажать на ссылку в gitlab.

Нам сейчас нужно использовать этот механизм для формирования наших доменов в ingress. И передать в приложение, если требуется.
Передадим .helm в .base_deploy это переменную аналогично тому что уже сделано с environment:


```
     --set "global.env=${CI_ENVIRONMENT_SLUG}"
     --set "global.ci_url=$(basename ${CI_ENVIRONMENT_URL})"

```

* CI_ENVIRONMENT_SLUG - встроенная перменная gitlab, "очишенная" от некорректных с точки зрения DNS, URL. В нашем случае позволяет избавиться от символов, которые kubernetes не может воспринять - слеши, подчеркивания. Заменяет на "-".
* CI_ENVIRONMENT_URL - мы прокидываем переменную которую указали в environment url.

В .helm сделали себе переменную ci_url. Теперь можно использовать её в [ingress](gitlab-java-springboot-files/05-demo-complete/.helm/templates/90-ingress.yaml:10):

```
  - host: {{ .Values.global.ci_url }}
```

или в любом месте где объявляется она - например, как в случае с [ассетами](gitlab-java-springboot-files/05-demo-complete/.helm/templates/01-cm.yaml:19)

При выкате такого окружения мы получим полноценное окружение - redis, mysql, ассеты. Все это собирается одинаково, что для stage, что для review. 

Разумеется, в production не стоит использовать БД в кубернетес - её нужно будет исключить из шаблона, как показано в примере. Для всего что не production - выкатываем бд в кубернетес. А для всего остального - service и связанный с ним объект endpoint -подробнее -в [документации kubernetes](https://kubernetes.io/docs/concepts/services-networking/service/). И в репозитории - в шаблоне [mysql](gitlab-java-springboot-files/05-demo-complete/.helm/templates/20-mysql.yaml:54-78) и в [values](gitlab-java-springboot-files/05-demo-complete/.helm/values.yaml:18-21), куда добавляются данные для подключения к внешней БД. Но с точки зрения работы приложения - это никак его не касается - он так же ходит на хост mysql. Все остальные параметры продолжают браться из старых перменных.

Таким образом мы можем автоматически заводить временные окружения в кубернетес и автоматически их закрывать при завершении тестирования задачи. При этом использовать один и тот же образ начиная с review вплоть до production, что позволяет нам тестировать именно то, что в итоге будет работать в production.
