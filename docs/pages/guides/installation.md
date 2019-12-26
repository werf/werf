---
title: Installation
sidebar: documentation
permalink: documentation/guides/installation.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

## Installing Dependencies

{% include /readme/installing_dependencies.md %}

## Installing werf 

### Method 1 (recommended): by using multiwerf

{% include /readme/installing_with_multiwerf.md %}

#### multiwerf use command in details

The command `multiwerf use MAJOR.MINOR CHANNEL` allows to use the actual werf binary and to be always up-to-date. 
It does not matter where to use this command whether in **CI** or **on the local machine**. 
Source result script output or script file (with `--as-file` option) and work with the actual werf binary **during shell session**.

The script can be divided into two logic parts: update and werf alias or function definition depending on shell type. 
The update part consists of multiwerf self-update and getting the actual werf binary for specified `MAJOR.MINOR` version and `CHANNEL` (read more about werf versioning in [Backward Compatibility Promise](https://github.com/flant/werf#backward-compatibility-promise) section).
Update is performed by the `multiwerf update` command. 
If the script is launched first time or there is not suitable werf binary locally these steps run consistently. 
Otherwise, update runs in the background and werf alias or function binds to the existing werf binary based on local channel mapping.

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

During update multiwerf tries to download the desirable werf version based on a channel mapping. 
The channel mapping is the special file that keeps relations between channels and werf versions.
By default multiwerf uses the mapping file which is maintained in werf repository ([https://github.com/flant/werf/blob/multiwerf/multiwerf.json](https://github.com/flant/werf/blob/multiwerf/multiwerf.json))

Such an approach allows a user not to think about updates and to use the same werf binary version on CI and on the local machine. 
We create new releases with fixes and features and manage channels while you just use a single command everywhere.

### Method 2: by downloading binary package

The latest release can be reached via [this page](https://bintray.com/flant/werf/werf/_latestVersion)

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

# open new cmd.exe session and start using werf
```

### Method 3: by compiling from source

```shell
go get github.com/flant/werf/cmd/werf
```

# Backward Compatibility Promise

{% include /readme/backward_compatibility_promise.md %}
