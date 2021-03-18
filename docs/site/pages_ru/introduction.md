---
title: Введение
description: Как werf работает?
permalink: introduction.html
layout: plain
banner: guides
breadcrumbs: none
---

{% asset introduction.css %}
{% asset introduction.js %}
<div markdown="1">
## Что такое werf?

werf — это CLI-инструмент для организации полного цикла развертывания приложения с Git в качестве единого и универсального "источника истины". werf может:

 - Собирать docker-образы.
 - Деплоить приложение в кластер Kubernetes.
 - Убеждаться в том, что приложение запустилось и нормально работает после завершения развертывания.
 - Пересобирать docker-образы при внесении изменений в код приложения.
 - Ре-деплоить приложение в кластер Kubernetes при необходимости.
 - Удалять ненужные и неиспользуемые образы.

## Как werf работает?
</div>
<div id="introduction-presentation" class="introduction-presentation">
    <div id="introduction-presentation-controls" class="introduction-presentation__controls">
        <a href="javascript:void(0)" class="introduction-presentation__controls-nav">
            <img src="{% asset introduction/nav.svg @path %}" />
        </a>
        <div class="introduction-presentation__controls-stage">
            Определить желаемое состояние
        </div>
        <div class="introduction-presentation__controls-step">
            0. Конфигурация
        </div>
        <div class="introduction-presentation__controls-selector">
            <div class="introduction-presentation__controls-selector-stage">
                Определить желаемое состояние
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="0"
                    data-presentation-selector-stage="Определить желаемое состояние">
                    0. Конфигурация
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="1"
                    data-presentation-selector-stage="Определить желаемое состояние">
                    1. Копируем Dockerfiles в репозиторий приложения
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="2"
                    data-presentation-selector-stage="Определить желаемое состояние">
                    2. Создаем файл конфигурации werf.yaml
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="3"
                    data-presentation-selector-stage="Определить желаемое состояние">
                    3. Описываем шаблоны helm-чартов для развертывания приложения
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="4"
                    data-presentation-selector-stage="Определить желаемое состояние">
                    4. Делаем коммит
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-stage">
                Converge
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="5"
                    data-presentation-selector-stage="Converge">
                    1. Вычисляем сконфигурированные образы исходя из текущего состояния Git-коммита.
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="6"
                    data-presentation-selector-stage="Converge">
                    2. Считываем образы, доступные в container registry.
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="7"
                    data-presentation-selector-stage="Converge">
                    3. Вычисляем разницу между образами, соответствующими состоянию для текущего Git-коммита и теми, которые уже доступны в container registry.
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="8"
                    data-presentation-selector-stage="Converge">
                    4. Собираем и публикуем только те образы, которые отсутствуют в container registry (если такие имеются).
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="9"
                    data-presentation-selector-stage="Converge">
                    5. Считываем целевую конфигурацию ресурсов Kubernetes, связанную с текущим состоянием, определяемым Git-коммитом.
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="10"
                    data-presentation-selector-stage="Converge">
                    6. Считываем конфигурацию имеющихся в кластере Kubernetes-ресурсов.
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="11"
                    data-presentation-selector-stage="Converge">
                    7. Подсчитываем разницу между целевой для текущего состояния Git и имеющейся конфигурацией Kubernetes-ресурсов в кластере.
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="12"
                    data-presentation-selector-stage="Converge">
                    8. Применяем изменения (при необходимости) к ресурсам Kubernetes, чтобы они соответствовали состоянию, определенному в Git-коммите.
                </a>
            </div>
            <div class="introduction-presentation__controls-selector-step">
                <a href="javascript:void(0)"
                    data-presentation-selector-option="13"
                    data-presentation-selector-stage="Converge">
                    9. Убеждаемся в работоспособности всех ресурсов, незамедлительно сообщаем об ошибках (при этом при возникновении ошибки можно откатить кластер в предыдущее состояние).
                </a>
            </div>
        </div>
    </div>
    <div class="introduction-presentation__container">
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text">
                Конфигурация werf должна храниться в Git-репозитории приложения вместе с его кодом.
            </div>
            <img src="{% asset introduction/s-1.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-2.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text">
