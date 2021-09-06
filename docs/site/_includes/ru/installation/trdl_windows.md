##### 1. Установка [trdl](https://github.com/werf/trdl)

```shell
$TRDL_BIN_PATH = "C:\ProgramData\trdl\bin"
mkdir $TRDL_BIN_PATH

Invoke-WebRequest -Uri https://tuf.trdl.dev/targets/releases/0.1.3/windows-{{ include.arch }}/bin/trdl.exe -OutFile $TRDL_BIN_PATH\trdl.exe

[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine) + "$TRDL_BIN_PATH",
    [EnvironmentVariableTarget]::Machine)

$env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
```

##### 2. Добавление официального TUF-репозитория werf в trdl

```shell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

##### 3. Использование werf в текущей сессии shell

```shell
Invoke-Expression -Command "trdl use werf {{ include.version }} {{ include.channel }} --shell powershell" | Out-String -OutVariable WERF_USE_SCRIPT_PATH
. $WERF_USE_SCRIPT_PATH.Trim()
```
