##### PowerShell

```shell
$MULTIWERF_BIN_PATH = "C:\ProgramData\multiwerf\bin"
mkdir $MULTIWERF_BIN_PATH

Invoke-WebRequest -Uri https://storage.yandexcloud.net/multiwerf/targets/releases/v1.4.7/multiwerf-windows-amd64-v1.4.7.exe -OutFile $MULTIWERF_BIN_PATH\multiwerf.exe

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + "$MULTIWERF_BIN_PATH",
    [EnvironmentVariableTarget]::Machine)

$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

###### Использование werf в текущей сессии shell

```shell
Invoke-Expression -Command "multiwerf use {{ include.version }} {{ include.channel }} --as-file --shell powershell" | Out-String -OutVariable WERF_USE_SCRIPT_PATH
. $WERF_USE_SCRIPT_PATH.Trim()
```

##### cmd.exe

Необходимо запустить `cmd.exe` от имени администратора и затем выполнить следующие команды:

```shell
set MULTIWERF_BIN_PATH="C:\ProgramData\multiwerf\bin"
mkdir %MULTIWERF_BIN_PATH%
bitsadmin.exe /transfer "multiwerf" https://storage.yandexcloud.net/multiwerf/targets/releases/v1.4.7/multiwerf-windows-amd64-v1.4.7.exe %MULTIWERF_BIN_PATH%\multiwerf.exe
setx /M PATH "%PATH%;%MULTIWERF_BIN_PATH%"
```

После этого **необходимо открыть новую сессию `cmd.exe`**, чтобы начать пользоваться werf.

###### Использование werf в текущей сессии shell

```shell
FOR /F "tokens=*" %g IN ('multiwerf use {{ include.version }} {{ include.channel }} --as-file --shell cmdexe') do (SET WERF_USE_SCRIPT_PATH=%g)
%WERF_USE_SCRIPT_PATH%
```
