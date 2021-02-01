---
title: Guides
permalink: documentation/guides.html
sidebar: documentation
---

{%- asset overview.css %}

<p>These guides provides reasonably detailed information that combines the theory and practice of development (Dev) and operation (Ops).</p>

<p>Its contents are aimed at developers seeking to acquire basic DevOps skills in organizing the continuous delivery of applications to Kubernetes. The DevOps engineers who want to solve their tasks more efficiently will also benefit from this tutorial.</p>

<p>We will gradually consider all the tasks related to developing services and implementing the CI/CD process: building, deploying, working with dependencies and assets, working with databases and in-memory storage, using e-mail and file storage, implementing autotests, and more.</p>

<p>Each version of the tutorial takes into account the specifics of the programming language/framework and includes examples of the application source code and infrastructure (IaC).</p>

<br>

<div class="overview__frameworks">
    <div class="overview__framework">
        <img src="/images/guides/nodejs.png" width="129" height="79" class="overview__framework-logo" />
        <a href="/guides/nodejs/100_basic.html" class="overview__framework-action">
            <span>Node.js</span>
            <img src="{% asset arrow.svg @path %}" class="flip-horizontal" height="12" />
        </a>
    </div>
    <div class="overview__framework">
        <img src="/images/guides/springboot.png" width="149" height="78" class="overview__framework-logo" />
        <a href="/guides/java_springboot/100_basic.html" class="overview__framework-action">
            <span>Spring Boot</span>
            <img src="{% asset arrow.svg @path %}" class="flip-horizontal" height="12" />
        </a>
    </div>
    <div class="overview__framework">
        <img src="/images/guides/django.png" width="156" height="54" class="overview__framework-logo" />
        <span class="overview__framework-action disabled">
            <span>soon...</span>
        </span>
    </div>
</div>
<div class="overview__frameworks">
    <div class="overview__framework">
        <img src="/images/guides/go.svg" width="134" height="50" class="overview__framework-logo" />
        <span class="overview__framework-action disabled">
            <span>soon...</span>
        </span>
    </div>
    <div class="overview__framework">
        <img src="/images/guides/rails.svg" width="156" height="54" class="overview__framework-logo" />
        <span class="overview__framework-action disabled">
            <span>soon...</span>
        </span>
    </div>
    <div class="overview__framework">
        <img src="/images/guides/laravel.svg" width="175" height="51" class="overview__framework-logo" />
        <span class="overview__framework-action disabled">
            <span>soon...</span>
        </span>
    </div>
</div>