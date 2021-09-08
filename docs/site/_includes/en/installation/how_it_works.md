The command `trdl use werf MAJOR.MINOR CHANNEL` allows using the actual werf binary and to be always up-to-date. 
You can use this command both in **CI/CD system** and **on the local machine**. 
The command returns a path to the script file that must be used as an argument to the `source` statement in the case of using shell. As a result, the current version of werf will be available **during the shell session**.

The script can be divided into two logic parts: updating and adding werf into the PATH. The update part performs trdl self-update and gets the actual werf binary for the specified `MAJOR.MINOR` version and `CHANNEL` (read more about werf versioning in the [Backward Compatibility Promise]({{ "installation.html#backward-compatibility-promise" | true_relative_url }}) section). The update is performed by the `trdl update werf` command.

If the script is launched for the first time or there is no suitable werf binary found locally, this script will wait until trdl downloads new werf binary. Otherwise, the update runs in the background, and werf alias or a function binds to the existing werf binary based on local channel mapping.

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

During the update, trdl tries to download the desirable werf version based on a channel mapping. 
The channel mapping is a special file that keeps relations between channels and werf versions.
By default, trdl uses the mapping file which is maintained in the werf repository: [https://raw.githubusercontent.com/werf/werf/multiwerf/trdl_channels.yaml](https://raw.githubusercontent.com/werf/werf/multiwerf/trdl_channels.yaml).

Such an approach allows a user not to worry about updates and use the same werf binary version on CI and the local machine. 
We create new releases with fixes and features and manage channels while you simply use a single command everywhere.
