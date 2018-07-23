---
title: Chef builder
sidebar: reference
permalink: build_chef.html
folder: advanced_build
---

**Notice! Chef builder is no more updated. Use [Ansible builder](build_yaml.html), Luke!**

## File structure

- `/Dappfile` - Repository should contain dappfile with build instructions.
- `/.dapp_chef/` - folder for chef recipes, config files for your software and misc 
- `/.dapp_chef/recipes/` - folder for chef recipes
- `/.dapp_chef/files/` - folder for config files for your software and misc
- `/.dapp_chef/chefignore` - chef ignore file
- `/.helm/secret/` - folder for [dapp secret files](kube_secret.html)

## dappfile syntax

Example of configuration:

```
dimg_group do
  docker do
    from 'registry.flant.com/dapp/ubuntu-dimg:8'
    workdir '/app'
  end
  
  chef do
    cookbook 'apt'
    attributes['ruby-version'] = "2.3.4"
    recipe 'main'
  end
  
  git do
    owner 'app'
    group 'app'
    add '/' do
      exclude_paths 'public/assets', 'vendor', '.helm'
      to '/app'

      stage_dependencies do
        install 'package.json', 'Bowerfile', 'Gemfile.lock', 'app/assets/*'
      end
    end
end
```

We use [docker directives](docker_directives.html) here to set base image and work dir.

Chef directive
```
  chef do
    cookbook 'apt'
    attributes['ruby-version'] = "2.3.4"
    recipe 'main'
  end
```
tells builder to run chef recipes on [each build stage](stages.html):

- `.dapp_chef/recipes/before_install/main.rb`
- `.dapp_chef/recipes/install/main.rb`
- `.dapp_chef/recipes/after_install/main.rb`
- `.dapp_chef/recipes/before_setup/main.rb`
- `.dapp_chef/recipes/setup/main.rb`
- `.dapp_chef/recipes/after_setup/main.rb`

## Additional info

- Use [chef directives](chef_directives.html)
- Run [shell commands](shell.html)
- Build [multiple images](multiple_images_for_build.html)
- [Mount directories](mount_directives.html)
- Use [chef dimod](chef_dimod.html)