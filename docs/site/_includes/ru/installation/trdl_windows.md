Убедитесь, что [Git](https://git-scm.com/download/win) версии 2.18.0 или новее, gpg и [Docker](https://docs.docker.com/get-docker) установлены. Дальнейшие инструкции должны выполняться в PowerShell.

[Установите trdl](https://github.com/werf/trdl/releases/) в `<диск>:\Users\<имя пользователя>\bin\trdl`. `trdl` будет отвечать за установку и обновление `werf`. Добавьте `<диск>:\Users\<имя пользователя>\bin\` в переменную окружения $PATH.

Добавьте `werf`-репозиторий в `trdl`:
```powershell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```
 
Для использования `werf` на рабочей машине мы рекомендуем настроить для `werf` _автоматическую активацию_. Для этого команда активации должна запускаться для каждой новой PowerShell-сессии. В PowerShell для этого обычно надо добавить команду активации в [$PROFILE-файл](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_profiles). Команда активации `werf` для текущей PowerShell-сессии:
```powershell
. $(trdl use werf {{ include.version }} {{ include.channel }})
```

Для использования `werf` в CI вместо автоматической активации предпочитайте активацию `werf` вручную. Для этого выполните команду активации в начале вашей CI job, до вызова самого `werf`.

После активации `werf` должен быть доступен в той же PowerShell-сессии, в которой он был активирован:
```powershell
werf version
```
