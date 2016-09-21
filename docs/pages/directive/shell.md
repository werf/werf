---
title: Shell директивы
sidebar: doc_sidebar
permalink: shell_directives.html
folder: directive
---

### shell.before_install, shell.before_setup, shell.install, shell.setup \<cmd\>\[, \<cmd\>, cache_version: \<cache_version\>\]

Директивы позволяют добавить bash-комманды для выполнения на соответствующих стадиях [shell приложения](definitions.html#shell-приложение).

* Опциональный параметр **\<cache_version\>** участвует в формировании сигнатуры стадии.

### shell.build_artifact \<cmd\>\[, \<cmd\>, cache_version: \<cache_version\>\]

Позволяет добавить bash-комманды для выполнения на соответствующей стадии [артефакта](definitions.html#артефакт).

* Опциональный параметр **\<cache_version\>** участвует в формировании сигнатуры стадии.

### shell.reset_before_install, shell.reset_before_setup, shell.reset_setup, shell.reset_install, shell.reset_build_artifact

Позволяет сбросить объявленные ранее bash-комманды соответсвующей стадии.

### shell.reset_all

Позволяет сбросить все объявленные ранее bash-комманды стадий.
