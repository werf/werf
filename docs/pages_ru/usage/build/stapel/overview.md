---
title: Обзор
permalink: usage/build/stapel/overview.html
---

В werf встроен альтернативный синтаксис описания сборочных инструкций, называемый stapel, который даёт следующие возможности:

1. Удобство поддержки и параметризации комплексной конфигурации, возможность переиспользовать общие части и генерировать конфигурацию однотипных образов за счет использования YAML-формата и шаблонизации.
2. Специальные инструкции для интеграции с Git, позволяющие задействовать инкрементальную пересборку с учетом истории Git-репозитория.
3. Наследование образов и импортирование файлов из образов (аналог multi-stage для Dockerfile).
4. Запуск произвольных сборочных инструкций, опции монтирования директорий и другие инструменты продвинутого уровня для сборки образов.
5. Более эффективная механика кеширования слоёв (сейчас в pre-альфа аналогичная схема поддерживается и для слоёв Dockerfile при сборке с Buildah).

<!-- TODO(staged-dockerfile): удалить 5 пункт как неактуальный -->

Сборка образов через сборщик stapel предполагает описание сборочных инструкций в конфигурационном файле `werf.yaml`. stapel поддерживается как для сборочного бекенда Docker-сервера (сборка через shell-инструкции или ansible), так и для Buildah (только shell-инструкции).

В данном разделе рассмотрено, как описывать сборку образов с помощью сборщика stapel, а также описание дополнительных возможностей и как ими пользоваться.

<div class="details">
<a href="javascript:void(0)" class="details__summary">Как устроен конвеер стадий stapel</a>
<div class="details__content" markdown="1">

_Конвейер стадий_ — набор условий и правил выполнения стадий, подразумевающий также четко определенный порядок выполнения стадий.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-image-tab')">stapel-образ</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-artifact-tab')">stapel-артефакт</a>
</div>

<div id="stapel-image-tab" class="tabs__content">
<a class="google-drawings" href="{{ "images/reference/stages_and_images2.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images2_preview.png" | true_relative_url }}" >
</a>
</div>

<div id="stapel-artifact-tab" class="tabs__content">
<a class="google-drawings" href="{{ "images/reference/stages_and_images3.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images3_preview.png" | true_relative_url }}">
</a>
</div>

Для каждой _стадии_ werf подсчитывает уникальный сборочный идентификатор — дайджест стадии.

В случае отсутствия у стадии зависимостей она пропускается, и, соответственно, _конвейер стадий_ уменьшается на одну стадию.

<a class="google-drawings" href="{{ "images/reference/stages_and_images4.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images4_preview.png" | true_relative_url }}">
</a>

_Зависимости стадии_ — это данные, которые напрямую связаны и влияют на дайджест стадии. К зависимостям стадии относятся:
 - файлы (и их содержимое) из Git-репозиториев;
 - инструкции сборки стадии из файла `werf.yaml`;
 - произвольные строки, указанные пользователем в `werf.yaml`
 - и т.п.

Большинство _зависимостей стадии_ определяется в файле конфигурации `werf.yaml`, остальные — во время запуска.

Следующая таблица иллюстрирует зависимости в stapel-образе и stapel-артефакте. Каждая строка таблицы описывает зависимости для определенной стадии. Левая колонка содержит краткое описание зависимостей, правая содержит соответствующую часть `werf.yaml` и ссылки на разделы с более подробной информацией.

<div class="tabs">
  <a href="javascript:void(0)" id="image-dependencies" class="tabs__btn dependencies-btn active">stapel-образ</a>
  <a href="javascript:void(0)" id="artifact-dependencies" class="tabs__btn dependencies-btn">stapel-артефакт</a>
</div>

<div id="dependencies">
{% for stage in site.data.stages.ru.entries %}
<div class="stage {{stage.type}}">
  <div class="stage-body">
    <div class="stage-base">
      <p>stage {{ stage.name | escape }}</p>

      {% if stage.dependencies %}
      <div class="dependencies">
        {% for dependency in stage.dependencies %}
        <div class="dependency">
          {{ dependency | escape }}
        </div>
        {% endfor %}
      </div>
      {% endif %}
    </div>

<div class="werf-config" markdown="1">

{% if stage.werf_config %}
```yaml
{{ stage.werf_config }}
```
{% endif %}

{% if stage.references %}
<div class="references">
    Подробнее:
    <ul>
    {% for reference in stage.references %}
        <li><a href="{{ reference.link | true_relative_url }}">{{ reference.name }}</a></li>
    {% endfor %}
    </ul>
</div>
{% endif %}

</div>

    </div>
</div>
{% endfor %}
</div>

<link rel="stylesheet" type="text/css" href="{{ assets["stages.css"].digest_path | true_relative_url }}" />

<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.4.1/jquery.min.js"></script>
<script>
function application() {
  if ($("a[id=image-dependencies]").hasClass('active')) {
    $(".artifact").addClass('hidden');
    $(".image").removeClass('hidden')
  }
  else if ($("a[id=artifact-dependencies]").hasClass('active')) {
    $(".image").addClass('hidden');
    $(".artifact").removeClass('hidden')
  }
  else {
    $(".image").addClass('hidden');
    $(".artifact").addClass('hidden')
  }
}

$('.tabs').on('click', '.dependencies-btn', function() {
  $(this).toggleClass('active').siblings().removeClass('active');
  application()
});

application();
$.noConflict();
</script>

</div>
</div>
