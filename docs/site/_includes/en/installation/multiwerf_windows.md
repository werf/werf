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

###### Using werf in the current shell

```shell
Invoke-Expression -Command "multiwerf use {{ include.version }} {{ include.channel }} --as-file --shell powershell" | Out-String -OutVariable WERF_USE_SCRIPT_PATH
. $WERF_USE_SCRIPT_PATH.Trim()
```

##### cmd.exe

Run cmd.exe as Administrator and then do the following:

```shell
set MULTIWERF_BIN_PATH="C:\ProgramData\multiwerf\bin"
mkdir %MULTIWERF_BIN_PATH%
bitsadmin.exe /transfer "multiwerf" https://storage.yandexcloud.net/multiwerf/targets/releases/v1.4.7/multiwerf-windows-amd64-v1.4.7.exe %MULTIWERF_BIN_PATH%\multiwerf.exe
setx /M PATH "%PATH%;%MULTIWERF_BIN_PATH%"
```

Next it is required to open **new cmd.exe session** to start using werf.

###### Using werf in the current shell

```shell
FOR /F "tokens=*" %g IN ('multiwerf use {{ include.version }} {{ include.channel }} --as-file --shell cmdexe') do (SET WERF_USE_SCRIPT_PATH=%g)
%WERF_USE_SCRIPT_PATH%
```
