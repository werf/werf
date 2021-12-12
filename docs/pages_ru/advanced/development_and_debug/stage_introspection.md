---
title: Интроспекция стадий
permalink: advanced/development_and_debug/stage_introspection.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
summary: |
  <div class="language-bash highlighter-rouge">
  <div class="highlight"><pre class="highlight">
  <code><span class="c"># introspect a specific stage</span>
  werf build <span class="nt">--introspect-stage</span> <span class="o">[</span>IMAGE_NAME/]STAGE_NAME

  <span class="c"># introspect a stage before or after execution of a dysfunctional set of instructions</span>
  werf build <span class="nt">--introspect-error</span>
  werf build <span class="nt">--introspect-before-error</span></code>
  </pre></div>
  </div>
---


Написание конфигурации на начальном этапе может вызывать трудности из-за того, что при выполнении инструкций сборки какой-либо стадии не до конца понятно состояние системы в _сборочном контейнере_.

Благодаря опциям интроспекции вы можете получить доступ к конкретной _стадии_ непосредственно в процессе сборки. 
Во время интроспекции вы получаете такое же состояние контейнера, как и во время сборки, с теми же переменными окружения, с доступом к тем же служебным инструментам, используемым werf во время сборки. 
Эти служебные инструменты добавляются с помощью монтирования директорий из специального служебного контейнера — _stapel_ (доступен по адресу `/.werf/stapel` в _сборочном контейнере_). По сути, интроспекция, — это запуск _сборочного контейнера_ в интерактивном режиме для работы в нем пользователя.

Параметр `--introspect-stage` может быть указан несколько раз для интроспекции нескольких стадий. Формат использования:

* `IMAGE_NAME/STAGE_NAME` для интроспекции стадии `STAGE_NAME` **образа или артефакта** `IMAGE_NAME`. Безымянный образ можно указать как `~`.;
* `STAGE_NAME` или `*/STAGE_NAME` для интроспекции всех существующих стадий с именем `STAGE_NAME`.

**Во время разработки**, использование интроспекции позволяет сначала получить результат в сборочном контейнере, а затем перенести необходимые шаги и инструкции в конфигурацию соответствующей _стадии_. Такой подход удобен и позволяет быстрее достичь результата, когда вам понятно что должно быть в итоге, но сами шаги процесса не очевидны и требуют некоторых экспериментов и проверок.


<div class="videoWrapper">
<iframe width="560" height="315" src="https://www.youtube.com/embed/quoWwLSM_-4" frameborder="0" allow="encrypted-media" allowfullscreen></iframe>
</div>

**Во время отладки**, использование интроспекции позволяет быстрее понять, почему сборка завершилась с ошибкой, или полученный результат неожиданный. Вы можете также проверить содержимое необходимых файлов на какой-либо стадии сборки, либо проверить состояние системы.

<div class="videoWrapper">
<iframe width="560" height="315" src="https://www.youtube.com/embed/GiEbEhF2Pes" frameborder="0" allow="encrypted-media" allowfullscreen></iframe>
</div>

Наконец, при использовании интроспекции для приложений с **Ansible**, вы можете отлаживать Ansible-плейбуки в _сборочном контейнере_, а затем переносить их на необходимые _стадии_ конфигурации сборки.

<div class="videoWrapper">
<iframe width="560" height="315" src="https://www.youtube.com/embed/TEpn0yFvJik" frameborder="0" allow="encrypted-media" allowfullscreen></iframe>
</div>
