---
title: Установка
sidebar: documentation
permalink: documentation/guides/installation.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

## Установка зависимостей 

{% include /readme_ru/installing_dependencies.md %}

## Установка werf 

### Способ 1 (рекомендуемый): multiwerf

{% include /readme/installing_with_multiwerf.md %}

#### Команда multiwerf use в деталях

Команда `multiwerf use MAJOR.MINOR CHANNEL` позволяет использовать актуальную версию werf для указанных параметров `MAJOR.MINOR` и `CHANNEL`. 
Команду следует использовать как в CI, так и при локальной разработке. 
Команда возвращает скрипт или, при использовании опции `--as-file`, путь к файлу со скриптом, который необходимо использовать как аргумент команды `source`. 
В результате в shell-сессии будет доступна актуальная версия werf. 

Скрипт можно разделить на две составляющие: обновление и создание alias или функции werf, в зависимости от shell. 
На стадии обновления выполняется проверка и обновление multiwerf, а также скачивается актуальная версия werf на основе указанных параметров `MAJOR.MINOR` и `CHANNEL` (подробнее про версионирование werf в разделе [Обещание обратной совместимости](https://github.com/flant/werf/blob/master/README_ru.md#обещание-обратной-совместимости)). 

Если скрипт выполняется впервые или для указанных параметров `MAJOR.MINOR` и `CHANNEL` локально нет подходящих версий, исходя из локального файла соответствий версий и каналов, то шаги обновления и создания alias werf выполняются последовательно. 
Иначе обновление запускается в фоновом режиме, alias или функция werf ссылается на существующий локально бинарный файл, а результат выполнения обновления никак не повлияет на текущую сессию.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'unix')">Unix shell</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'powershell')">PowerShell</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'cmdexe')">cmd.exe</a>
</div>

<div id="unix" class="tabs__content active" markdown="1">

```shell
if multiwerf werf-path MAJOR.MINOR CHANNEL >~\.multiwerf\multiwerf_use_first_werf_path.log 2>&1; then
    (multiwerf update --with-cache MAJOR.MINOR CHANNEL >~\.multiwerf\background_update.log 2>&1 </dev/null &)
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

<div id="powershell" class="tabs__content" markdown="1">

```shell
if (Invoke-Expression -Command "multiwerf werf-path MAJOR.MINOR CHANNEL >~\.multiwerf\multiwerf_use_first_werf_path.log 2>&1" | Out-String -OutVariable WERF_PATH) {
    Start-Job { multiwerf update --with-cache MAJOR.MINOR CHANNEL >~\.multiwerf\background_update.log 2>&1 }
} else {
    multiwerf update MAJOR.MINOR CHANNEL
    Invoke-Expression -Command "multiwerf werf-path MAJOR.MINOR CHANNEL" | Out-String -OutVariable WERF_PATH
}

function werf { & $WERF_PATH.Trim() $args }
```

</div>

<div id="cmdexe" class="tabs__content" markdown="1">

```shell
FOR /F "tokens=*" %%g IN ('multiwerf werf-path MAJOR.MINOR CHANNEL') do (SET WERF_PATH=%%g)

IF %ERRORLEVEL% NEQ 0 (
    multiwerf update MAJOR.MINOR CHANNEL 
    FOR /F "tokens=*" %%g IN ('multiwerf werf-path MAJOR.MINOR CHANNEL') do (SET WERF_PATH=%%g)
) ELSE (
    START /B multiwerf update MAJOR.MINOR CHANNEL >~/.multiwerf/background_update.log 2>&1
)

DOSKEY werf=%WERF_PATH% $*
```

</div>

При обновлении multiwerf пытается скачать правильную версию werf, основываясь на специальном файле соответствий версий и каналов.
По умолчанию multiwerf использует файл соответствий, который хранится и сопровождается в репозитории werf ([https://github.com/flant/werf/blob/multiwerf/multiwerf.json](https://github.com/flant/werf/blob/multiwerf/multiwerf.json))

Такой подход позволяет использовать одни и теже версии werf локально и в CI, а также не задумываться об обновлениях. 
Разработчики werf создают релизы с правками и новым функционалом, переключают версии в каналах, в то время как пользователь, везде использует одну единственную команду.

### Способ 2: установка бинарного файла

Выберите подходящую версию из [релизов на GitHub](https://github.com/flant/werf/releases) или на [bintray](https://bintray.com/flant/werf/werf/_latestVersion) и используйте один из предлженных подходов с выбранным URL-адресом.

#### MacOS

```shell
curl -L https://dl.bintray.com/flant/werf/v1.0.6-rc.5/werf-darwin-amd64-v1.0.6-rc.5 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

#### Linux

```shell
curl -L https://dl.bintray.com/flant/werf/v1.0.6-rc.5/werf-linux-amd64-v1.0.6-rc.5 -o /tmp/werf
chmod +x /tmp/werf
sudo mv /tmp/werf /usr/local/bin/werf
```

#### Windows

##### PowerShell

```shell
$WERF_BIN_PATH = "C:\ProgramData\werf\bin"
mkdir $WERF_BIN_PATH

Invoke-WebRequest -Uri https://dl.bintray.com/flant/werf/v1.0.6-rc.5/werf-windows-amd64-v1.0.6-rc.5.exe -OutFile $WERF_BIN_PATH\werf.exe

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + "$WERF_BIN_PATH",
    [EnvironmentVariableTarget]::Machine)

$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

##### cmd.exe (run as Administrator)

```shell
set WERF_BIN_PATH="C:\ProgramData\werf\bin"
mkdir %WERF_BIN_PATH%
bitsadmin.exe /transfer "werf" https://dl.bintray.com/flant/werf/v1.0.6-rc.5/werf-windows-amd64-v1.0.6-rc.5.exe %WERF_BIN_PATH%\werf.exe
setx /M PATH "%PATH%;%WERF_BIN_PATH%"

# откройте новую сессию и начните использовать werf
```

### Способ 3: сборка из исходников

```shell
go get github.com/flant/werf/cmd/werf
```

# Обещание обратной совместимости

{% include /readme_ru/backward_compatibility_promise.md %}
