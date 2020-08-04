---
title: Работа с файлами
sidebar: applications_guide
guide_code: ____________
permalink: documentation/guides/applications_guide/____________/050_files.html
layout: guide
toc: false
---

{% filesused title="Файлы, упомянутые в главе" %}
- .helm/templates/deployment.yaml
- .helm/secret-values.yaml
- ____________
- ____________
{% endfilesused %}

В этой главе мы настроим в нашем базовом приложении работу с пользовательскими файлами. Для этого нам нужно персистентное хранилище.

В идеале — нужно добиться, чтобы приложение было stateless, а данные хранились в S3-совместимом хранилище, например minio или aws s3. Это обеспечивает простое масштабирование, работу в HA режиме и высокую доступность.

{% offtopic title="А есть какие-то способы кроме S3?" %}
Первый и более общий способ — это использовать как [volume](https://kubernetes.io/docs/concepts/storage/volumes/) хранилище [NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs), [CephFS](https://kubernetes.io/docs/concepts/storage/volumes/#cephfs) или [hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath).

Мы не рекомендуем этот способ, потому что при возникновении неполадок с такими типами volume’ов мы будем влиять на работоспособность контейнера и всего демона docker в целом, тем самым могут пострадать приложения, не имеющие никакого отношения к вашему.

Более надёжный путь — пользоваться S3. Таким образом мы используем отдельный сервис, который имеет возможность масштабироваться, работать в HA режиме, и иметь высокую доступность. Можно воспользоваться cloud решением, таким, как AWS S3, Google Cloud Storage, Microsoft Blobs Storage и т.д.

Если мы будем сохранять файлы в какой-либо директории у приложения запущенного в Kubernetes, то после перезапуска контейнера все изменения пропадут.
{% endofftopic %}

Данная настройка производится полностью в рамках приложения, а нам остается только передать необходимые переменные окружения при запуске приложения.

____________
____________
____________
____________
____________
____________

Для работы с S3 необходимо пробросить в ключи доступа в приложение. Для этого стоит использовать [механизм секретных переменных](https://ru.werf.io/documentation/reference/deploy_process/working_with_secrets.html). *Вопрос работы с секретными переменными рассматривался подробнее, [когда мы делали базовое приложение](020_basic.html#secret-values-yaml)*

{% snippetcut name="secret-values.yaml (расшифрованный)" url="#" %}
{% raw %}
```yaml
app:
  ____________
  ____________
  ____________
```
{% endraw %}
{% endsnippetcut %}

А не секретные значения — храним в `values.yaml`

{% snippetcut name="values.yaml" url="#" %}
{% raw %}
```yaml
  ____________
  ____________
  ____________
```
{% endraw %}
{% endsnippetcut %}

После того, как значения корректно прописаны и зашифрованы — мы можем пробросить соответствующие значения в Deployment.

{% snippetcut name="deployment.yaml" url="#" %}
{% raw %}
```yaml
        ____________
        ____________
        ____________
```
{% endraw %}
{% endsnippetcut %}

____________
____________
____________

<div>
    <a href="060_email.html" class="nav-btn">Далее: Работа с электронной почтой</a>
</div>
