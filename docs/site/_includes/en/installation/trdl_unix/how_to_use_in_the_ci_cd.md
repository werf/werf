First ensure that `trdl` binary exists and is executable, use the `type` command. The command prints a message to stderr if `trdl` is not found. Thus, diagnostics in a CI/CD environment becomes simpler.

Second enable werf in the CI/CD shell.

```shell
type trdl && . $(trdl use werf {{ include.version }} {{ include.channel }})
# We can use werf now
werf ...
```
