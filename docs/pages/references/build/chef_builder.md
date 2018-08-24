---
title: Chef builder
sidebar: reference
permalink: chef_builder.html
folder: advanced_build
---

**Notice! Chef builder is no more updated. Use [Ansible builder](ansible_builder.html), Luke!**

## Dappfile example

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
