Make sure you have [Git](https://git-scm.com/download/win) 2.18.0 or newer and [Docker](https://docs.docker.com/get-docker) installed.

Following instructions should be executed in the PowerShell.

Setup [trdl](https://github.com/werf/trdl) which will manage `werf` installation and updates:
```powershell
# Add %USERPROFILE%\bin to the PATH.
[Environment]::SetEnvironmentVariable("Path", "$env:USERPROFILE\bin" + [Environment]::GetEnvironmentVariable("Path", "User"), "User")
$env:Path = "$env:USERPROFILE\bin;$env:Path"

# Install trdl.
mkdir -Force "$env:USERPROFILE\bin"
Invoke-WebRequest -Uri "https://tuf.trdl.dev/targets/releases/0.1.3/windows-{{ include.arch }}/bin/trdl.exe" -OutFile "$env:USERPROFILE\bin\trdl.exe"
```

Add `werf` repo:
```powershell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

To use werf locally in your terminal, we recommend to enable its automatic activation. To make werf available in all new PowerShell sessions, you need to execute this command (just once, run under Administrator):
```powershell
# Allow execution of locally created scripts.
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# Activate werf binary automatically during PowerShell initializations.
if (!(Test-Path "$profile")) {
  New-Item -Path "$profile" -Force
}
Add-Content -Path "$profile" -Value '. $(trdl use werf {{ include.version }} {{ include.channel }})'
```

Now, if you log out and log in to the system again, werf will be always available. You can make sure of that by executing:

```shell
werf version
```

To get werf running in your current terminal only (before any logout/login is done), you can simply execute the `. $(trdl use werf {{ include.version }} {{ include.channel }})` command.

In CI, you need a different approaching with activating `werf` explicitly in the beginning of each job/pipeline by executing:

```shell
. $(trdl use werf {{ include.version }} {{ include.channel }})
```
