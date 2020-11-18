---
title: Стадии и образы
sidebar: documentation
permalink: documentation/reference/stages_and_images.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Мы предлагаем разделить сборочный процесс на этапы, каждый с четкими функциями и своим назначением. Каждый такой этап соответствует промежуточному образу, подобно слоям в Docker. В werf такой этап называется [стадией](#стадии), и конечный [образ](#образы) в итоге состоит из набора собранных стадий. Все стадии хранятся в [хранилище стадий](#хранилище-стадий), которое можно рассматривать как кэш сборки приложения, хотя по сути это скорее часть контекста сборки.

## Стадии

Стадии — это этапы сборочного процесса, кирпичи, из которых в итоге собирается конечный образ.
***Стадия*** собирается из группы сборочных инструкций, указанных в конфигурации. Причем группировка этих инструкций не случайна, имеет определенную логику и учитывает условия и правила сборки. С каждой _стадией_ связан конкретный Docker-образ.

Сборочный процесс werf подразумевает последовательную сборку стадий с использованием _конвейера стадий_. _Конвейер стадий_ — набор условий и правил выполнения стадий, подразумевающий также четко определенный порядок выполнения стадий. werf использует не один, а несколько _конвейеров стадий_ в своей работе, по-разному собирая образы в зависимости от их описанной конфигурации.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'dockerfile-image-tab')">Dockerfile-образ</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-image-tab')">Stapel-образ</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-artifact-tab')">Stapel-артефакт</a>
</div>

<div id="dockerfile-image-tab" class="tabs__content active">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRrzxht-PmC-4NKq95DtLS9E7JrvtuHy0JpMKdylzlZtEZ5m7bJwEMJ6rXTLevFosWZXmi9t3rDVaPB/pub?w=2031&amp;h=144" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRrzxht-PmC-4NKq95DtLS9E7JrvtuHy0JpMKdylzlZtEZ5m7bJwEMJ6rXTLevFosWZXmi9t3rDVaPB/pub?w=821&amp;h=59">
</a>
</div>

<div id="stapel-image-tab" class="tabs__content">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=2000&amp;h=881" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=821&amp;h=362" >
</a>
</div>

<div id="stapel-artifact-tab" class="tabs__content">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=1800&amp;h=850" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=640&amp;h=301">
</a>
</div>

**Пользователю нужно только написать правильную конфигурацию, остальная работа со стадиями выполняется werf.**

