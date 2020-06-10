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

The command `multiwerf use MAJOR.MINOR CHANNEL` allows using the actual werf binary and to be always up-to-date. 
You can use this command both in **CI** and **on the local machine**. 
The command returns a script or a path to the script file (when used with an `--as-file` option) that must be used as an argument to the `source` command. As a result, the current version of werf will be available **during the shell session**.

The script can be divided into two logic parts: updating and creating werf alias or the definition of the function depending on shell type. 
The update part performs multiwerf self-update and gets the actual werf binary for the specified `MAJOR.MINOR` version and `CHANNEL` (read more about werf versioning in the [Backward Compatibility Promise](https://github.com/werf/werf#backward-compatibility-promise) section).
The update is performed by the `multiwerf update` command. 
If the script is launched for the first time or there is no suitable werf binary found locally, these steps are being run consistently. 
Otherwise, the update runs in the background, and werf alias or a function binds to the existing werf binary based on local channel mapping.

<div class="tabs">
  <a href="javascript:void(0)" class="tabs__btn active" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'unix_tab')">Unix shell</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'powershell_tab')">PowerShell</a>
  <a href="javascript:void(0)" class="tabs__btn" onclick="openTab(event, 'tabs__btn', 'tabs__content', 'cmdexe_tab')">cmd.exe</a>
</div>

<div id="unix_tab" class="tabs__content active" markdown="1">

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

<div id="powershell_tab" class="tabs__content" markdown="1">

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

<div id="cmdexe_tab" class="tabs__content" markdown="1">

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

During the update, multiwerf tries to download the desirable werf version based on a channel mapping. 
The channel mapping is a special file that keeps relations between channels and werf versions.
By default, multiwerf uses the mapping file which is maintained in the werf repository ([https://github.com/werf/werf/blob/multiwerf/multiwerf.json](https://github.com/werf/werf/blob/multiwerf/multiwerf.json))

Such an approach allows a user not to worry about updates and use the same werf binary version on CI and the local machine. 
We create new releases with fixes and features and manage channels while you simply use a single command everywhere.

### Method 2: by downloading binary package

The latest release can be found at [this page](https://bintray.com/flant/werf/werf/_latestVersion)

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
go get github.com/werf/werf/cmd/werf
```

# Backward Compatibility Promise

{% include /readme/backward_compatibility_promise.md %}
