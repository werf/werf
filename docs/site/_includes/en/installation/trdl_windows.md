Make sure you have [Git](https://git-scm.com/download/win) 2.18.0 or newer, gpg and [Docker](https://docs.docker.com/get-docker) installed. Following instructions should be executed in PowerShell.

[Install trdl](https://github.com/werf/trdl/releases/) to `<disk>:\Users\<your username>\bin\trdl`, which will manage `werf` installation and updates. Add `<disk>:\Users\<your username>\bin\` to your $PATH environment variable.

Add `werf` repo to `trdl`:
```powershell
trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2
```

To use `werf` on a workstation we recommend setting up `werf` _automatic activation_. For this the activation command should be executed for each new PowerShell session. For PowerShell this is usually achieved by adding the activation command to [$PROFILE file](https://docs.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_profiles). This is the `werf` activation command for the current PowerShell-session:
```powershell
. $(trdl use werf {{ include.version }} {{ include.channel }})
```

To use `werf` in CI prefer activating `werf` manually instead. For this execute the activation command in the beginning of your CI job, before calling the `werf` binary.

After activation `werf` should be available in the PowerShell-session from which it was activated:
```powershell
werf version
```
