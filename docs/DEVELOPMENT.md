<p align="center">
  <img src="https://raw.githubusercontent.com/werf/website/main/assets/images/werf-logo.svg?sanitize=true" style="max-height:100%;" height="175">
</p>
___

## Development

### Run the documentation part of the site locally

#### Variant 1

- Install [werf](http://werf.io/installation.html). 
- Run:
  ```shell
  werf compose up
  ```
- Wait (approximately 60 seconds) for the message "Server running..." from the `en_1` and `ru_1` containers.   
- Check the English version is available on [https://localhost](http://localhost), and the Russian version on [http://ru.localhost](https://ru.localhost) (add `ru.localhost` record in your `/etc/hosts` to access the Russian version of the site). 

#### Variant 2 (slower)

- Install [werf](http://werf.io/installation.html). 
- Run (add `--follow --docker-compose-command-options="-d"` if necessary):
  ```shell
  werf compose up --docker-compose-options="-f docker-compose-slow.yml" --dev
  ```
- Check the English version is available on [https://localhost](http://localhost), and the Russian version on [http://ru.localhost](https://ru.localhost) (add `ru.localhost` record in your `/etc/hosts` to access the Russian version of the site). 
