<p align="center">
  <img src="https://raw.githubusercontent.com/werf/website/main/assets/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>
___

## Development

### Run the documentation part of the site locally

#### Variant 1 (full)

Run `jekyll serve` with --watch option to test changes in "real time". Requires to run werf compose in werf/website to access documentation in browser.

- Install [werf](http://werf.io/installation.html).
- Install [task](https://taskfile.dev/installation/).
- Run:
  ```shell
  task compose:up
  ```
- Wait (approximately 60 seconds) for the message "done in X.XXX seconds" from the `docs-en-1` and `docs-ru-1` containers.
- Run werf/website:
  ```shell
  cd ../website
  task compose:up
  ```
- Check the English version is available on [https://localhost](http://localhost), and the Russian version on [http://ru.localhost](https://ru.localhost) (add `ru.localhost` record in your `/etc/hosts` to access the Russian version of the site). 

#### Variant 2 (standalone)

Run `jekyll serve` with --watch option to test changes in "real time". Use scripts/styles/images from werf.io site.

- Install [werf](http://werf.io/installation.html). 
- Install [task](https://taskfile.dev/installation/)
- Run (add `--follow --docker-compose-command-options="-d"` if necessary):
  ```shell
  task compose:up:standalone
  ```
- Check the English version is available on [https://localhost](http://localhost), and the Russian version on [http://ru.localhost](https://ru.localhost) (add `ru.localhost` record in your `/etc/hosts` to access the Russian version of the site). 
