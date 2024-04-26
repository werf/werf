---
title: Миграция с v1.2 на v2.0
permalink: resources/migration_from_v1_2_to_v2_0.html
toc: false
---

## Обратно несовместимые изменения в v2.0

Ключевые изменения:
1. Новая подсистема развертывания Nelm включена по умолчанию. Старую подсистему развертывания более нельзя использовать.
1. Команды `werf converge`, `werf plan` и `werf bundle apply` имеют улучшенную валидацию ресурсов, что может потребовать исправлений ваших чартов.
1. Команды `werf render` и `werf bundle render` теперь форматируют результат, убирая комментарии, сортируя поля и форматируя значения.
1. Команды `werf render` и `werf bundle render` сортируют манифесты в результате в другом порядке.
1. Удалены команды `werf bundle download` и `werf bundle export`. Используйте `werf bundle copy --from REPO:TAG --to archive:mybundle.tar.gz`.
1. Переименована опция `--skip-build` в `--require-built-images`.
1. Заменена функция Helm-шаблонизатора `werf_image` на {% raw %}`{{ $.Values.werf.image.<MY_IMAGE_NAME> }}`{% endraw %}.
1. Заменены опции `--report-path`, `--report-format` на `--save-build-report`, `--build-report-path`.
1. В команде `werf bundle copy` заменены опции `--repo`, `--tag`, `--to-tag` на `--from=REPO`, `--from=REPO:TAG`, `--to=REPO:TAG`.
1. Удалена автоматическая миграция Helm 2 релизов на Helm 3 релизы.
    
Прочие изменения:
1. Заменены опции `--repo-implementation`, `--final-repo-implementation` на `--repo-container-registry`, `--final-repo-container-registry`.
1. Заменены Selectel Container Registry опции `--repo-selectel-account`, `--repo-selectel-password`, `--repo-selectel-username`, `--repo-selectel-vpc`, `--repo-selectel-vpc-id`, `--final-repo-selectel-account`, `--final-repo-selectel-password`, `--final-repo-selectel-username`, `--final-repo-selectel-vpc`, `--final-repo-selectel-vpc-id`. Используйте обычную аутентификацию в Container Registry.
1. Специальные аннотации werf вроде `werf.io/version` или `project.werf.io/name` больше не сохраняются в Helm-релизах (т. е. в Secret-ресурсах, по умолчанию), но по-прежнему применяются в кластер.
