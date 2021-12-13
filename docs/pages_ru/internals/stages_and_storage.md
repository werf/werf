---
title: Стадии и хранилище
permalink: internals/stages_and_storage.html
---

Мы разделили сборочный процесс образов, описанных в файле конфигурации [werf.yaml]({{ "reference/werf_yaml.html" | true_relative_url }}) на этапы, [с четкими функциями и назначением](#зависимости-стадии). Каждый такой этап соответствует промежуточному образу, подобно слоям в Docker. В werf такой этап называется [стадией](#конвеер-стадий), а **конечный образ** соответствует последней собранной стадии для определённого состояния git и конфигурации werf.yaml.

Стадии — это этапы сборочного процесса. ***Стадия*** определяется группой инструкций, указанных в конфигурации. Причем группировка этих инструкций не случайна, имеет определенную логику и учитывает условия и правила сборки. С каждой _стадией_ связан конкретный Docker-образ. Все стадии хранятся в [хранилище](#хранилище).

Вы можете рассматривать стадии как кэш сборки приложения, но в действительности это не совсем кэш, а часть сборочного контекста.

## Конвеер стадий

_Конвейер стадий_ — набор условий и правил выполнения стадий, подразумевающий также четко определенный порядок выполнения стадий. werf использует не один, а несколько _конвейеров стадий_ в своей работе, по-разному собирая образы в зависимости от их описанной конфигурации.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'dockerfile-image-tab')">Dockerfile-образ</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-image-tab')">Stapel-образ</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'stapel-artifact-tab')">Stapel-артефакт</a>
</div>

<div id="dockerfile-image-tab" class="tabs__content active">
<a class="google-drawings" href="{{ "images/reference/stages_and_images1.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images1_preview.png" | true_relative_url }}">
</a>
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

**Пользователю нужно только написать правильную конфигурацию, остальная работа со стадиями выполняется werf.**

