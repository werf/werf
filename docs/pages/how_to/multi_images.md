---
title: Building multiple images
sidebar: how_to
permalink: how_to/multi_images.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

Often a single application consists of several microservices. It can be microservices built using different technologies and programming languages. E.g., a Yii application which has logic application and worker application. The common practice is to place Dockerfiles into separate directories. So, with Dockerfile, you can't describe all components of the application in one file. As you need to describe image configuration in separate files, you can't share a part of configuration between images.

Werf allows describing all images of a project in a one config. This approach gives you more convenience.

In this article, we will build an example application — [AtSea Shop](https://github.com/dockersamples/atsea-sample-shop-app), to demonstrate how to describe multiple images in a one config.

## Requirements

* Installed [multiwerf](https://github.com/flant/multiwerf) on the host system.

### Select werf version

This command should be run prior running any werf command in your shell session:

```shell
source <(multiwerf use 1.0 beta)
```

## Building the application

The example application is the [AtSea Shop](https://github.com/dockersamples/atsea-sample-shop-app) Demonstration Application from the [Official Docker Samples repository](https://github.com/dockersamples). The application is a prototype of a small shop application consisting of several components.

It's frontend written in React and backend written in Java Spring Boot. There will be nginx reverse proxy and payment gateway added in the project to make it more real.

## Application components

### Backend

It is the `app` image. The backend container handles HTTP requests from the frontend container. The source code of the application is in the `/app` directory. It consists of Java application and ReactJS application. To build the backend image there are two artifact images (read more about artifacts [here]({{ site.baseurl }}/reference/build/artifact.html)) - `storefront` and `appserver`.

Image of the backend base on the official java image. It uses files from artifacts and doesn't need any steps for downloading packages or building.

```yaml
image: app
from: java:8-jdk-alpine
docker:
  ENTRYPOINT: ["java", "-jar", "/app/AtSea-0.0.1-SNAPSHOT.jar"]
  CMD: ["--spring.profiles.active=postgres"]
shell:
  beforeInstall:
  - mkdir /app
  - adduser -Dh /home/gordon gordon
import:
- artifact: storefront
  add: /usr/src/atsea/app/react-app/build
  to: /static
  after: install
- artifact: appserver
  add: /usr/src/atsea/target/AtSea-0.0.1-SNAPSHOT.jar
  to: /app/AtSea-0.0.1-SNAPSHOT.jar
  after: install
```

#### Storefront artifact

Builds assets. After building werf imports assets into the `/static` directory of the `app` image. To increase the efficiency of the building `storefront` image, build instructions divided into two stages — _install_ and _setup_.

```yaml
artifact: storefront
from: node:latest
git:
- add: /app/react-app
  to: /usr/src/atsea/app/react-app
  stageDependencies:
    install:
    - package.json
    setup:
    - src
    - public
shell:
  install:
  - cd /usr/src/atsea/app/react-app
  - npm install
  setup:
  - cd /usr/src/atsea/app/react-app
  - npm run build
```

#### Appserver artifact

Builds a Java code. Werf imports the resulting jarfile `AtSea-0.0.1-SNAPSHOT.jar` into the `/app` directory of the `app` image. To increase the efficiency of the building `appserver` image, build instructions divided into two stages — _install_ and _setup_. Also, the `/usr/share/maven/ref/repository` directory mounts with the `build_dir` directives to allow some caching (read more about mount directives [here]({{ site.baseurl }}/reference/build/mount_directive.html)).

```yaml
artifact: appserver
from: maven:latest
mount:
- from: build_dir
  to: /usr/share/maven/ref/repository
git:
- add: /app
  to: /usr/src/atsea
  stageDependencies:
    install:
    - pom.xml
    setup:
    - src
shell:
  install:
  - cd /usr/src/atsea
  - mvn -B -f pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:go-offline
  setup:
  - cd /usr/src/atsea
  - mvn -B -s /usr/share/maven/ref/settings-docker.xml package -DskipTests
```

### Frontend

It is the `reverse_proxy` image. This image base on the official image of the [NGINX](https://www.nginx.com) server. It acts as a frontend and is configured as a reverse proxy. The frontend container handles all incoming traffic, cache it and pass requests to the backend container.

{% raw %}
```yaml
image: reverse_proxy
from: nginx:alpine
ansible:
  install:
  - name: "Copy nginx.conf"
    copy:
      content: |
{{ .Files.Get "reverse_proxy/nginx.conf" | indent 8 }}
      dest: /etc/nginx/nginx.conf
  - name: "Copy SSL certificates"
    file:
      path: /run/secrets
      state: directory
      owner: nginx
  - copy:
      content: |
{{ .Files.Get "reverse_proxy/certs/revprox_cert" | indent 8 }}
      dest: /run/secrets/revprox_cert
  - copy:
      content: |
{{ .Files.Get "reverse_proxy/certs/revprox_key" | indent 8 }}
      dest: /run/secrets/revprox_key
```
{% endraw %}

### Database

It is the `database` image. This image base on the official image of the PostgreSQL server. Werf adds configs and SQL file for bootstrap in this image. The backend container uses the database to store its data.

{% raw %}
```yaml
image: database
from: postgres:11
docker:
  ENV:
    POSTGRES_USER: gordonuser
    POSTGRES_DB: atsea
ansible:
  install:
  - raw: mkdir -p /images/
  - name: "Copy DB configs"
    copy:
      content: |
{{ .Files.Get "database/pg_hba.conf" | indent 8 }}
      dest: /usr/share/postgresql/11/pg_hba.conf
  - copy:
      content: |
{{ .Files.Get "database/postgresql.conf" | indent 8 }}
      dest:  /usr/share/postgresql/11/postgresql.conf
git:
- add: /database/docker-entrypoint-initdb.d/
  to:  /docker-entrypoint-initdb.d/
```
{% endraw %}

### Payment gateway

It is the `payment_gw` image. This image is an example of the payment gateway application. It does nothing except infinitely writes messages to stdout. Payment gateway acts as another component of the application.

{% raw %}
```yaml
image: payment_gw
from: alpine
docker:
  CMD: ["/home/payment/process.sh"]
ansible:
  beforeInstall:
  - name: "Create payment user"
    user:
      name: payment
      comment: "Payment user"
      shell: /bin/sh
      home: /home/payment
  - file:
      path: /run/secrets
      state: directory
      owner: payment
  - copy:
      content: |
        production
      dest: /run/secrets/payment_token
git:
- add: /payment_gateway/process.sh
  to: /home/payment/process.sh
  owner: payment
```
{% endraw %}

## Step 1: Clone the application repository

Clone the [AtSea Shop](https://github.com/dockersamples/atsea-sample-shop-app) repository:

```bash
git clone https://github.com/dockersamples/atsea-sample-shop-app.git
```

## Step 2: Create a config

To build an application with all of its components create the following `werf.yaml` **in the root folder** of the repository:

<div class="details active">
<a href="javascript:0;" class="details__summary">The complete <i>werf.yaml</i> file...</a>
<div class="details__content" markdown="1">

{% raw %}
```yaml
project: atsea-shop
configVersion: 1
---

artifact: storefront
from: node:latest
git:
- add: /app/react-app
  to: /usr/src/atsea/app/react-app
  stageDependencies:
    install:
    - package.json
    setup:
    - src
    - public
shell:
  install:
  - cd /usr/src/atsea/app/react-app
  - npm install
  setup:
  - cd /usr/src/atsea/app/react-app
  - npm run build
---
artifact: appserver
from: maven:latest
mount:
- from: build_dir
  to: /usr/share/maven/ref/repository
git:
- add: /app
  to: /usr/src/atsea
  stageDependencies:
    install:
    - pom.xml
    setup:
    - src
shell:
  install:
  - cd /usr/src/atsea
  - mvn -B -f pom.xml -s /usr/share/maven/ref/settings-docker.xml dependency:go-offline
  setup:
  - cd /usr/src/atsea
  - mvn -B -s /usr/share/maven/ref/settings-docker.xml package -DskipTests
---
image: app
from: java:8-jdk-alpine
docker:
  ENTRYPOINT: ["java", "-jar", "/app/AtSea-0.0.1-SNAPSHOT.jar"]
  CMD: ["--spring.profiles.active=postgres"]
shell:
  beforeInstall:
  - mkdir /app
  - adduser -Dh /home/gordon gordon
import:
- artifact: storefront
  add: /usr/src/atsea/app/react-app/build
  to: /static
  after: install
- artifact: appserver
  add: /usr/src/atsea/target/AtSea-0.0.1-SNAPSHOT.jar
  to: /app/AtSea-0.0.1-SNAPSHOT.jar
  after: install
---
image: reverse_proxy
from: nginx:alpine
ansible:
  install:
  - name: "Copy nginx.conf"
    copy:
      content: |
{{ .Files.Get "reverse_proxy/nginx.conf" | indent 8 }}
      dest: /etc/nginx/nginx.conf
  - name: "Copy SSL certificates"
    file:
      path: /run/secrets
      state: directory
      owner: nginx
  - copy:
      content: |
{{ .Files.Get "reverse_proxy/certs/revprox_cert" | indent 8 }}
      dest: /run/secrets/revprox_cert
  - copy:
      content: |
{{ .Files.Get "reverse_proxy/certs/revprox_key" | indent 8 }}
      dest: /run/secrets/revprox_key
---
image: database
from: postgres:11
docker:
  ENV:
    POSTGRES_USER: gordonuser
    POSTGRES_DB: atsea
ansible:
  install:
  - raw: mkdir -p /images/
  - name: "Copy DB configs"
    copy:
      content: |
{{ .Files.Get "database/pg_hba.conf" | indent 8 }}
      dest: /usr/share/postgresql/11/pg_hba.conf
  - copy:
      content: |
{{ .Files.Get "database/postgresql.conf" | indent 8 }}
      dest:  /usr/share/postgresql/11/postgresql.conf
git:
- add: /database/docker-entrypoint-initdb.d/
  to:  /docker-entrypoint-initdb.d/
---
image: payment_gw
from: alpine
docker:
  CMD: ["/home/payment/process.sh"]
ansible:
  beforeInstall:
  - name: "Create payment user"
    user:
      name: payment
      comment: "Payment user"
      shell: /bin/sh
      home: /home/payment
  - file:
      path: /run/secrets
      state: directory
      owner: payment
  - copy:
      content: |
        production
      dest: /run/secrets/payment_token
git:
- add: /payment_gateway/process.sh
  to: /home/payment/process.sh
  owner: payment
```
{% endraw %}

</div>
</div>

## Step 3: Create SSL certificates

The NGINX in the `reverse_proxy` image listen on SSL ports and need a key and certificate.

Execute the following command in the root folder of the project to create them:

```bash
mkdir -p reverse_proxy/certs && openssl req -newkey rsa:4096 -nodes -subj "/CN=atseashop.com;" -sha256 -keyout reverse_proxy/certs/revprox_key -x509 -days 365 -out reverse_proxy/certs/revprox_cert
```

## Step 4: Build images

Execute the following command in the root folder of the project to build all images:

```bash
werf build --stages-storage :local
```

## Step 5: Modify /etc/hosts file

To have an ability to open the example by the `http://atseashop.com` URL, add the `atseashop.com` name pointing to the address of your local interface into your `/etc/hosts` file. E.g.:

```bash
sudo sed -ri 's/^(127.0.0.1)(\s)+/\1\2atseashop.com /' /etc/hosts
```

## Step 6: Run the application

To run the application images, execute the following commands from the root folder of the project:

```bash
werf run --stages-storage :local --docker-options="-d --rm --name payment_gw" payment_gw  &&
werf run --stages-storage :local --docker-options="-d --rm --name database -p 5432:5432" database &&
werf run --stages-storage :local --docker-options="-d --rm --name app -p 8080:8080 --link database:database" app &&
werf run --stages-storage :local --docker-options="-d --rm --name reverse_proxy -p 80:80 -p 443:443 --link app:appserver" reverse_proxy
```

Check that all containers are running, by executing:
```bash
docker ps
```

You should see running containers with names: reverse_proxy, app, database, payment_gw and registry.

Wait for about 30 seconds or a bit more for all the containers to be ready, then open the [atseashop.com](http://atseashop.com) in your browser. You will be redirected by NGINX to `https://atseashop.com` and you will get a warning message from your browser about the security of the connection, because there is a self-signed certificate used in the example. You need to add an exception to open the [https://atseashop.com](atseashop.com) page.

## Stopping the application

To stop application containers execute the following command:

```bash
docker stop reverse_proxy app database payment_gw
```

## Conclusions

We've described all project images in a one config.

The example above shows the benefits:
* If your project has similar images, you can share some piece of images by mounting their folder with the `build_dir` directive (read more about mounts [here]({{ site.baseurl }}/reference/build/mount_directive.html)).
* You can share artifacts between images in single config.
* Common templates can be used in single config to describe configuration of multiple images.
