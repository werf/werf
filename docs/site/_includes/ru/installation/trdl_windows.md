Убедитесь, что [Git](https://git-scm.com/download/win) версии 2.18.0 или новее и [Docker](https://docs.docker.com/get-docker) установлены.

Описанные далее команды должны быть выполнены в PowerShell.

Установите [trdl](https://github.com/werf/trdl), который будет отвечать за установку и обновление `werf`:
```powershell
# Добавьте %USERPROFILE%\bin в PATH.
[Environment]::SetEnvironmentVariable("Path", "$env:USERPROFILE\bin" + [Environment]::GetEnvironmentVariable("Path", "User"), "User")
$env:Path = "$env:USERPROFILE\bin;$env:Path"

# Установите trdl.
mkdir -Force "$env:USERPROFILE\bin"
Invoke-WebRequest -Uri "https://tuf.trdl.dev/targets/releases/0.1.3/windows-{{ include.arch }}/bin/trdl.exe" -OutFile "$env:USERPROFILE\bin\trdl.exe"
```

Добавьте репозиторий с `werf`:
```powershell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

Для локального использования werf (в терминале) мы рекомендуем настроить автоматическую активацию утилиты. Чтобы werf была доступна во всех новых PowerShell-сессиях, выполните следующую команду (это потребуется сделать лишь один раз) от имени Администратора:
```powershell
# Разрешите выполнение локально созданных скриптов.
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# Включите автоматическую активацию werf при инициализации PowerShell.
if (!(Test-Path "$profile")) {
  New-Item -Path "$profile" -Force
}
Add-Content -Path "$profile" -Value '. $(trdl use werf {{ include.version }} {{ include.channel }})'
```

Теперь, если вы выйдете из системы и залогинитесь в неё обратно, werf всегда будет доступна. Убедиться в этом можно следующей командой:

```powershell
werf version
```

Чтобы получить werf только в текущем терминале (до того, как перезашли в систему), достаточно выполнить команду `. $(trdl use werf {{ include.version }} {{ include.channel }})`.

Для CI рекомендуется другой подход с явной активацией `werf` в начале каждого job/pipeline. Она выполняется командой:

```shell
. $(trdl use werf {{ include.version }} {{ include.channel }})
```