Для каждой _стадии_, werf подсчитывает уникальный сборочный идентификатор — [дайджест стадии](#дайджест-стадии).
 
В случае отсутствия у стадии [зависимостей стадии](#зависимости-стадии), она пропускается, и, соответственно, _конвейер стадий_ уменьшается на одну стадию. Таким образом конвейер стадий может уменьшаться, вплоть до единственной стадии _from_.

<a class="google-drawings" href="{{ "images/reference/stages_and_images4.png" | true_relative_url }}" data-featherlight="image">
<img src="{{ "images/reference/stages_and_images4_preview.png" | true_relative_url }}">
</a>

## Дайджест стадии

_Дайджест стадии_ используется для [тегирования](#именование-стадий) _стадии_ (дайджест является только частью тега) в _хранилище_.
werf не собирает стадию, если стадия с таким же _дайджестом_ уже находится в _хранилище_ (это поведение похоже на кэширование в Docker, только имеет более сложную логику).

***Дайджест стадии*** — это контрольная сумма от:
 - контрольной суммы [зависимостей стадии](#зависимости-стадии).
 - дайджеста предыдущей стадии;
 - идентификатора git коммита связанного с предыдущей стадией (если эта стадия связана с git).

_Дайджест_ стадии идентифицирует содержимое стадии и зависит от истории правок в git, которые привели к этому коммиту.

## Зависимости стадии

_Зависимости стадии_ — это данные, которые напрямую связаны и влияют на [дайджест стадии](#дайджест-стадии). К зависимостям стадии относятся:

 - файлы (и их содержимое) из git-репозиториев;
 - инструкции сборки стадии из файла `werf.yaml`;
 - произвольные строки указанные пользователем в `werf.yaml`
 - и т.п.

Большинство _зависимостей стадии_ определяется в файле конфигурации `werf.yaml`, остальные — во время запуска.

Следующая таблица иллюстрирует зависимости в Dockerfile-образе, Stapel-образе и [Stapel-артефакте]({{ "advanced/building_images_with_stapel/artifacts.html" | true_relative_url }}).
Каждая строка таблицы описывает зависимости для определенной стадии. Левая колонка содержит краткое описание зависимостей, правая содержит соответствующую часть `werf.yaml` и ссылки на разделы с более подробной информацией.

<div class="tabs">
  <a href="javascript:void(0)" id="image-from-dockerfile-dependencies" class="tabs__btn dependencies-btn active">Dockerfile-образ</a>
  <a href="javascript:void(0)" id="image-dependencies" class="tabs__btn dependencies-btn">Stapel-образ</a>
  <a href="javascript:void(0)" id="artifact-dependencies" class="tabs__btn dependencies-btn">Stapel-артефакт</a>
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

## Хранилище

_Хранилище_ хранит стадии и метаданные проекта. Эти данные могут храниться локально на хост-машине, либо в Docker Repo.

Большинство команд werf используют _стадии_. Такие команды требуют указания места размещения _хранилища_ с помощью ключа `--repo` или переменной окружения `WERF_REPO`.

Существует 2 типа хранилища:
 1. _Локальное хранилище_. Использует локальный docker-server для хранения docker-образов.
 2. _Удалённое хранилище_. Использует container registry для хранения docker-образов. Включается опцией `--repo=CONTAINER_REGISTRY_REPO`, например, `--repo=registry.mycompany.com/web`. **ЗАМЕЧАНИЕ** Каждый проект должен использовать в качестве хранилища уникальный адрес репозитория, который используется только этим проектом.

Стадии будут [именоваться по-разному](#именование-стадий) в зависимости от типа используемого хранилища.

При использовании container registry для хранения стадий, локальный docker-server на всех хостах, где запускают werf, используется как кеш. Этот кеш может быть очищен автоматически самим werf-ом, либо удалён с помощью других инструментов (например `docker rmi`) без каких-либо последствий.

### Именование стадий

Стадии в _локальном хранилище_ именуются согласно следующей схемы: `PROJECT_NAME:STAGE_DIGEST-TIMESTAMP_MILLISEC`. Например:

```
myproject                   9f3a82975136d66d04ebcb9ce90b14428077099417b6c170e2ef2fef-1589786063772   274bd7e41dd9        16 seconds ago      65.4MB
myproject                   7a29ff1ba40e2f601d1f9ead88214d4429835c43a0efd440e052e068-1589786061907   e455d998a06e        18 seconds ago      65.4MB
myproject                   878f70c2034f41558e2e13f9d4e7d3c6127cdbee515812a44fef61b6-1589786056879   771f2c139561        23 seconds ago      65.4MB
myproject                   5e4cb0dcd255ac2963ec0905df3c8c8a9be64bbdfa57467aabeaeb91-1589786050923   699770c600e6        29 seconds ago      65.4MB
myproject                   14df0fe44a98f492b7b085055f6bc82ffc7a4fb55cd97d30331f0a93-1589786048987   54d5e60e052e        31 seconds ago      64.2MB
```

Стадии в _удалённом хранилище_ именуются согласно следующей схемы: `CONTAINER_REGISTRY_REPO:STAGE_DIGEST-TIMESTAMP_MILLISEC`. Например:

```
localhost:5000/myproject-stages                 d4bf3e71015d1e757a8481536eeabda98f51f1891d68b539cc50753a-1589714365467   7c834f0ff026        20 hours ago        66.7MB
localhost:5000/myproject-stages                 e6073b8f03231e122fa3b7d3294ff69a5060c332c4395e7d0b3231e3-1589714362300   2fc39536332d        20 hours ago        66.7MB
localhost:5000/myproject-stages                 20dcf519ff499da126ada17dbc1e09f98dd1d9aecb85a7fd917ccc96-1589714359522   f9815cec0867        20 hours ago        65.4MB
localhost:5000/myproject-stages                 1dbdae9cc1c9d5d8d3721e32be5ed5542199def38ff6e28270581cdc-1589714352200   6a37070d1b46        20 hours ago        65.4MB
localhost:5000/myproject-stages                 f88cb5a1c353a8aed65d7ad797859b39d357b49a802a671d881bd3b6-1589714347985   5295f82d8796        20 hours ago        65.4MB
localhost:5000/myproject-stages                 796e905d0cc975e718b3f8b3ea0199ea4d52668ecc12c4dbf85a136d-1589714344546   a02ec3540da5        20 hours ago        64.2MB
```

 - `PROJECT_NAME` — имя проекта;
 - `CONTAINER_REGISTRY_REPO` — репозиторий, заданный опцией `--repo`;
 - `STAGE_DIGEST` — дайджест стадии. Дайджест является идентификатором содержимого стадии и также зависит от истории правок в git репозитории, которые привели к такому содержимому.
 - `TIMESTAMP_MILLISEC` — уникальный идентификатор, который генерируется в процессе [процедуры сохранения стадии]({{ "internals/build_process.html#сохранение-стадий-в-хранилище" | true_relative_url }}) после того как стадия была собрана.
