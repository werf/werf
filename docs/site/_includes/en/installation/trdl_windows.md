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

For local usage we recommend automatically activating `werf` for new PowerShell sessions. Run in the PowerShell under Administrator:
```powershell
# Allow execution of locally created scripts.
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# Activate werf binary

. $(trdl use werf {{ include.version }} {{ include.channel }})

# Activate werf binary automatically during PowerShell initializations.
if (!(Test-Path "$profile")) {
  New-Item -Path "$profile" -Force
}
Add-Content -Path "$profile" -Value '. $(trdl use werf {{ include.version }} {{ include.channel }})'
```

But in CI you should prefer activating `werf` explicitly in the beginning of each job/pipeline:
```shell
. $(trdl use werf {{ include.version }} {{ include.channel }})
```

Make sure that `werf` is available now (open new PowerShell session if you chose automatic activation):
```powershell
werf version
```
