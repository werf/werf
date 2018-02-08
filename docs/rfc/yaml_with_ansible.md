# Builder configuration

Builder configuration is a collection of yaml-docs. These yaml-docs could at the same time reside in:

* `dappfile.yml` or `dappfile.yaml` file in the project root;
* In multiples yaml files from directory `.dappfiles`.

Dapp will read files from directory `.dappfiles` in the alphabetical order. `dappfile.yml` will preceed any files from `.dappfiles`

## Ansible

Supported modules:

* Command
* Shell
* Copy

### How to create config file in image from ansible jinja-template

* Use module `Copy` with `content` directive.
* Put config template directly into dappfile.
* Or put config into separate project file, then read content of file into dappfile using go-templates function `.Files.Get <path>`
