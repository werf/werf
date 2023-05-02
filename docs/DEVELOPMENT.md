<p align="center">
  <img src="https://raw.githubusercontent.com/werf/website/main/assets/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>
___

## Development

### Run the documentation part of the site locally

#### Variant 1

Run `jekyll serve` with --watch option to test changes in "real time".

- Install [werf](http://werf.io/installation.html).
- Run:
  ```shell
  werf compose up jekyll_base --dev
  ```
- Or run with specific architecture (e.g. ARM-based Macbooks):
  ```shell
  werf compose up jekyll_base --dev --platform='linux/amd64'
  ```
- Wait (approximately 60 seconds) for the message "done in X.XXX seconds" from the `docs-en-1` and `docs-ru-1` containers.   
- Check the English version is available on [https://localhost](http://localhost), and the Russian version on [http://ru.localhost](https://ru.localhost) (add `ru.localhost` record in your `/etc/hosts` to access the Russian version of the site). 

#### Variant 2 (slower)

Run fully built 'web' image.

- Install [werf](http://werf.io/installation.html). 
- Run (add `--follow --docker-compose-command-options="-d"` if necessary):
  ```shell
  werf compose up --docker-compose-options="-f docker-compose-slow.yml" --dev
  ```
- Check the English version is available on [https://localhost](http://localhost), and the Russian version on [http://ru.localhost](https://ru.localhost) (add `ru.localhost` record in your `/etc/hosts` to access the Russian version of the site). 
