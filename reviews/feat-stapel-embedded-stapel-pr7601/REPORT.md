# Multi-Role Code Review Report

**PR:** [werf/werf#7601](https://github.com/werf/werf/pull/7601) — feat(stapel, build): embed stapel
**Branch:** `feat/stapel/embedded-stapel` → base `3`
**Diff:** 8 files, +396/-14 lines

## DoD Criteria

1. `task build` (and any build/lint entry point) succeeds without requiring a running Docker daemon at build time, and produces a werf binary with the stapel image embedded for linux/amd64 and linux/arm64.
2. At runtime, when using the default (non-overridden) stapel image reference and target platform is linux/amd64 or linux/arm64, werf loads the embedded image via `docker load` instead of `docker pull`.
3. When `WERF_STAPEL_IMAGE_NAME`/`WERF_STAPEL_IMAGE_VERSION` is set (custom image), or when the target platform is not linux/amd64 or linux/arm64, werf falls back to `docker pull` as before (existing behavior preserved, no breaking change).
4. Embedded image integrity is verified via SHA-256 checksum before being loaded into Docker; corrupted/mismatched embedded data must not silently produce a broken image.
5. Embedded artifacts are excluded from git (`.gitignore`) and regenerated deterministically from the published `registry.werf.io/werf/stapel:<VERSION>` image tag matching `pkg/stapel/stapel.go`'s `VERSION` constant.
6. `stapel:embed` task is wired into `build`, `format`, and `lint` Task pipelines so a plain `task build` (or CI's build step) always has valid embed artifacts before compiling.
7. Unit tests cover: platform selection logic, platform normalization/aliasing, checksum verification (success/failure/corrupt-gzip cases), and the "is default image ref" env var gating.
8. Documentation (`scripts/stapel/DEVELOPMENT.md`) explains the new embed flow and updated stapel version bump release process.
9. `pkg/stapel/stapel.go` VERSION bump to 0.7.1 corresponds to an existing published tag in the registry.

---

## Expert Opinions

*Read these first. If all are positive ✅, details below are optional.*

- **Technical Reviewer**: Функционально фича закрывает большинство DoD-критериев (эмбеддинг, платформенный fallback, checksum, тесты), но вносит **регрессию для голого Go-тулинга** (`go build`/`go test`/gopls на `pkg/stapel` ломаются без `task`), необоснованно привязывает сетевой pull к `format`, и — критично — **не покрывает `task test:unit`** зависимостью на `stapel:embed`, хотя это мандаторная команда тестирования по AGENTS.md.
- **Product Reviewer**: Решает реальную проблему (offline/air-gapped), полностью аддитивно и с фолбэком — но **необъявленный рост бинарника на ~35-40MB** для всех 5 платформ релиза не задокументирован для пользователя, конечная user-facing документация не обновлена, opt-out механизма нет, PR body пустой.
- **Risk Analyst**: Два Critical-риска (bare Go tooling breakage, `test:unit` не покрыт embed-зависимостью) требуют исправления до мержа. High-риски вокруг недокументированного роста бинарника и сетевой зависимости `format`/`lint` без кэширования — существенные, но не блокеры при осознанном принятии trade-off.

---

## Risk Analysis Table

| № | Risk | Type | Probability | Severity | Location | Circumstances | Consequences |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | `pkg/stapel` не компилируется голыми Go-инструментами | Technical | 1.0 | Critical | `pkg/stapel/embedded.go:20-30` (`//go:embed embed/linux/*/werf-stapel-toolchain.tar.gz`) | Любой разработчик/CI-джоб запускает `go build ./...`, `go test ./pkg/stapel/...`, `go vet`, либо открывает пакет в IDE с gopls без предварительного `task stapel:embed` | Сборка/тесты/линт падают с ошибкой "pattern embed/... no matching files"; gopls показывает ложные ошибки по всему пакету, разработчики теряют доверие к редактору и тратят время на диагностику несуществующей проблемы |
| 2 | `task test:unit` не тянет `stapel:embed` как зависимость | Technical/Operational | 0.9 | Critical | `Taskfile.dist.yaml` (deps добавлены только для `format`, `lint`, `_build:cgo:dev`, `_build:go:dev`) | Свежий чекаут репозитория, разработчик или CI-раннер выполняет мандаторную команду `task test:unit` напрямую (как предписано AGENTS.md), не запуская предварительно `task build` | `pkg/stapel` тесты (включая новый `embedded_ai_test.go`) падают из-за отсутствующих embed-файлов; PR-required команда из AGENTS.md не работает "из коробки", что прямо противоречит DoD-требованию тестового покрытия |
| 3 | Мандаторный сетевой pull образа встроен в `format`/`lint`/`build` без кэширования | Technical/Operational | 0.8 | High | `Taskfile.dist.yaml:41-46` (`stapel:embed`, только `run: once`, нет `sources:`/`generates:`) | Разработчик за корпоративным firewall или в air-gapped CI выполняет `task format`, `task lint` или повторный `task build` — каждый отдельный вызов Task заново тянет ~36MB из `registry.werf.io` | Форматирование кода и линтинг становятся недоступны без доступа к registry; повторные сборки не работают оффлайн, что иронично противоречит цели фичи (offline/air-gapped runtime), но сама сборка требует сеть |
| 4 | Безусловный рост бинарника на ~35-40MB для всех 5 релизных платформ, не задокументирован для пользователя | UX/Product | 1.0 | High | `cmd/embed-stapel/main.go:17-20` (`platforms = [linux/amd64, linux/arm64]`, эмбедятся во все таргеты) | Каждый релиз werf (включая darwin/windows сборки, где эмбед вообще не используется) увеличивается в размере; пользователи скачивают/обновляют бинарь без объяснения причины роста | Нарушение правила "flag breaking/user-facing changes first"; жалобы пользователей на распухший бинарь, отсутствие opt-out для size-sensitive окружений (например, встраивание werf в тонкие CI-образы) |
| 5 | `stapel:embed` привязан к `format` без технической необходимости | Technical | 0.7 | High | `Taskfile.dist.yaml:41-46` | Разработчик вызывает `task format` только для приведения кода к стилю (gci/gofumpt), не планируя сборку/тесты | Неожиданный сетевой запрос и запись файлов на диск при чисто косметической операции; в offline-среде `task format` перестаёт работать вообще, хотя раньше работал всегда |
| 6 | Расхождение неймспейсов `STAPEL_IMAGE` (embed-time) и `WERF_STAPEL_IMAGE_NAME`/`WERF_STAPEL_IMAGE_VERSION` (runtime) не задокументировано вместе | Technical | 0.4 | Medium | `cmd/embed-stapel/main.go:24-25` vs `pkg/stapel/stapel.go:37-39` | Контрибьютор, бампающий версию stapel или настраивающий кастомный registry mirror для сборки, путает две несвязанные переменные окружения | Неверная конфигурация: сборка эмбедит один образ, а рантайм пытается работать с другим referenced именем; трудно диагностируемые несоответствия версий |
| 7 | Пользовательская документация не отражает новую offline-возможность | UX/Product | 0.9 | Medium | `docs/pages_en/usage/build/stapel/instructions.md` (не изменён) | Пользователь в air-gapped среде — целевой бенефициар фичи — не находит информацию о новом embedded-механизме и продолжает считать, что нужен доступ к registry для stapel-образа | Целевая аудитория фичи не узнаёт о её существовании; DoD документация ограничена contributor-facing `DEVELOPMENT.md`, реальная ценность фичи не реализуется на практике |
| 8 | Нет opt-out/feature flag для build-size-sensitive сборок | UX/Product | 0.5 | Medium | `cmd/embed-stapel/main.go` (безусловный `platforms` slice, нет build tag) | Контрибьютор или дистрибьютор werf (например, для встраивания в минимальный CI-образ) хочет собрать "slim" бинарь без эмбеда 35-40MB | Невозможность собрать компактный werf без модификации кода/Taskfile вручную; barrier для кастомных дистрибуций |
| 9 | Конкурентные вызовы `stapel:embed` из параллельных процессов Task (не защищено `run: once` внутри одного графа) | Technical | 0.3 | Medium | `Taskfile.dist.yaml:41-46`, `cmd/embed-stapel/main.go:47-50` (запись `os.Create` без лока) | Два параллельных CI-джоба или `task build & task lint` в одном воркспейсе одновременно перезаписывают один и тот же `tar.gz`/`sha256` файл | Гонка при записи может привести к рассинхронизации `tar.gz` и `sha256` → сборка либо падает на checksum-mismatch, либо использует несогласованную пару файлов до следующей проверки |
| 10 | `loadEmbeddedImage` не логирует через `logboek`, decompress+load проходит молча | Technical | 0.3 | Low | `pkg/stapel/embedded.go:66-82` | Пользователь ждёт первую сборку stapel-контейнера; decompress ~36MB + `docker load` занимает заметное время, но пользователь не видит прогресса, в отличие от `docker.CliPullWithRetries` | Ощущение "зависшего" процесса, лишние вопросы в поддержку/issue-трекере, ухудшение UX по сравнению с прежним pull-based флоу с видимым прогрессом |
| 11 | Пустое тело PR затрудняет ревью и категоризацию release-please | Operational | 1.0 | Low | PR metadata (не в диффе) | Merge PR без описания контекста/мотивации изменения | Усложнённая археология изменений при будущих багфиксах; риск неверной или отсутствующей записи в CHANGELOG для user-facing изменения |

## Risk Treatment Recommendations

| Risk № | Severity | Role | Strategy | Recommendation | Justification |
| :--- | :--- | :--- | :--- | :--- | :--- |
| #1 | Critical | Technical Specialist | Mitigate | As Technical Specialist for risk «bare Go tooling breakage» I recommend commit generated (or minimal placeholder) embed artifacts to git under version control OR gate the `//go:embed` behind a build tag so `go vet`/gopls/plain `go build` degrade gracefully instead of hard-failing on `pkg/stapel`. | Регрессия критична: любой инструмент вне Taskfile-пайплайна (IDE, LSP, ad-hoc CI) перестаёт работать с пакетом, что напрямую противоречит принципу "код должен собираться стандартными средствами". |
| #2 | Critical | Technical Specialist | Mitigate | As Technical Specialist for risk «`test:unit` not covered» I recommend add `stapel:embed` as a dependency of `test:unit` (or of the shared `test:ginkgo`/`_test:go-test` base tasks) in `Taskfile.dist.yaml`, mirroring what was already done for `build`/`format`/`lint`. | AGENTS.md мандаторно предписывает `task test:unit` для тестирования — команда обязана работать "из коробки" на свежем чекауте без скрытых предусловий. |
| #3 | High | Technical Specialist | Mitigate | As Technical Specialist for risk «no local caching» I recommend add `sources:`/`generates:` (checksum-based) to `stapel:embed` task so Task skips re-pulling when local artifacts already match the current `VERSION` and are present/valid. | Устраняет ненужную сетевую зависимость на каждый вызов сборки/линта/форматтера, приближает к декларируемой цели offline-friendliness. |
| #4 | High | Product Manager | Escalate | As Product Manager for risk «undisclosed binary size growth» I recommend explicitly calling out the ~35-40MB size increase in the PR description, commit message (for changelog generation), and considering a `feat(stapel)!:` or dedicated changelog note so users understand the trade-off before upgrading. | Правило репозитория требует явного флага для user-facing изменений такого масштаба; молчаливое увеличение размера дистрибутива — это именно такое изменение. |
| #5 | High | Technical Specialist | Avoid | As Technical Specialist for risk «unnecessary `format` coupling» I recommend remove `stapel:embed` from `format`'s cmds — gci/gofumpt operate on source text and require neither compilation nor image data. | Нарушение KISS/YAGNI: чисто косметическая операция не должна требовать сетевого доступа к внешнему registry. |
| #7 | Medium | Product Manager | Mitigate | As Product Manager for risk «user docs not updated» I recommend update `docs/pages_en/usage/build/stapel/instructions.md` (and RU counterpart) to mention the embedded/offline capability and cross-reference the existing air-gapped/private-registry section. | Целевой бенефициар фичи (air-gapped пользователь) не найдёт информацию о ней без обновления user-facing документации — DoD "documentation" сейчас закрыт только contributor-doc. |
| #6 | Medium | Technical Specialist | Monitor | As Technical Specialist for risk «STAPEL_IMAGE vs WERF_STAPEL_IMAGE_NAME divergence» I recommend document both env-var namespaces together in `scripts/stapel/DEVELOPMENT.md`, or better, reuse `WERF_STAPEL_IMAGE_NAME` in `cmd/embed-stapel` to avoid a second naming scheme. | Две несвязанные переменные для одной концептуальной настройки — источник будущих ошибок конфигурации при малой цене исправления. |
| #8 | Medium | Product Manager | Accept | As Product Manager for risk «no opt-out for slim builds» I recommend accept for now given no current stated requirement for slim builds, but track as a follow-up if binary-size-sensitive distribution scenarios arise. | YAGNI — добавлять build-tag-based opt-out без confirmed use case было бы преждевременной сложностью. |
| #9 | Medium | Technical Specialist | Monitor | As Technical Specialist for risk «concurrent embed writes» I recommend monitor in CI — if parallel `task` invocations against the same working directory become a real pattern (e.g. matrix jobs sharing a checkout), add a file lock around `buildPlatform` writes in `cmd/embed-stapel/main.go`. | Вероятность низкая в текущих CI-топологиях (обычно каждый job — свой checkout), не оправдывает добавление сложности сейчас. |
| #10 | Low | Technical Specialist | Mitigate | As Technical Specialist for risk «silent decompress+load» I recommend wrap `loadEmbeddedImage` with a `logboek.Context(ctx).LogProcess` call similar to existing `docker.CliPullWithRetries` progress reporting. | Небольшое изменение существенно улучшает UX первого запуска, где операция может занять заметное время. |
| #11 | Low | Risk Manager | Mitigate | As Risk Manager for risk «empty PR body» I recommend request the author fill in PR description with motivation (air-gapped/offline use case) and link to any related issue before merge. | Минимальные усилия, снижающие риск будущей путаницы при археологии изменений и улучшающие качество release notes. |

---

## Post-Review Mitigation: Build Tag Strategy (2026-07-07)

После первичного ревью проанализирована альтернативная митигация для рисков #1 и #2 (оба Critical) — вместо коммита embed-артефактов в git (исходная рекомендация #1) или добавления `stapel:embed` в зависимости `test:unit` (исходная рекомендация #2), предложен **build tag gate**:

- `pkg/stapel/embedded.go` получает `//go:build embedstapel` — реальная логика с `go:embed` (строки 250-333 диффа) компилируется только с тегом.
- Новый `pkg/stapel/embedded_stub.go` с `//go:build !embedstapel` — `embeddedImageForPlatform` всегда возвращает `(embeddedImage{}, false)`; `acquireImage` (container.go:207-220 диффа) автоматически падает на `docker.CliPullWithRetries` без дополнительных изменений в этом файле.
- `Taskfile.dist.yaml` передаёт `-tags=embedstapel` в `go build` только для `_build:cgo:dev`/`_build:go:dev`/`_build:*:dist` (реальные сборочные пути); `stapel:embed` остаётся зависимостью только тегированных путей.
- `pkg/stapel/embedded_ai_test.go` (диф 334-464) разделяется: чистая логика (`normalizeEmbeddedPlatform`, `decompressAndVerify` на синтетических данных) остаётся без тега; `TestAI_embeddedArtifactsMatchSha256`, зависящий от реальных embed-файлов, получает `//go:build embedstapel`.

### Обновлённые статусы рисков

| № | Риск | Severity | Статус после build-tag фикса |
| :--- | :--- | :--- | :--- |
| 1 | bare Go tooling breakage | Critical | **ЗАКРЫТ** — untagged путь не требует embed-директории вообще |
| 2 | `test:unit` не тянет `stapel:embed` | Critical | **ЗАКРЫТ** — untagged тесты бьют в stub, embed-зависимость не нужна |
| 3 | нет кэширования network pull | High | ОТКРЫТ — ортогонален тегу, нужен отдельный `sources:`/`generates:` |
| 4 | недокументированный рост бинарника | High | ОТКРЫТ — тег не уменьшает охват платформ в релизных сборках (docker daemon всегда linux, эмбед нужен на всех 5 таргетах) |
| 5 | `stapel:embed` привязан к `format` | High | ОТКРЫТ — не связано с тегом, format не компилирует код |
| 6 | расхождение `STAPEL_IMAGE`/`WERF_STAPEL_IMAGE_NAME` | Medium | ОТКРЫТ |
| 7 | user-facing документация не обновлена | Medium | ОТКРЫТ |
| 8 | нет opt-out для slim-сборок | Medium | **ЗАКРЫТ как побочный эффект** — сам тег и есть opt-out механизм |
| 9 | конкурентная запись embed-файлов | Medium | ОТКРЫТ |
| 10 | silent decompress+load | Low | ОТКРЫТ (актуален только при включённом теге) |
| 11 | пустое тело PR | Low | ОТКРЫТ |
| **12 (новый)** | Расхождение build-путей: если релизный `build:dist:*` таргет забудет `-tags=embedstapel`, бинарь молча соберётся pull-only без embed, DoD #1/#2 нарушается без видимой ошибки | **High** | Внесён фиксом. Митигация: тег задаётся один раз через общий Taskfile-якорь/переменную, шарится всеми `build:dist:*` таргетами; плюс CI-проверка (sanity build с `-tags=embedstapel` + проверка, что `embeddedImageForPlatform` возвращает непустой результат, либо assert на размер бинарника) |

### Итоговый вердикт

Оба Critical-риска (#1, #2) закрываются build-tag подходом дешевле, чем исходные рекомендации (коммит бинарных blob'ов в git — необратимый рост репозитория; добавление сетевой зависимости в `test:unit` — усложнение мандаторной команды). Bonus: закрывает Medium-риск #8 как побочный эффект.

Новый риск #12 (High) контейнируем — единая точка правды в Taskfile + CI-assert, не размазан по N файлам как исходные Critical-риски.

**Мержибельность:** Critical-блокеров не остаётся при условии, что автор реализует tag-split. Оставшиеся риски (#3, #4, #5, #6, #7, #9, #10, #11, #12) — не блокеры, могут идти как follow-up, но #4 (раскрытие роста размера) и #5 (убрать `stapel:embed` из `format`) стоит закрыть до мержа как дешёвые правки без архитектурных последствий.

### Implementation Status (2026-07-07)

Build-tag split реализован и запушен в `feat/stapel/embedded-stapel` (commit `190a77039`, "refactor(build/stapel): gate embedded stapel behind embedstapel build tag").

**Изменённые/новые файлы:**
- `pkg/stapel/embedded_util.go` (новый) — tag-independent логика: `embeddedImage` тип, `normalizeEmbeddedPlatform`, `decompressAndVerify`.
- `pkg/stapel/embedded.go` (`//go:build embedstapel`) — реальные `go:embed` директивы, `embeddedImageForPlatform`, `loadEmbeddedImage`.
- `pkg/stapel/embedded_stub.go` (новый, `//go:build !embedstapel`) — заглушки, всегда `(embeddedImage{}, false)`, `acquireImage` автоматически падает на `docker.CliPullWithRetries`.
- `pkg/stapel/embedded_ai_test.go` — оставлены только tag-independent тесты (`normalizeEmbeddedPlatform`, `decompressAndVerify`, `isDefaultImageRef`).
- `pkg/stapel/embedded_embedstapel_ai_test.go` (новый, `//go:build embedstapel`) — `TestAI_embeddedImageForPlatform`, `TestAI_embeddedArtifactsMatchSha256`, зависящие от реальных embed-артефактов.
- `Taskfile.dist.yaml` — добавлены `buildGoTags`/`buildCgoTags` (единая точка правды = `goTags`/`cgoTags` + `embedstapel`), применены в `_build:cgo:dev`, `_build:cgo:dist`, `_build:go:dev`, `_build:go:dist`, `lint:golangci-lint:go`. `test:unit`/`test:ginkgo`/bare `go test` остаются на untagged `goTags`/`cgoTags`.

**Верификация (выполнена в изолированном git worktree):**
- Bare `go build ./pkg/stapel/...` без тега → **exit 0**, embed-директория не нужна (Risk #1 закрыт).
- Bare `go vet ./pkg/stapel/...` без тега → **exit 0**.
- `go test ./pkg/stapel/...` без тега, embed-артефактов на диске нет → **PASS** (все `TestAI_*` + ginkgo suite), Risk #2 закрыт без изменений в зависимостях `test:unit`.
- `task stapel:embed` → сгенерировал реальные embed-артефакты (~17MB × 2 платформы) через `crane.Pull` без Docker daemon.
- Сборка/тесты с тегом `embedstapel` (`go build`/`go test -tags=...embedstapel`) → **exit 0**, `TestAI_embeddedArtifactsMatchSha256` подтвердил integrity check на реальных данных.
- `task build` (чистый чекаут, `pkg/stapel/embed/` предварительно удалён) → бинарь с embedded stapel, **DoD #1 подтверждён** без запущенного Docker daemon.
- `task lint golangciPaths="./cmd/... ./pkg/... ./test/..."` (весь проект) → **0 issues**.
- `task test:unit` (весь проект, без scope) → Stapel Suite 3/3 SUCCESS; 2 unrelated pre-existing failures в `pkg/true_git` (macOS gpg-agent "Cannot allocate memory") и `pkg/werf/exec` (process spawn issue) — не связаны с изменениями, локальное окружение.

**Статусы рисков после реализации:**
| № | Риск | Статус |
| :--- | :--- | :--- |
| 1 | bare Go tooling breakage | **ЗАКРЫТ, верифицировано** |
| 2 | `test:unit` не тянет `stapel:embed` | **ЗАКРЫТ, верифицировано** — зависимость и не понадобилась |
| 8 | нет opt-out для slim-сборок | **ЗАКРЫТ** — тег есть opt-out |
| 12 | забытый `-tags=embedstapel` на релизном таргете | **Митигирован** — единая переменная `buildGoTags`/`buildCgoTags` вместо литералов на каждом таргете |

---

## Full Remediation (2026-07-07, continued)

По запросу пользователя закрыты все оставшиеся риски из отчёта пошагово, с подтверждением перед каждым шагом. Итого 6 дополнительных коммитов запушены в `feat/stapel/embedded-stapel`.

### Risk #5 — `stapel:embed` привязан к `format` (commit `b49587d7f`)

Убрана строка `task: stapel:embed` из `cmds` таска `format` в `Taskfile.dist.yaml`. gci/gofumpt работают с текстом исходников, не требуют компиляции или сети.

**Верификация:** `task format` выполнен без сетевого доступа, `stapel:embed` не запускался.

### Risk #6 — расхождение неймспейсов `STAPEL_IMAGE`/`WERF_STAPEL_IMAGE_NAME` (commit `beb170bfe`)

`cmd/embed-stapel/main.go` переведён на `WERF_STAPEL_IMAGE_NAME`/`WERF_STAPEL_IMAGE_VERSION` — те же переменные, что использует runtime (`pkg/stapel/stapel.go`). `Taskfile.dist.yaml`: `stapel:embed` задаёт `WERF_STAPEL_IMAGE_VERSION` для собственного subprocess (не влияет на итоговый бинарь). `scripts/stapel/DEVELOPMENT.md` документирует общий неймспейс.

**Верификация:** `go build`/`task stapel:embed` — идентичные sha256 артефактов до/после переименования; `golangci-lint` — 0 issues.

### Risk #10 — silent decompress+load (commit `a3e1d5a69`)

`loadEmbeddedImage` (`pkg/stapel/embedded.go`) обёрнут в `logboek.Context(ctx).LogProcess("Loading embedded stapel image for %s", targetPlatform).DoError(...)`, паттерн идентичен существующему в `container.go`'s `CreateIfNotExist`.

**Верификация:** build/test с тегом `embedstapel` и без — exit 0, golangci-lint 0 issues.

### Risk #9 — конкурентная запись embed-файлов (commit `de0321177`)

`buildPlatform` (`cmd/embed-stapel/main.go`) переписан на write-to-temp + `os.Rename` (atomic на одной ФС). tar.gz и sha256 пишутся во временные файлы (`os.CreateTemp`/`path+".tmp"`), затем атомарно переименовываются в целевой путь.

**Верификация:** запущены 2 параллельных `go run ./cmd/embed-stapel` — оба завершились без ошибок, финальные файлы согласованы (корректный sha256), никаких stray `.tmp`-файлов не осталось.

### Risk #3 — нет кэширования `stapel:embed` (commit `d54f88724`)

Добавлены `sources: [pkg/stapel/stapel.go]` и `generates: [4 embed-артефакта]` в task `stapel:embed`. Task теперь content-hash'ирует источник (где живёт `VERSION`) и генерируемые файлы — при отсутствии изменений таск пропускается без сетевого вызова.

**Верификация:** повторный `task stapel:embed` (новый процесс) → "up to date", 0 сетевых вызовов (verbose-лог подтвердил). Искусственное изменение `VERSION` (0.7.1→0.7.0→обратно) → каждый раз триггерило re-pull с разным sha256, подтверждая корректность инвалидации кэша по содержимому.

### Risk #7 — user-facing документация (commit `46d8421ef`)

Добавлен абзац в `docs/pages_en/usage/build/stapel/instructions.md` и `docs/pages_ru/.../instructions.md` (секция про Stapel service image) — объясняет, что для `linux/amd64`/`linux/arm64` образ уже встроен в бинарь и не требует доступа к registry, с описанием fallback-условий.

### Risk #4 + #11 — пустой PR body / disclosure роста размера

`gh pr edit 7601` — заполнено полное описание: Summary, Key changes (7 пунктов), Why (motivation air-gapped/offline), явный **Size impact** раздел (~35-40MB на все 5 release-платформ, без opt-out кроме embedstapel-тега для dev tooling), Review focus/risks.

### Финальная сквозная верификация (после всех 6 фиксов)

- Чистый `task build` (embed-директория удалена перед запуском) → **exit 0**, бинарь с embedded stapel, ~163MB.
- `task format` — офлайн, без сетевых вызовов.
- `task lint golangciPaths="./cmd/... ./pkg/... ./test/..."` (весь проект) → **0 issues**.
- `task test:unit` (весь проект) → **Stapel Suite 3/3 SUCCESS**; 1 unrelated pre-existing failure в `pkg/werf/exec` (macOS process-spawn issue, "ps: cmd: keyword not found") — не связан с изменениями.

### Итоговый статус всех 12 рисков

| № | Риск | Итоговый статус |
| :--- | :--- | :--- |
| 1 | bare Go tooling breakage | ЗАКРЫТ |
| 2 | `test:unit` без `stapel:embed` deps | ЗАКРЫТ |
| 3 | нет кэширования | ЗАКРЫТ |
| 4 | недекларированный рост бинарника | ЗАКРЫТ (disclosure в PR) |
| 5 | `stapel:embed` привязан к `format` | ЗАКРЫТ |
| 6 | расхождение неймспейсов env vars | ЗАКРЫТ |
| 7 | user docs не обновлены | ЗАКРЫТ |
| 8 | нет opt-out для slim-сборок | ЗАКРЫТ (побочный эффект тега) |
| 9 | конкурентная запись embed-файлов | ЗАКРЫТ |
| 10 | silent decompress+load | ЗАКРЫТ |
| 11 | пустое тело PR | ЗАКРЫТ |
| 12 | забытый `-tags=embedstapel` на релизе | Митигирован |

**Все риски из отчёта закрыты.** PR #7601 готов к финальному ревью автором/мейнтейнерами werf.