<div markdown="1">
Обратите особое внимание на параметр `project` — он содержит _название проекта_. В дальнейшем werf будет активно его использовать во время _converge-процесса_. Изменение этого параметра потом, когда приложение уже развернуто и работает, будет связано с простоем и потребует ре-деплоя приложения.
</div>
            </div>
            <img src="{% asset introduction/s-3.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text">
<div markdown="1">
Специальная шаблон-функция `werf_image` позволяет сгенерировать полное имя собираемого образа. У этой функции имеется параметр name, который соответствует образу, определенному в `werf.yaml` (`"frontend"` или `"backend"` в нашем примере).
</div>
            </div>
            <img src="{% asset introduction/s-4.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-5.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text">
<div markdown="1">
На этом шаге werf генерирует имена целевых образов. Имена могут менять или оставаться прежними после очередного коммита в зависимости от изменений в репозитории Git. Обратите внимание, что имена образов детерминированы и воспроизводимы и привязаны к соответствующему коммиту.
</div>
            </div>
            <img src="{% asset introduction/s-6.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-7.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-8.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-9.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-10.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-11.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text"></div>
            <img src="{% asset introduction/s-12.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title"></div>
            </div>
            <img src="{% asset introduction/s-13.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
        <div class="introduction-presentation__slide">
            <div class="introduction-presentation__slide-text">
                <div class="introduction-presentation__slide-title"></div>
            </div>
            <img src="{% asset introduction/s-14.svg @path %}"
            class="introduction-presentation__slide-img" />
        </div>
    </div>
</div>
<div markdown="1">
## Что такое converge?

**Converge** — это процесс сборки docker-образов (и их пересоздания в ответ на изменения), деплоя приложения в кластер Kubernetes (и ре-деплоя при необходимости) и контроля за работоспособностью приложения.

Команда `werf converge` запускает этот процесс. Ее может вызывать как пользователь, так и Ci/CD-система или оператор в ответ на изменения в состоянии приложения, описанном в Git. Поведение `werf converge` полностью детерминировано и прозрачно с точки зрения Git-репозитория (подробнее про гитерминизм можно прочитать [здесь]({{ "documentation/advanced/giterminism.html" | true_relative_url }})). После завершения converge-процесса, приложение будет соответствовать состоянию, описанному в целевом Git-коммите. Для того, чтобы откатить приложение к предыдущей версии, обычно достаточно выполнить converge на соответствующем более раннем коммите (при этом werf будет использовать образы для этого коммита).

Выполните следующую команду в корневой директории своего проекта, чтобы запустить converge:

```shell
werf converge --docker-repo myregistry.domain.org/example-app [--kube-config ~/.kube/config]
```

Обычно у команды converge имеется только один обязательный аргумент: адрес docker-репозитория. werf будет использовать этот репозиторий для хранения собранных образов и их использования в Kubernetes (то есть репозиторий должен быть доступен из кластера Kubernetes). Kube-config — необязательный аргумент; определяет целевой кластер Kubernetes для подключения. Также имется опциональный параметр `--env` (и переменная окружения `WERF_ENV`), позволяющий развертывать приложение в различные [окружения]({{ "documentation/advanced/ci_cd/ci_cd_workflow_basics.html#окружение" | true_relative_url }}).

_Примечание: Если ваше приложение не использует кастомные docker-образы (а использует только публичные), параметр `--docker-repo` можно не указывать._

## Дальнейшие шаги

[Краткое руководство по началу работы]({{ "documentation/quickstart.html" | true_relative_url }}) поможет вам развернуть и запустить демо-приложение. [Руководства]({{ "documentation/guides.html" | true_relative_url }}) рассказывают о конфигурировании различных приложений, написанных на различных языках программирования и базирующихся на разных фреймворках. Здесь вы можете найти руководство, подходящее для вашего приложения, и воспользоваться приведенными в нем инструкциями.

Желающие получить более глубокое представление о рабочих процессах CI/CD, которые можно реализовать с помощью werf, могут обратиться к [этой статье]({{ "documentation/advanced/ci_cd/ci_cd_workflow_basics.html" | true_relative_url }}).
</div>
