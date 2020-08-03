---
title: Введение в разработку и деплой приложений в Kubernetes с использованием Werf 
sidebar: documentation
permalink: documentation/guides/applications_guide/index.html
author: Igor Tsupko <igor.tsupko@flant.com>
toc: false
---

Мы разработали гайд, который призван помочь вам разобраться в том, как разворачивать комплексные приложения в Kubernetes с помощью Werf.

Мы не будем подробно описывать создание платформы (Kubernetes-кластера), но опишем ключевые аспекты. Вы должны быть готовы

- либо развернуть кластер на базе услуг какого-то cloud provider-а — на youtube достаточно обучающих роликов (например, про [Yandex Cloud](https://www.youtube.com/watch?v=Ngadh9T2dOI))
- либо установить кластер своими руками
- либо вам нужно будет найти тех, кто установит кластер за вас
- или вы можете установить minikube локально

Гайд рассчитан на классический Kubernetes-кластер, но должно быть несложно адаптировать его под кастомизированные сборки. 
 
В гайде последовательно рассматриваются все основные задачи, возникающие при разработке сервисов: сборка, деплой, работа с зависимостями и ассетами, работа с базами данных и in-memory хранилищами, работа с почтой, файловыми хранилищами, организация автотестов и другие. В каждом из вариантов гайда учтена специфика языка/фреймворка и приложены примеры исходных кодов приложения и инфраструктуры. 

<h2>Выберите фреймворк</h2>

<div class="nav-btn-list">
<!--    <a href="gitlab_rails/000_task.html" class="nav-btn">Ruby On Rails и GitLab</a> -->
    <a href="gitlab_nodejs/000_task.html" class="nav-btn">NodeJS и GitLab</a>
    <a href="gitlab_python_django/000_task.html" class="nav-btn">Python: Django и GitLab</a>
    <a href="gitlab_java_springboot/000_task.html" class="nav-btn">Java: Springboot и GitLab</a>
</div>
