---
title: Сборка нескольких образов
sidebar: reference
permalink: multiple_images_for_build.html
folder: build
---

Dapp can build multiple images from one repository.

## Usage

### YAML syntax (dappfile.yml)

Вы можете объявить несколько образов, разделяя их строкой

{% raw %}
```yaml

---

```
{% endraw %}

В приведённом ниже примере будет собрано два образа: `curl_and_shell` и `nginx`:

{% raw %}
```yaml
# We set variable "BaseImage" here!
{{ $_ := set . "BaseImage" "registry.flant.com/dapp/ubuntu-dimg:10" }}

dimg: "curl_and_shell"
from: "{{ .BaseImage }}"
docker:
  WORKDIR: /app
  USER: app
  ENV:
    TERM: xterm
ansible:
  beforeInstall:
  - name: "Install Curl"
    apt:
      name: curl
      state: present
      update_cache: yes
git:
  # Including block "git application files"
  {{- include "git application files" . }}

---

dimg: "nginx"
from: "{{ .BaseImage }}"
docker:
  WORKDIR: /app
  USER: app
  ENV:
    TERM: xterm
ansible:
  beforeInstall:
  - name: "Install nginx"
    apt:
      name: nginx
      state: present
      update_cache: yes
git:
  # Including block "git application files"
  {{- include "git application files" . }}

###################################################################

{{- define "git application files" }}
  - add: /
    to: /app
    excludePaths:
    - .helm
    - .gitlab-ci.yml
    - .dappfiles
    - dappfile.yaml
    owner: app
    group: app
    stageDependencies:
      install:
      - composer.json
      - composer.lock
{{- end }}
```
{% endraw %}

В приведённом примере вы также можете заметить полезные для этого кейса вещи:

- объявление переменной, которую можно повторно использовать в разных image-ах. Делается на основании правил go templates.
- повторяюшийся блок вынесен в include.

### Ruby syntax (Dappfile)


В приведённом ниже примере будет собрано два образа: `curl_and_shell` и `nginx`:

```ruby
dimg_group do
  docker do
    from 'registry.flant.com/dapp/ubuntu-dimg:10'
    workdir '/app'
  end

  dimg_group do
    git do
      owner 'app'
      group 'app'
      add '/' do
        exclude_paths '.helm', '.gitlab-ci.yml', '.dappfiles', 'dappfile.yaml'
        to '/app'

        stage_dependencies do
          install 'composer.json', 'composer.lock'
        end
      end
    end

    dimg 'curl_and_shell' do
      chef.recipe 'my_nginx_installer'
      docker.expose 8080
      docker.user 'app'
    end

    dimg 'curl_and_shell' do
      chef.recipe 'curl_and_shell'
      docker.expose 80
      docker.expose 443
    end
  end
end
```

Также в этом примере вы можете обратить внимание на
- использование [dimg_group о котором можно почитать дополнительно](directives_images.html)
- использование [модулей chef](chef_dimod.html), позволяющих более гибко управлять образами  
