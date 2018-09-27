---
title: Minimal dappfile
sidebar: reference
permalink: reference/dappfile/minimal_dappfile.html
---


Dappfile should contains the following directives:
- `dimg`
- `from`


The `dimg` directive can be the one and can be without name.


Here's a minimal dappfile. It builds an image named `example` from a base image named `alpine`:

```yaml
dimg: example
from: alpine
```

Here's the same dappfile, except the `dimg` directive has no dimg name:

```
dimg: ~
from: alpine
```
