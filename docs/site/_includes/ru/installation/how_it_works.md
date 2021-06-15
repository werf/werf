Команда `multiwerf use MAJOR.MINOR CHANNEL` активирует определённую версию werf и автоматически обновляет werf до последней актуальной версии. Данную команду можно использовать как в **CI/CD системе**, так и на **локальной машине**. Команда печатает на экран скрипт (или путь до временного файла, содержащего скрипт — при использовании опции `--as-file`). Данный скрипт необходимо выполнить, например, с помощью директивы `source` в shell. В результате работы данного скрипта активируется определённая версия `werf` **в данной shell-сессии**.

Данный скрипт можно разбить на 2 логических этапа: обновление и создание символьного имени `werf`, которое связано с бинарным файлом определённой версии werf (в зависимости от используемого shell, это может быть shell-алиас или shell-функция). На этапе обновления происходит автоматическое обновление самого бинарника `multiwerf`, затем получение актуальной версии бинарника `werf`, соответствующего указанным параметрам `MAJOR.MINOR` и `CHANNEL` (см. больше информации про [гарантии обратной совместимости версий werf]({{ "installation.html#гарантии-обратной-совместимости" | true_relative_url }}). Этап обновления можно запустить отдельно командой `multiwerf update`.

В случае если скрипт был запущен впервые, он будет ожидать пока multiwerf скачает подходящую версию werf. Иначе, если на локальной машине уже есть какая-то версия werf, то скрипт активирует её, а обновление запустится в фоновом режиме. Соответственно при следующем запуске будет активирована уже обновлённая версия `werf`. 

<div class="tabs tabs_simple">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'unix_tab')">Unix shell</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'powershell_tab')">PowerShell</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'cmdexe_tab')">cmd.exe</a>
</div>

<div id="unix_tab" class="tabs__content tabs__content_simple active" markdown="1">

```shell
if multiwerf werf-path MAJOR.MINOR CHANNEL >~/.multiwerf/multiwerf_use_first_werf_path.log 2>&1; then
    multiwerf update MAJOR.MINOR CHANNEL --in-background --output-file=~/.multiwerf/multiwerf_use_background_update.log --with-cache
else
    multiwerf update MAJOR.MINOR CHANNEL
fi

WERF_PATH=$(multiwerf werf-path MAJOR.MINOR CHANNEL)
WERF_FUNC=$(cat <<EOF
werf()
{
    $WERF_PATH "\$@"
}
EOF
)

eval "$WERF_FUNC"
```

</div>

<div id="powershell_tab" class="tabs__content tabs__content_simple" markdown="1">

```shell
if ((Invoke-Expression -Command "multiwerf werf-path MAJOR.MINOR CHANNEL" | Out-String -OutVariable WERF_PATH) -and ($LastExitCode -eq 0)) {
    multiwerf update MAJOR.MINOR CHANNEL --in-background --output-file=~\.multiwerf\multiwerf_use_background_update.log --with-cache
} else {
    multiwerf update MAJOR.MINOR CHANNEL
    Invoke-Expression -Command "multiwerf werf-path MAJOR.MINOR CHANNEL" | Out-String -OutVariable WERF_PATH
}

function werf { & $WERF_PATH.Trim() $args }
```

</div>

<div id="cmdexe_tab" class="tabs__content tabs__content_simple" markdown="1">

```shell
FOR /F "tokens=*" %%g IN ('multiwerf werf-path MAJOR.MINOR CHANNEL') do (SET WERF_PATH=%%g)

IF %ERRORLEVEL% NEQ 0 (
    multiwerf update MAJOR.MINOR CHANNEL
    FOR /F "tokens=*" %%g IN ('multiwerf werf-path MAJOR.MINOR CHANNEL') do (SET WERF_PATH=%%g)
) ELSE (
    multiwerf update MAJOR.MINOR CHANNEL --in-background --output-file=~\.multiwerf\multiwerf_use_background_update.log --with-cache
)

DOSKEY werf=%WERF_PATH% $*
```

</div>

Во время этапа обновления, multiwerf попытается скачать требуемую версию werf на основе файла соответствия версий и каналов. По умолчанию multiwerf использует файл соответствия версий, который объявлен в репозитории werf: [https://github.com/werf/werf/blob/multiwerf/multiwerf.json](https://github.com/werf/werf/blob/multiwerf/multiwerf.json). Однако с помощью опции `--channel-mapping-url` можно указать любой url, по которому может быть доступен произвольный файл соответствия версий и каналов.

Данный подход позволяет пользователю не думать об обновлениях werf и получать исправления проблем и новые возможности автоматически как в CI/CD системе, так и на локальной машине.
