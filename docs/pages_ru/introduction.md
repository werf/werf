---
title: Введение
permalink: introduction.html
sidebar: documentation
---

{% asset introduction.css %}
{% asset introduction.js %}

## Что такое werf?

werf - это CLI-инструмент для организации полного цикла развертывания приложения с Git в качестве единого и универсального "источника истины". werf может:

 - Собирать docker-образы.
 - Деплоить приложение в кластер Kubernetes.
 - Убеждаться в том, что приложение запустилось и нормально работает после завершения развертывания.
 - Пересобирать docker-образы при внесении изменений в код приложения.
 - Ре-деплоить приложение в кластер Kubernetes при необходимости.
 - Удалять ненужные и неиспользуемые образы.

## Как werf работает?

<div id="introduction-presentation" class="introduction-presentation">
    <div class="introduction-presentation__container">
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-1.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                Конфигурация werf должна храниться в Git-репозитории приложения вместе с его кодом.
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-2.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    1. Копируем Dockerfiles в репозиторий приложения
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-3.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    2. Создаем файл конфигурации <code>werf.yaml</code>
                </div>
<div markdown="1">
Обратите особое внимание на параметр `project` - он содержит _название проекта_. В дальнейшем werf будет активно его использовать во время _converge-процесса_. Изменение этого параметра потом, когда приложение уже развернуто и работает, будет связано с простоем и потребует ре-деплоя приложения.
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-4.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    3. Описываем шаблоны helm-чартов для развертывания приложения
                </div>
<div markdown="1">
Специальная шаблон-функция `werf_image` позволяет сгенерировать полное имя собираемого образа. У этой функции имеется параметр name, который соответствует образу, определенному в `werf.yaml` (`"frontend"` или `"backend"` в нашем примере).
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/dds-5.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    4. Делаем коммит
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-1.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    5. Вычисляем сконфигурированные образы исходя из текущего состояния Git-коммита.
                </div>
<div markdown="1">
На этом шаге werf генерирует имена целевых образов. Имена могут менять или оставаться прежними после очередного коммита в зависимости от изменений в репозитории Git. Обратите внимание, что имена образов детерминированы и воспроизводимы и привязаны к соответствующему коммиту.
</div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-2.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    6. Считываем образы, доступные в Docker Registry.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-3.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    7. Вычисляем разницу между образами, соответствующими состоянию для текущего Git-коммита и теми, которые уже доступны в реестре Docker.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-4.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    8. Собираем и публикуем только те образы, которые отсутствуют Docker Registry (если такие имеются).
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-5.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    9. Считываем целевую конфигурацию ресурсов Kubernetes, связанную с текущим состоянием, определяемым Git-коммитом.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-6.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    10. Считываем конфигурацию имеющихся в кластере Kubernetes-ресурсов.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-7.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    11. Подсчитываем разницу между целевой для текущего состояния Git и имеющейся конфигурацией Kubernetes-ресурсов в кластере.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-8.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    12. Применяем изменения (при необходимости) к ресурсам Kubernetes, чтобы они соответствовали состоянию, определенному в Git-коммите.
                </div>
            </div>
        </div>
        <div class="introduction-presentation__slide">
            <img src="{% asset introduction/c-9.png @path %}"
            class="introduction-presentation__slide-img" />
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title">
                    13. Убеждаемся в работоспособности всех ресурсов, незамедлительно сообщаем об ошибках (при этом при возникновении ошибки можно откатить кластер в предыдущее состояние).
                </div>
            </div>
        </div>
    </div>
</div>

## Что такое converge?

**Converge** - это процесс сборки docker-образов (и их пересоздания в ответ на изменения), деплоя приложения в кластер Kubernetes (и ре-деплоя при необходимости) и контроля за работоспособностью приложения.

Команда `werf converge` запускает этот процесс. Ее может вызывать как пользователь, так и Ci/CD-система или оператор в ответ на изменения в состоянии приложения, описанном в Git. Поведение `werf converge` полностью детерминировано и прозрачно с точки зрения Git-репозитория. После завершения converge-процесса, приложение будет соответствовать состоянию, описанному в целевом Git-коммите. Для того, чтобы откатить приложение к предыдущей версии, обычно достаточно выполнить converge на соответствующем более раннем коммите (при этом werf будет использовать образы для этого коммита).

Выполните следующую команду в корневой директории своего проекта, чтобы запустить converge:

```
werf converge --docker-repo myregistry.domain.org/example-app [--kube-config ~/.kube/config]
```

Обычно у команды converge имеется только один обязательный аргумент: адрес docker-репозитория. werf будет использовать этот репозиторий для хранения собранных образов и их использования в Kubernetes (то есть репозиторий должен быть доступен из кластера Kubernetes). Kube-config - необязательный аргумент; определяет целевой кластер Kubernetes для подключения. Также имется опциональный параметр `--env` (и переменная окружения `WERF_ENV`), позволяющий развертывать приложение в различные [окружения]({{ site.baseurl }}/pages_ru/documentation/advanced/ci_cd/ci_cd_workflow_basics.html#окружение).

_Примечание: Если ваше приложение не использует кастомные docker-образы (а использует только публичные), параметр `--docker-repo` можно не указывать._

## Дальнейшие шаги

[Краткое руководство по началу работы]({{ site.baseurl }}/documentation/quickstart.html) поможет вам развернуть и запустить демо-приложение. [Руководства]({{ site.baseurl }}/pages_ru/documentation/guides.html) рассказывают о конфигурировании различных приложений, написанных на различных языках программирования и базирующихся на разных фреймворках. Здесь вы можете найти руководство, подходящее для вашего приложения, и воспользоваться приведенными в нем инструкциями.

Желающие получить более глубокое представление о рабочих процессах CI/CD, которые можно реализовать с помощью werf, могут обратиться к [этой статье]({{ site.baseurl }}/pages_ru/documentation/advanced/ci_cd/ci_cd_workflow_basics.html).
