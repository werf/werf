First ensure that multiwerf exists and is executable, use the `type` command. The command prints a message to stderr if multiwerf is not found. Thus, diagnostics in a CI/CD environment becomes simpler. Second enable werf in the CI/CD shell.

```shell
type multiwerf && . $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)
# We can use werf now
werf ...
```
