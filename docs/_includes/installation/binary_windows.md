##### PowerShell

```shell
$WERF_BIN_PATH = "C:\ProgramData\werf\bin"
mkdir $WERF_BIN_PATH

Invoke-WebRequest -Uri https://dl.bintray.com/flant/werf/v1.1.21+fix22/werf-windows-amd64-v1.1.21+fix22.exe -OutFile $WERF_BIN_PATH\werf.exe

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
bitsadmin.exe /transfer "werf" https://dl.bintray.com/flant/werf/v1.1.21+fix22/werf-windows-amd64-v1.1.21+fix22.exe %WERF_BIN_PATH%\werf.exe
setx /M PATH "%PATH%;%WERF_BIN_PATH%"
```
