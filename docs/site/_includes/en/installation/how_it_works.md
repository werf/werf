The command `multiwerf use MAJOR.MINOR CHANNEL` allows using the actual werf binary and to be always up-to-date. 
You can use this command both in **CI/CD system** and **on the local machine**. 
The command returns a script or a path to the script file (when used with an `--as-file` option) that must be used as an argument to the `source` statement in the case of using shell. As a result, the current version of werf will be available **during the shell session**.

The script can be divided into two logic parts: updating and creating werf alias or the definition of the function depending on shell type. The update part performs multiwerf self-update and gets the actual werf binary for the specified `MAJOR.MINOR` version and `CHANNEL` (read more about werf versioning in the [Backward Compatibility Promise]({{ "installation.html#backward-compatibility-promise" | true_relative_url }}) section). The update is performed by the `multiwerf update` command.

If the script is launched for the first time or there is no suitable werf binary found locally, this script will wait until multiwerf downloads new werf binary. Otherwise, the update runs in the background, and werf alias or a function binds to the existing werf binary based on local channel mapping.

<div class="tabs tabs_simple">
  <a href="javascript:void(0)" class="tabs__btn tabs__howitworks__btn active" onclick="openTab(event, 'tabs__howitworks__btn', 'tabs__howitworks__content', 'howitworks__unix_tab')">Unix shell</a>
  <a href="javascript:void(0)" class="tabs__btn tabs__howitworks__btn" onclick="openTab(event, 'tabs__howitworks__btn', 'tabs__howitworks__content', 'howitworks__powershell_tab')">PowerShell</a>
  <a href="javascript:void(0)" class="tabs__btn tabs__howitworks__btn" onclick="openTab(event, 'tabs__howitworks__btn', 'tabs__howitworks__content', 'howitworks__cmdexe_tab')">cmd.exe</a>
</div>

<div id="howitworks__unix_tab" class="tabs__content tabs__howitworks__content tabs__content_simple active" markdown="1">

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

<div id="howitworks__powershell_tab" class="tabs__content tabs__howitworks__content tabs__content_simple" markdown="1">

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

<div id="howitworks__cmdexe_tab" class="tabs__content tabs__howitworks__content tabs__content_simple" markdown="1">

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
By default, multiwerf uses the mapping file which is maintained in the werf repository: [https://github.com/werf/werf/blob/multiwerf/multiwerf.json](https://github.com/werf/werf/blob/multiwerf/multiwerf.json). You can always use `--channel-mapping-url` option and specify arbitrary url with channel mapping which you can control by yourself.

Such an approach allows a user not to worry about updates and use the same werf binary version on CI and the local machine. 
We create new releases with fixes and features and manage channels while you simply use a single command everywhere.
