---
title: Сборка нескольких образов (#TODO)
sidebar: doc_sidebar
permalink: multiple_images_for_build.html
folder: build
---

Поддерживается и chef и ansible сборщиком

* Problem: one repo has multiple kinds of app.
* Example: yii app has logic application and worker application.
* Example 2: several microservices in one repo.
* Example 3: you may want to pack all static failes into scratch volume to mount it into nginx container.
* Example 4: debug and release configurations (release may depend only on changes of version.c file).
* Example 5: memcached
* More on Dappfile syntax
* directives: dimg_group
