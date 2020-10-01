---
title: GitOps CLI tool
permalink: /
layout: default
---

<div class="presentation" id="presentation">
    <div class="presentation__bg" id="presentation-bg"></div>
    <div class="page__container presentation__container">
        <div class="presentation__row">
            <div class="presentation__row-item" id="presentation-title">
                <div class="presentation__subtitle">Consistent delivery tool</div>
                <h1 class="presentation__title">What you Git<br/> is what you get!</h1>
                <ul class="presentation__features">
                    <li>Git as a single source of truth.</li>
                    <li>Build. Deploy to Kubernetes. Stay in sync.</li>
                    <li>Open Source CLI tool. <a href="https://github.com/werf/werf">Written in Go.</a></li>
                </ul>
            </div>
            <div class="presentation__row-item presentation__row-item_scheme">
                {% include scheme.md %}
            </div>
        </div>
    </div>
</div>

<div class="welcome">
    <div class="page__container">
        <div class="welcome__content">
            <h1 class="welcome__title">
                It’s GitOps,<br/>
                but done <span>another way</span>!
            </h1>
            <div class="welcome__subtitle">
                Git as a single source of&nbsp;truth allows you to&nbsp;make the&nbsp;entire delivery pipeline deterministic and&nbsp;idempotent. 
                You can use it manually, from within your CI/CD system or&nbsp;as&nbsp;an&nbsp;operator (coming&nbsp;soon).
            </div>
            <div class="welcome__extra-content">
                <div class="welcome__extra-content-text">
                    <ul class="intro__list">
                        <li>
                            werf is not a SAAS, it is self-contained client-side <b>CLI tool</b>,<br/>which implements building blocks to construct CI/CD workflow for your application.
                        </li>
                        <li>
                            werf just works out of the box with minimal familiar configuration,<br/>does not need special knowledge of devops techniques to combine multiple tools<br/>and provides <b>plenty of guides</b> to quickly setup deployment of your application into Kubernetes,<br/>either for local development or production.
                        </li>
                        <li>
                            werf makes use of <code>git</code> and <code>docker</code> external dependencies to build docker images,<br/>yet it implements own <b>advanced distributed building</b> algorithm,<br/>which enables really fast pipelines due to distributed caching.
                        </li>
                        <li>
                            werf uses builtin helm binary to implement <b>helm-compatible deployment</b> process<br/>with neat extensions like clear-sighted and sharp <b>deploy process tracking</b><br/>and does not need <code>helm</code> tool to be installed.
                        </li>
                        <li>
                            werf implements smart <b>cleaning of unused docker images</b> algorithm<br/>based on Git-history of your application.
                        </li>
                        <li>
                            werf can be embedded into any existing CI/CD system (like GitLab CI/CD) <br>to construct pipelines using built-in building blocks:
                            <ul>
                                <li><code>werf converge</code>;</li>
                                <li><code>werf dismiss</code>;</li>
                                <li><code>werf cleanup</code>.</li>
                            </ul>
                        </li>
                        <li>
                            Open Source, <a href="https://github.com/werf/werf">written in Go</a>.
                        </li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</div>

<div class="page__container">
    <div class="documentation">
        <div class="documentation__image">
        </div>
        <div class="documentation__btns">
            <a href="{{ site.baseurl }}/introduction.html" target="_blank" class="page__btn page__btn_b documentation__btn">
                Introduction
            </a>
            <a href="{{ site.baseurl }}/installation.html" class="page__btn page__btn_o documentation__btn">
                Installation
            </a>
            <a href="{{ site.baseurl }}/documentation/index.html" class="page__btn page__btn_o documentation__btn">
                Documentation
            </a>
        </div>
    </div>
</div>

<div class="stats">
    <div class="page__container">
        <div class="stats__content">
            <div class="stats__title">Active development & adoption</div>
            <ul class="stats__list">
                <li class="stats__list-item">
                    <div class="stats__list-item-num">4</div>
                    <div class="stats__list-item-title">releases per week</div>
                    <div class="stats__list-item-subtitle">on average during the last year</div>
                </li>
                <li class="stats__list-item">
                    <div class="stats__list-item-num">1400</div>
                    <div class="stats__list-item-title">installations</div>
                    <div class="stats__list-item-subtitle">for large and small projects</div>
                </li>
                <li class="stats__list-item">
                    <div class="stats__list-item-num gh_counter">1470</div>
                    <div class="stats__list-item-title">stars on GitHub</div>
                    <div class="stats__list-item-subtitle">let’s make it more ;)</div>
                </li>
            </ul>
        </div>
    </div>
</div>

<div class="community">
    <div class="page__container">
        <div class="community__content">
            <div class="community__title">Friendly and rapidly growing community</div>
            <div class="community__subtitle">werf’s developers are always in contact with the community.<br/> You can reach us in Twitter and Slack.</div>
            <div class="community__btns">
                <a href="{{ site.social_links[page.lang].twitter }}" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_twitter"></span>
                    Join us in Twitter
                </a>
                <a href="#" data-open-popup="slack" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_slack"></span>
                    Join us in Slack
                </a>
            </div>
        </div>
    </div>
</div>

