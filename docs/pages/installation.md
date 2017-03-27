---
title: Установка
sidebar: doc_sidebar
permalink: installation.html
---

## [Install rvm and ruby](https://rvm.io/rvm/install)
## [Install docker](https://docs.docker.com/engine/installation/)
## Install dapp
### Из rubygems
```bash
gem install dapp
```
### Из исходников
```bash
git clone https://github.com/flant/dapp.git
cd dapp
gem build dapp.gemspec
gem install dapp-*.gem
```
