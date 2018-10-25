---
title: Building multiple images
sidebar: how_to
permalink: how_to/multi_images.html
author: Artem Kladov <artem.kladov@flant.com>
---

## Task Overview

Often a single application consists of several microservices. It can be microservices built using different technologies and programming languages. E.g., a Yii application which has logic application and worker application. The common practice is to place Dockerfiles into separate directories. So, with Dockerfile, you can't describe all components of the application in one file. As you need to describe image configuration in separate files, you can't share a part of configuration between images.

Dapp allows describing all images of a project in a one dappfile. This approach gives you more convenience.

In this article, we will build an example application — [AtSea Shop](https://github.com/dockersamples/atsea-sample-shop-app), to demonstrate how to describe multiple images in a one dappfile.

## Requirements

Installed [docker-compose](https://docs.docker.com/compose/install/).

## Building the application

The example application is the [AtSea Shop](https://github.com/dockersamples/atsea-sample-shop-app) Demonstration Application from the [Official Docker Samples repository](https://github.com/dockersamples). The application is a prototype of a small shop application consisting of several components.

It's frontend written in React and backend written in Java Spring Boot. There will be nginx reverse proxy and payment gateway added in the project to make it more real.

## Application components

### Backend

It is the `app` image. The backend container handles HTTP requests from the frontend container. The source code of the application is in the `/app` directory. It consists of Java application and ReactJS application. To build the backend image there are two artifact images (read more about artifacts [here]({{ site.baseurl }}/reference/build/artifact.html)) - `storefront` and `appserver`.

Image of the backend base on the official java image. It uses files from artifacts and doesn't need any steps for downloading packages or building.

```yaml
dimg: app
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

Builds assets. After building dapp imports assets into the `/static` directory of the `app` image. To increase the efficiency of the building `storefront` image, build instructions divided into two stages — _install_ and _setup_.

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

Builds a Java code. Dapp imports the resulting jarfile `AtSea-0.0.1-SNAPSHOT.jar` into the `/app` directory of the `app` image. To increase the efficiency of the building `appserver` image, build instructions divided into two stages — _install_ and _setup_. Also, the `/usr/share/maven/ref/repository` directory mounts with the `build_dir` directives to allow some caching (read more about mount directives [here]({{ site.baseurl }}/reference/build/mount_directive.html)).

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
dimg: reverse_proxy
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

It is the `database` image. This image base on the official image of the PostgreSQL server. Dapp adds configs and SQL file for bootstrap in this image. The backend container uses the database to store its data.

{% raw %}
```yaml
dimg: database
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
dimg: payment_gw
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

## Step 2: Сreate a dappfile

To build an application with all of its components create the following `dappfile.yml` **in the root folder** of the repository:

<details markdown="1">
<summary>The complete <b><i>dappfile.yml</i></b> file...</summary>

{% raw %}
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
dimg: app
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
dimg: reverse_proxy
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
dimg: database
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
dimg: payment_gw
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

</details>

## Step 3: Сreate SSL certificates

The NGINX in the `reverse_proxy` image listen on SSL ports and need a key and certificate.

Execute the following command in the root folder of the project to create them:

```bash
mkdir -p reverse_proxy/certs && openssl req -newkey rsa:4096 -nodes -subj "/CN=atseashop.com;" -sha256 -keyout reverse_proxy/certs/revprox_key -x509 -days 365 -out reverse_proxy/certs/revprox_cert
```

## Step 4: Build images

Execute the following command in the root folder of the project to build all images:

```bash
dapp dimg build
```

## Step 5: Tag images

Execute the following command in the root folder of the project to tag all images:

```bash
dapp dimg tag --tag-plain dapp atsea
```

## Step 6: Add docker-compose-dapp.yml file

Existing in the repo `docker-compose.yml` file assumes building some images. As we want to use already built images instead, we need to modify the `docker-compose.yml` file to use images with tag `dapp` (we built and tagged earlier).

Create the following `docker-compose-dapp.yml` file in the root folder of the project:

<details markdown="1">
<summary>The <b><i>docker-compose-dapp.yml</i></b> file...</summary>

{% raw %}
```yaml
version: "3.1"

services:
  reverse_proxy:
    image: atsea/reverse_proxy:dapp
    ports:
    - "80:80"
    - "443:443"
    networks:
    - front-tier
    - back-tier

  database:
    image: atsea/database:dapp
    user: postgres
    environment:
      POSTGRES_USER: gordonuser
      POSTGRES_DB: atsea
    ports:
    - "5432:5432"
    networks:
    - back-tier

  appserver:
    image: atsea/app:dapp
    user: gordon
    ports:
    - "8080:8080"
    - "5005:5005"
    networks:
    - front-tier
    - back-tier

  payment_gateway:
    image: atsea/payment_gw:dapp

networks:
  front-tier:
  back-tier:

networks:
  front-tier:
  back-tier:
```
{% endraw %}

</details>

## Step 7: Modify /etc/hosts file

To have an ability to open the example by the `atseashop.com` URL, add the `atseashop.com` name pointing to the address of your local interface into your `/etc/hosts` file. E.g.:

```bash
sed -ri 's/^(127.0.0.1)(\s)+/\1\2atseashop.com /' /etc/hosts
```

## Running the application

To run the application, execute the following command from the root folder of the project:

```bash
docker-compose -f docker-compose-dapp.yml up --no-build
```

> If you get an error like `ERROR: This node is not a swarm manager...` execute `docker swarm init` or `docker swarm init --advertise-addr <ip>` (where ip is the address of on the active interface).

Open the [atseashop.com](http://atseashop.com) in your browser, and you will be redirected by NGINX to `https://atseashop.com`. You will get a warning message from your browser about the security of the connection because there is a self-signed certificate used in the example. You need to add an exception to open the [https://atseashop.com](atseashop.com) page.

## Conclusions

We've described all project images in a one dappfile.

The example above shows the benefits of using dappfile:
* Using dapp allows describing all images of a project in a one dappfile. This approach gives you more convenience.
* If your project has similar images, you can share some piece of images by mounting their folder with the `build_dir` directive (read more about mounts [here]({{ site.baseurl }}/reference/build/mount_directive.html)).
* You can share artifacts between dimgs in single dappfile.