При каждой сборке, для каждой _стадии_, werf подсчитывает уникальный сборочный идентификатор — _сигнатуру стадии_.
Сборка каждой _стадии_ выполняется в ***сборочном контейнере***, который основан на предыдущей стадии согласно _конвейеру стадий_. После завершения работы сборочного контейнера, стадия сохраняется в [хранилище стадий](#хранилище-стадий).

_Сигнатура стадии_ используется для [тегирования](#именование-стадий) _стадии_ в _хранилище стадий_.
werf не собирает стадию, если стадия с такой _сигнатурой_ уже находится в _хранилище стадий_ (это поведение похоже на кэширование в Docker, только имеет более сложную логику).

***Сигнатура стадии*** — это контрольная сумма от:
- сигнатуры предыдущей стадии;
- контрольной суммы [зависимостей стадии]({{ site.baseurl }}/documentation/reference/stages_and_images.html#зависимости-стадии).

В случае отсутствия у стадии _зависимостей стадии_, она пропускается, и, соответственно, _конвейер стадий_ уменьшается на одну стадию. Таким образом конвейер стадий может уменьшаться на несколько стадий, вплоть до единственной стадии _from_.

<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=572&amp;h=577" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=286&amp;h=288">
</a>

## Зависимости стадии

_Зависимости стадии_ — это данные, которые напрямую связаны и влияют на _сигнатуру стадии_. К зависимостям стадии относятся:

 - файлы (и их содержимое) из git-репозиториев;
 - инструкции сборки стадии из файла `werf.yaml`;
 - произвольные строки указанные пользователем в `werf.yaml`
 - и т.п.

Большинство _зависимостей стадии_ определяется в файле конфигурации `werf.yaml`, остальные — во время запуска.

Следующая таблица иллюстрирует зависимости в Dockerfile-образе, Stapel-образе и [Stapel-артефакте]({{ site.baseurl }}/documentation/configuration/stapel_artifact.html).
Каждая строка таблицы описывает зависимости для определенной стадии. Левая колонка содержит краткое описание зависимостей, правая содержит соответствующую часть `werf.yaml` и ссылки на разделы с более подробной информацией.

<div class="tabs">
  <a href="javascript:void(0)" id="image-from-dockerfile-dependencies" class="tabs__btn dependencies-btn">Dockerfile-образ</a>
  <a href="javascript:void(0)" id="image-dependencies" class="tabs__btn dependencies-btn">Stapel-образ</a>
  <a href="javascript:void(0)" id="artifact-dependencies" class="tabs__btn dependencies-btn">Stapel-артефакт</a>
</div>

<div id="dependencies">
{% for stage in site.data.stages.entries %}
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
    References:
    <ul>
    {% for reference in stage.references %}
        {% if reference.nonexistentRuAnchor %}
        <li><a data-proofer-ignore href="{{ reference.link | relative_url }}">{{ reference.name }}</a></li>
        {% else %}
        <li><a href="{{ reference.link | relative_url }}">{{ reference.name }}</a></li>
        {% endif %}
    {% endfor %}
    </ul>
</div>
{% endif %}

</div>

    </div>
</div>
{% endfor %}
</div>

{% asset stages.css %}
<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.4.1/jquery.min.js"></script>
<script>
function application() {
  if ($("a[id=image-from-dockerfile-dependencies]").hasClass('active')) {
    $(".image").addClass('hidden');
    $(".artifact").addClass('hidden');
    $(".image-from-dockerfile").removeClass('hidden')
  }
  else if ($("a[id=image-dependencies]").hasClass('active')) {
    $(".image-from-dockerfile").addClass('hidden');
    $(".artifact").addClass('hidden');
    $(".image").removeClass('hidden')
  }
  else if ($("a[id=artifact-dependencies]").hasClass('active')) {
    $(".image-from-dockerfile").addClass('hidden');
    $(".image").addClass('hidden');
    $(".artifact").removeClass('hidden')
  }
  else {
    $(".image-from-dockerfile").addClass('hidden');
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

## Хранилище стадий

_Хранилище стадий_ содержит стадии проекта. Стадии могут храниться локально на хост-машине, либо в Docker registry.

Большинство команд werf используют _стадии_. Такие команды требуют указания места размещения _хранилища стадий_ с помощью ключа `--stages-storage` или переменной окружения option or `WERF_STAGES_STORAGE`. На текущий момент поддерживается только локальное размещение _хранилища стадий_ — ключ `:local`.

### Именование стадий

_Стадии_ в _хранилище стадий_ именуются согласно следующей схемы: `werf-stages-storage/PROJECT_NAME:STAGE_SIGNATURE`
,где:
- `PROJECT_NAME` — имя проекта
- `STAGE_SIGNATURE` — сигнатура стадии

## Образы

_Образ_ — это **готовый к использованию** Docker-образ, относящийся к опеределенному состоянию приложения в соответствии со [стратегией тегирования]({{ site.baseurl }}/documentation/reference/publish_process.html).

Как было написано [выше](#стадии), _стадии_ — это этапы сборочного процесса, кирпичи, из которых в итоге собирается конечный _образ_.
_Стадии_, в отличие от конечных _образов_ не предназначены для прямого использования. Основное отличие между конечными образами и стадиями — разное поведение [политики очистки]({{ site.baseurl }}/documentation/reference/cleaning_process.html#очистка-по-политикам) по отношению к ним, я также различия в структуре хранимой мета-информации.
Очистка _хранилища стадий_ основана только на наличии в _репозитории образов_ связанных с соответствующими стадиями образов.

werf создает _образы_ используя _хранилище стадий_.
На текущий момент, _образы_ создаются только в [_процессе публикации_]({{ site.baseurl }}/documentation/reference/publish_process.html) (publish) и хранятся в _репозитории образов_.

Конфигурация образов должна быть описана в файле конфигурации `werf.yaml`.

Читай [подробнее]({{ site.baseurl }}/documentation/reference/build_process.html) о процессе сборки.
