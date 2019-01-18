---
title: Getting started
sidebar: how_to
permalink: how_to/getting_started.html
author: Ivan Mikheykin <ivan.mikheykin@flant.com>
---

## Task Overview

In this tutorial, we will build an image of simple PHP [Symfony application](https://github.com/symfony/demo). It includes the following steps:

1. Installing required software and dependencies: `php`, `curl`, `php-sqlite` for the application,  `php-xml` and `php-zip` for the composer.
1. Setting up an `app` user and group for the web server.
1. Installing the composer from a `phar` file, which is first downloaded with `curl`.
1. Installing other project dependencies with the composer.
1. Adding the application code to the `/app` directory of the resulting image.
   This directory and all files in it should belong to `app:app`.
1. Setting up the IP address that the web server will listen to. This is done with a setting in `/opt/start.sh`, which will run when the container starts.
1. Making custom setup actions. As an illustration for the setup stage, we will write current date to `version.txt`.

Also, we will check that the application works and push the image in a docker registry. 

## Requirements

Minimal knowledge of [docker](https://www.docker.com/) and [Dockerfile directives](https://docs.docker.com/engine/reference/builder/).

## Step 1: Add a config

To implement these steps and requirements with werf we will add a special file called `werf.yaml` to the application's source code.

1. Clone the [Symfony Demo Application](https://github.com/symfony/demo) repository to get the source code:

    ```shell
    git clone https://github.com/symfony/symfony-demo.git
    cd symfony-demo
    ```

2.  In the project root directory create a `werf.yaml` with the following contents:

    <div class="tab">
      <button class="tablinks active" onclick="openTab(event, 'Ansible')">Ansible</button>
      <button class="tablinks" onclick="openTab(event, 'Shell')">Shell</button>
    </div>

    <div id="Ansible" class="tabcontent active" markdown="1">
      {% raw %}
      ```yaml
      dimg: ~
      from: ubuntu:16.04
      docker:
        WORKDIR: /app
        # Non-root user
        USER: app
        EXPOSE: "80"
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
            - locales
            - ca-certificates
        - name: "Generate en_US.UTF-8 default locale"
          locale_gen:
            name: en_US.UTF-8
            state: present
        - name: "Create non-root group for the main application"
          group:
            name: app
            state: present
            gid: 242
        - name: "Create non-root user for the main application"
          user:
            name: app
            comment: "Create non-root user for the main application"
            uid: 242
            group: app
            shell: /bin/bash
            home: /app
        - name: Add repository key
          apt_key:
            keyserver: keyserver.ubuntu.com
            id: E5267A6C
        - name: "Add PHP apt repository"
          apt_repository:
            repo: 'deb http://ppa.launchpad.net/ondrej/php/ubuntu xenial main'
            update_cache: yes
        - name: "Install PHP and modules"
          apt:
            name: "{{`{{ item }}`}}"
            state: present
            update_cache: yes
          with_items:
            - php7.2
            - php-sqlite3
            - php-xml
            - php-zip
            - php-mbstring
            - php-intl
        - name: Install composer
          get_url:
            url: https://getcomposer.org/download/1.6.5/composer.phar
            dest: /usr/local/bin/composer
            mode: a+x
        install:
        - name: "Install app deps"
          # NOTICE: Always use `composer install` command in real world environment!
          shell: composer update
          become: yes
          become_user: app
          args:
            creates: /app/vendor/
            chdir: /app/
        setup:
        - name: "Create start script"
          copy:
            content: |
              #!/bin/bash
              php bin/console server:run 0.0.0.0:8000
            dest: /app/start.sh
            owner: app
            group: app
            mode: 0755
        - raw: echo `date` > /app/version.txt
        - raw: chown app:app /app/version.txt
      git:
      - add: /
        to: /app
        owner: app
        group: app
      ```   
      {% endraw %}
    </div>

    <div id="Shell" class="tabcontent" markdown="1">
      {% raw %}
      ```yaml
      dimg: ~
      from: ubuntu:16.04
      docker:
        WORKDIR: /app
        # Non-root user
        USER: app
        EXPOSE: "80"
        ENV:
          LC_ALL: en_US.UTF-8
      shell:
        beforeInstall:
        - apt-get update
        - apt-get install -y locales ca-certificates curl software-properties-common
        - locale-gen en_US.UTF-8
        - groupadd -g 242 app
        - useradd -m -d /app -g 242 -u 242 -s /bin/bash app
        # https://askubuntu.com/posts/490910/revisions
        - LC_ALL=C.UTF-8 add-apt-repository -y ppa:ondrej/php
        - apt-get update
        - apt-get install -y php7.2 php-sqlite3 php-xml php-zip php-mbstring php-intl
        - curl -LsS https://getcomposer.org/download/1.4.1/composer.phar -o /usr/local/bin/composer
        - chmod a+x /usr/local/bin/composer
        install:
        - cd /app
        # NOTICE: Always use `composer install` command in real world environment!
        - su -c 'composer update' app
        setup:
        - "echo '#!/bin/bash' >> /app/start.sh"
        - echo 'php bin/console server:run 0.0.0.0:8000' >> /app/start.sh
        - echo `date` > /app/version.txt
        - chown app:app /app/start.sh /app/version.txt
        - chmod +x /app/start.sh
      git:
      - add: /
        to: /app
        owner: app
        group: app
      ```
      {% endraw %}
    </div>

## Step 2: Build and Run the Application

Let's build and run our first application.

1.  `cd` to the project root directory.

2.  Build an image:

    ```shell
    werf dimg build
    ```

3.  Run a container from the image:

    ```shell
    werf dimg run -d -p 8000:8000 -- /app/start.sh
    ```

4.  Check that the application runs and responds:

    ```shell
    curl localhost:8000
    ```

## Step 3: Push image into docker registry

Werf can be used to push a built image into docker-registry.

1. Run local docker-registry:

    ```shell
    docker run -d -p 5000:5000 --restart=always --name registry registry:2
    ```

2. Push image with werf using default `latest` tag:

    ```shell
    werf dimg push localhost:5000/symfony-demo
    ```

## What Can Be Improved

This example has space for further improvement:

* Set of commands for creating `start.sh` can be easily replaced with a single git command, and the file itself stored in the git repository.
* As we copy files with a git command, we can set file permissions with the same command.
* `composer install` instead of `composer update` should be used to install dependencies with versions fixed in files `composer.lock`, `package.json` and `yarn.lock`. Also, it's best to first check these files and run `composer install` when needed. To solve this problem werf have so-called `stageDependencies` directive.

These issues are further discussed in [reference]({{ site.baseurl }}/reference/build/git_directive.html).
