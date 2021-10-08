Убедитесь, что Git версии 2.18.0 или новее и [Docker](https://docs.docker.com/get-docker) установлены.

Выполните в PowerShell:
```powershell
# Add %USERPROFILE%\bin to the PATH.
[Environment]::SetEnvironmentVariable("Path", "$env:USERPROFILE\bin" + [Environment]::GetEnvironmentVariable("Path", "User"), "User")
$env:Path = "$env:USERPROFILE\bin;$env:Path"

# Install werf.
mkdir -Force "$env:USERPROFILE\bin"
Invoke-WebRequest -Uri "https://tuf.werf.io/targets/releases/{{ include.version }}/windows-{{ include.arch }}/bin/werf.exe" -OutFile "$env:USERPROFILE\bin\werf.exe"
```
