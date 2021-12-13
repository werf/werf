## Changelog


### Bug Fixes

RELEASE MESSAGE

## Installation

To install `werf` we strongly recommend following [these instructions](https://werf.io/installation.html).

Alternatively, you can download `werf` binaries from here:
* [Linux amd64](https://tuf.werf.io/targets/releases/$VERSION/linux-amd64/bin/werf) ([PGP signature](https://tuf.werf.io/targets/signatures/$VERSION/linux-amd64/bin/werf.sig))
* [Linux arm64](https://tuf.werf.io/targets/releases/$VERSION/linux-arm64/bin/werf) ([PGP signature](https://tuf.werf.io/targets/signatures/$VERSION/linux-arm64/bin/werf.sig))
* [macOS amd64](https://tuf.werf.io/targets/releases/$VERSION/darwin-amd64/bin/werf) ([PGP signature](https://tuf.werf.io/targets/signatures/$VERSION/darwin-amd64/bin/werf.sig))
* [macOS arm64](https://tuf.werf.io/targets/releases/$VERSION/darwin-arm64/bin/werf) ([PGP signature](https://tuf.werf.io/targets/signatures/$VERSION/darwin-arm64/bin/werf.sig))
* [Windows amd64](https://tuf.werf.io/targets/releases/$VERSION/windows-amd64/bin/werf.exe) ([PGP signature](https://tuf.werf.io/targets/signatures/$VERSION/windows-amd64/bin/werf.exe.sig))

These binaries were signed with PGP and could be verified with the [werf PGP public key](https://werf.io/werf.asc). For example, `werf` binary can be downloaded and verified with `gpg` on Linux with these commands:
```shell
curl -sSLO "https://tuf.werf.io/targets/releases/$VERSION/linux-amd64/bin/werf" -O "https://tuf.werf.io/targets/signatures/$VERSION/linux-amd64/bin/werf.sig"
curl -sSL https://werf.io/werf.asc | gpg --import
gpg --verify werf.sig werf
```
