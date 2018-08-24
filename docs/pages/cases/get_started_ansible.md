---
title: First dapp application with Ansible builder
sidebar: how_to
permalink: get_started_ansible.html
---

In this tutorial we will build an image of simple PHP [Symfony application](https://github.com/symfony/demo) with dapp using [Ansible builder](ansible_builder.html).

It's recommended to have basic knowledge of Dockerfile and its [directives](https://docs.docker.com/engine/reference/builder/).

## Task Overview

Let's first take a look at what it takes to build a Symfony application. It includes the following steps:

1. Installing required software and dependencies: `php`, `curl`, `php7.0-sqlite3` for the application,  `php7.0-xml` and `php7.0-zip` for the composer.
1. Setting up an `app` user and group for the webserver.
1. Installing composer from a `phar` file, which is first downloaded with `curl`. 
1. Installing other project dependencies with composer.
1. Adding the application code to the `/demo` directory of the resulting image.
   This directory and all files in it should belong to `app:app`.
1. Setting up the IP address that the webserver will listen to. This is done with a setting in `/opt/start.sh`, which will run when the container starts.
1. Making custom setup actions. As an illustration for the setup stage, we will write current date to `version.txt`.

## Step 1: Add a Dappfile

To implement these steps and requirements with dapp we will add a special file called `dappfile.yaml` to the application's source code.

1. Clone the [Symfony Demo Application](https://github.com/symfony/demo) repository to get the source code:

    ```shell
    git clone https://github.com/symfony/symfony-demo.git
    cd symfony-demo
    ```

2.  In the project root directory create a `dappfile.yaml` with the following contents:

    {% raw %}
    ```yaml
    dimg: symfony-demo-app
    from: ubuntu:16.04
    docker:
      # app's working directory
      WORKDIR: /app
      # Non-root user 
      USER: app
      EXPOSE: '80'
      ENV:
        LC_ALL: en_US.UTF-8
    ansible:
      beforeInstall:
        - name: "Install additional packages"
          apt:
            name: "{{`{{ item }}`}}"
            state: present
            update_cache: yes
          with_items:
            - software-properties-common
            - locales
            - curl
        - name: "Generate en_US.UTF-8 default locale"
          locale_gen:
            name: en_US.UTF-8
            state: present
        # Install PHP
        - name: "Add PHP apt repository"
          apt_repository:
            repo: 'ppa:ondrej/php'
            codename: 'xenial'
            update_cache: yes
        - name: "Install PHP"
          apt:
            name: "php7.2"
            state: present
            update_cache: yes
        - name: "Create non-root group for the main application"
          group:
            name: app
            state: present
            gid: 242
        - user:
            name: app
            comment: "Create non-root user for the main application"
            uid: 242
            group: app
            shell: /bin/bash
            home: /app
        # Create /opt/start.sh to demonstrate application setup.
        # Application will listen on IP 0.0.0.0 and port 8000.
        - name: "Create start script"
          copy:
            content: |
              #!/bin/bash
              php bin/console server:run 0.0.0.0:8000
            dest: /app/start.sh
        - file:
            path: /app/start.sh
            owner: app
            group: app
            mode: 0755
      install:
        - name: "Install composer and required PHP modules"
          apt:
            name: "{{`{{ item }}`}}"
            state: present
            update_cache: yes
          with_items:
            - php-sqlite3
            - php-xml
            - php-zip
            - php-mbstring
          # install composer
        - raw: curl -LsS https://getcomposer.org/download/1.6.5/composer.phar -o /usr/local/bin/composer
        - file:
            path: /usr/local/bin/composer
            mode: "a+x"
      beforeSetup:
          # set file permissions and run composer install
        - file:
            path: /app
            state: directory
            owner: app
            group: app
            recurse: yes
        - raw: cd /app && su -c 'composer install' app
      setup:
          # use current date as the app's version
        - raw: echo `date` > /app/version.txt
        - raw: chown app:app /app/version.txt
    git:
      - add: '/'
        to: '/app'
    ```
    {% endraw %}


## Step 2: Build and Run the Application

Let's build and run our firs application.

0.  `cd` to the project root directory.

1.  Build an image:

    ```
    dapp dimg build
    ```

2.  Run a container from the image:
    
    ```
    dapp dimg run -d -p 8000:8000 -- /app/start.sh
    ```

3.  Check that the application runs and responds:
    
    ```
    curl localhost:8000
    ```

## What Can Be Improved

This example has space for further improvement:

*   Set of commands for creating `start.sh` can be easily replaced with a single git command, and the file itself stored in git repository.
*   As we copy files with a git command, we can set file permissions with the same command.
*   `composer install` is required only when `package.json` has changed. Thus it's best to first check `package.json` and run `composer install` when needed.

These issues are further discussed in [Adding source code with `git` directive](git.html).
