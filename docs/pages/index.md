---
title: Main page
permalink: /
layout: default
---

<div class="welcome">
    <div class="page__container">
        <div class="welcome__content">
            <h1 class="welcome__title">
                Content based<br/>
                delivery tool
            </h1>
            <div class="welcome__subtitle">
                mighty and carefully crafted
            </div>
            <form action="https://www.google.com/search" class="welcome__search" method="get" name="searchform" target="_blank">
                <input name="sitesearch" type="hidden" value="werf.io">
                <input autocomplete="on" class="page__input welcome__search-input" name="q" placeholder="Search the documentation" required="required"  type="text">
                <button type="submit" class="page__icon page__icon_search welcome__search-btn"></button>
            </form>
        </div>
    </div>
</div>

<div class="page__container">
    <div class="intro">
            <div class="intro__image"></div>
            <div class="intro__content">
                <div class="intro__title">
                    A missing part of CI/CD<br/> systems
                </div>
                <div class="intro__text">
                    The main idea behind Werf is to help DevOps teams organize the workflow of applications delivery. It is designed with CI/CD systems in mind and can be used to create comfortable pipelines in gitlab, jenkins, travis, circleci, etc. It improves ease of use of git, docker, and helm and solves their problems: image naming, distributed caching, images cleanup, deployed resources tracking, etc. We consider it a new generation of high-level CI/CD tools.
                </div>
            </div>
    </div>
</div>

<div class="page__container">
    <ul class="intro-extra">
        <li class="intro-extra__item">
            <div class="intro-extra__item-title">
                Advanced image builder
            </div>
            <div class="intro-extra__item-text">
                Rebuild images incrementally basing on git history. Build images with Ansible tasks. Push the cache to the remote registry.
            </div>
        </li>
        <li class="intro-extra__item">
            <div class="intro-extra__item-title">
                Comfortable deployment
            </div>
            <div class="intro-extra__item-text">
                Full compatibility with Helm. Easy RBAC definition. Control of the deployment process with annotations. Control of resources readiness. Logging and error reporting. Easy debugging of problems without unnecessary kubectl invocations.
            </div>
        </li>
        <li class="intro-extra__item">
            <div class="intro-extra__item-title">
                Lifecycle management
            </div>
            <div class="intro-extra__item-text">
                Automatic image naming. Policy based registry cleanup. Debugging and diagnostic tools.
            </div>
        </li>
    </ul>
</div>

<div class="stats">
    <div class="page__container">
        <div class="stats__content">
            <div class="stats__title">Active development & integration</div>
            <ul class="stats__list">
                <li class="stats__list-item">
                    <div class="stats__list-item-num">4</div>
                    <div class="stats__list-item-title">releases per week</div>
                    <div class="stats__list-item-subtitle">on average for the last year</div>
                </li>
                <li class="stats__list-item">
                    <div class="stats__list-item-num">1200</div>
                    <div class="stats__list-item-title">installations</div>
                    <div class="stats__list-item-subtitle">as part of both large and small projects</div>
                </li>
                <li class="stats__list-item">
                    <div class="stats__list-item-num gh_counter">563</div>
                    <div class="stats__list-item-title">stars on GitHub</div>
                    <div class="stats__list-item-subtitle">let’s make it more ;)</div>
                </li>
            </ul>
        </div>
    </div>
</div>

<div class="features">
    <div class="page__container">
        <div class="features__title">Full delivery cycle</div>
        <ul class="features__list">
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_lifecycle"></div>
                <div class="features__list-item-title">Complete application lifecycle management</div>
                <div class="features__list-item-text">Manage image building process, deploy applications into Kubernetes and remove unused images easily.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_changes"></div>
                <div class="features__list-item-title">Rapid delivery of changes</div>
                <div class="features__list-item-text">Don’t waste time on unchanged image parts. Optimize your building process and speed up deployment.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_config"></div>
                <div class="features__list-item-title">Compact configuration file</div>
                <div class="features__list-item-text">Build multiple images with a single configuration file, share common configuration parts using go-templates.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_size"></div>
                <div class="features__list-item-title">Reduce image size</div>
                <div class="features__list-item-text">Detach source data and build tools using artifacts, mounts and stapel.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_ansible"></div>
                <div class="features__list-item-title">Build images with <span>Ansible</span></div>
                <div class="features__list-item-text">Use the powerful and popular infrastructure-as-a-code tool.</div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_debug"></div>
                <div class="features__list-item-title">Advanced tools for debugging the build process</div>
                <div class="features__list-item-text">In the process of assembling, you can access a certain stage using introspection options.</div>
            </li>
            <li class="features__list-item"></li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_kubernetes"></div>
                <div class="features__list-item-title">Comfortable deployment to <span>Kubernetes</span></div>
                <div class="features__list-item-text">Deploy to Kubernetes using standard Kubernetes package manager with interactive tracking of the deployment process and real-time logs browsing.</div>
            </li>
            <li class="features__list-item"></li>
        </ul>
    </div>
</div>

<div class="community">
    <div class="page__container">
        <div class="community__content">
            <div class="community__title">Friendly growing community</div>
            <div class="community__subtitle">Werf’s developers are always in contact with community<br/> though Slack and Telegram.</div>
            <div class="community__btns">
                <a href="https://t.me/werf_ru" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_telegram"></span>
                    Join via Telegram
                </a>
                <a href="https://cloud-native.slack.com/messages/CHY2THYUU" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_slack"></span>
                    Join via Slack
                </a>
            </div>
        </div>
    </div>
</div>

<div class="page__container">
    <div class="documentation">
        <div class="documentation__image">
        </div>
        <div class="documentation__info">
            <div class="documentation__info-title">
                Complete documentation
            </div>
            <div class="documentation__info-text">
                Documentation of werf comprises ~100 articles which include common use cases (getting started, deploy to Kubernetes, CI/CD integration and more), comprehensive description of its functions & architecture, as well as CLI, commands.
            </div>
        </div>
        <div class="documentation__btns">
            <a href="https://github.com/flant/werf" target="_blank" class="page__btn page__btn_b documentation__btn">
                Get Werf
            </a>
            <a href="{{ site.baseurl }}/how_to/" class="page__btn page__btn_o documentation__btn">
                Starters guide
            </a>
            <a href="{{ site.baseurl }}/cli/" class="page__btn page__btn_o documentation__btn">
                Explore CLI
            </a>
        </div>
    </div>
</div>
