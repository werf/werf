```shell
$WERF_BIN_PATH = "C:\ProgramData\werf\bin"
mkdir $WERF_BIN_PATH

Invoke-WebRequest -Uri https://tuf.werf.io/targets/releases/{{ include.version }}/windows-{{ include.arch }}/bin/werf.exe -OutFile $WERF_BIN_PATH\werf.exe

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + "$WERF_BIN_PATH",
    [EnvironmentVariableTarget]::Machine)

$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```
