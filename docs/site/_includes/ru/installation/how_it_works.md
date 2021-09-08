Команда `trdl use werf MAJOR.MINOR CHANNEL` активирует определённую версию werf и автоматически обновляет werf до последней актуальной версии. Данную команду можно использовать как в **CI/CD системе**, так и на **локальной машине**. Команда печатает на экран путь до временного файла, содержащего скрипт. Данный скрипт необходимо выполнить, например, с помощью директивы `source` в shell. В результате работы данного скрипта активируется определённая версия `werf` **в данной shell-сессии**.

Данный скрипт можно разбить на 2 логических этапа: обновление и добавление в PATH бинарного файла werf определённой версии. На этапе обновления происходит автоматическое обновление самого бинарника `trdl`, затем получение актуальной версии бинарника `werf`, соответствующего указанным параметрам `MAJOR.MINOR` и `CHANNEL` (см. больше информации про [гарантии обратной совместимости версий werf]({{ "installation.html#гарантии-обратной-совместимости" | true_relative_url }}). Этап обновления можно запустить отдельно командой `trdl update werf`.

В случае если скрипт был запущен впервые, он будет ожидать пока trdl скачает подходящую версию werf. Иначе, если на локальной машине уже есть какая-то версия werf, то скрипт активирует её, а обновление запустится в фоновом режиме. Соответственно при следующем запуске будет активирована уже обновлённая версия `werf`. 

<div class="tabs tabs_simple">
  <a href="javascript:void(0)" class="tabs__btn tabs__howitworks__btn active" onclick="openTab(event, 'tabs__howitworks__btn', 'tabs__howitworks__content', 'howitworks__unix_tab')">Unix shell</a>
  <a href="javascript:void(0)" class="tabs__btn tabs__howitworks__btn" onclick="openTab(event, 'tabs__howitworks__btn', 'tabs__howitworks__content', 'howitworks__powershell_tab')">PowerShell</a>
</div>

<div id="howitworks__unix_tab" class="tabs__content tabs__howitworks__content tabs__content_simple active" markdown="1">

```shell
if [ -s "$HOME/.trdl/logs/repositories/werf/use_1.2_ea_unix_background_update_stderr.log" ]; then
   echo Previous run of "trdl update" in background generated following errors:
   cat "$HOME/.trdl/logs/repositories/werf/use_1.2_ea_unix_background_update_stderr.log"
fi

if trdl_repo_bin_path="$("trdl" bin-path werf 1.2 ea 2>/dev/null)"; then
   "trdl" update werf 1.2 ea --in-background --background-stdout-file="$HOME/.trdl/logs/repositories/werf/use_1.2_ea_unix_background_update_stdout.log" --background-stderr-file="$HOME/.trdl/logs/repositories/werf/use_1.2_ea_unix_background_update_stderr.log"
else
   "trdl" update werf 1.2 ea
   trdl_repo_bin_path="$("trdl" bin-path werf 1.2 ea)"
fi

export PATH="$trdl_repo_bin_path${PATH:+:${PATH}}"
```

</div>

<div id="howitworks__powershell_tab" class="tabs__content tabs__howitworks__content tabs__content_simple" markdown="1">

```shell
if (Test-Path "$HOME/.trdl/logs/repositories/werf/use_1.2_ea_pwsh_background_update_stderr.log" -PathType Leaf) {
  $trdlStderrLog = Get-Content "$HOME/.trdl/logs/repositories/werf/use_1.2_ea_pwsh_background_update_stderr.log"
  if (!([String]::IsNullOrWhiteSpace($trdlStderrLog))) {
    'Previous run of "trdl update" in background generated following errors:'
    $trdlStderrLog
  }
}

if ((Invoke-Expression -Command "trdl bin-path werf 1.2 ea" 2> $null | Out-String -OutVariable trdlRepoBinPath) -and ($LastExitCode -eq 0)) {
   trdl update werf 1.2 ea --in-background --background-stdout-file="$HOME/.trdl/logs/repositories/werf/use_1.2_ea_pwsh_background_update_stdout.log" --background-stderr-file="$HOME/.trdl/logs/repositories/werf/use_1.2_ea_pwsh_background_update_stderr.log"
} else {
   trdl update werf 1.2 ea
   $trdlRepoBinPath = trdl bin-path werf 1.2 ea
}

$trdlRepoBinPath = $trdlRepoBinPath.Trim()
$oldPath = [System.Environment]::GetEnvironmentVariable('PATH',[System.EnvironmentVariableTarget]::Process)
$newPath = "$trdlRepoBinPath;$oldPath"
[System.Environment]::SetEnvironmentVariable('Path',$newPath,[System.EnvironmentVariableTarget]::Process);
```

</div>

Во время этапа обновления, trdl попытается скачать требуемую версию werf на основе файла соответствия версий и каналов. По умолчанию trdl использует файл соответствия версий, который объявлен в репозитории werf: [https://raw.githubusercontent.com/werf/werf/multiwerf/trdl_channels.yaml](https://raw.githubusercontent.com/werf/werf/multiwerf/trdl_channels.yaml).

Данный подход позволяет пользователю не думать об обновлениях werf и получать исправления проблем и новые возможности автоматически как в CI/CD системе, так и на локальной машине.
