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
                    <li>Open Source CLI tool. <a href="https://github.com/werf/werf" target="_blank">Written in Go</a>.</li>
                </ul>
                <div class="presentation__btns page__btn-group">
                    <a href="{{ site.baseurl }}/introduction.html" target="_blank" class="page__btn page__btn_b page__btn_small">
                        Introduction
                    </a>
                    <a href="{{ site.baseurl }}/documentation/quickstart.html" class="page__btn page__btn_b page__btn_small">
                        Quickstart
                    </a>
                    <a href="{{ site.baseurl }}/documentation/index.html" class="page__btn page__btn_b page__btn_small">
                        Documentation
                    </a>
                </div>
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
        </div>
    </div>
</div>

<div class="features">
    <div class="page__container">
        <ul class="features__list">
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_lifecycle"></div>
                <div class="features__list-item-title">It’s a CLI tool</div>
                <div class="features__list-item-text">
                    werf is not a SAAS, it is an Open Souce, self-contained client-side CLI&nbsp;tool. Use this single tool for <b>local development</b> or <b>embed it</b> into <b>any existing CI/CD system</b> using its main commands as building blocks:
                    <ul>
                        <li><code>werf converge</code>;</li>
                        <li><code>werf dismiss</code>;</li>
                        <li><code>werf cleanup</code>.</li>
                    </ul>
                </div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_easy"></div>
                <div class="features__list-item-title">Easy to use</div>
                <div class="features__list-item-text">
                    werf just works out of the box with a minimal configuration, does not need special knowledge of DevOps/SRE techniques to combine multiple tools and provides a <a href="{{ site.baseurl }}/documentation/guides.html"><b>plenty of guides</b></a> to quickly deploy your app into Kubernetes, either for local development or production.
                </div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_config"></div>
                <div class="features__list-item-title">Combining the best</div>
                <div class="features__list-item-text">
                    werf glues well-established software forming a transparent, <b>integrated CI/CD platform</b>. Benefit from a conveniently controlled, smooth interaction of Git, Docker, your container registry &amp; existing CI system, Helm, Kubernetes!
                </div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_kubernetes"></div>
                <div class="features__list-item-title">Distributed building</div>
                <div class="features__list-item-text">
                    werf implements an advanced builder boasting a distributed algorithm that <b>makes your pipelines really fast</b> thanks to distributed caching.
                </div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_debug"></div>
                <div class="features__list-item-title">Built-in cleaning</div>
                <div class="features__list-item-text">
                    werf implements a sophisticated algorithm to <b>clean unused Docker images</b> that is based on the Git history of your application.
                </div>
            </li>
            <li class="features__list-item">
                <div class="features__list-item-icon features__list-item-icon_helm"></div>
                <div class="features__list-item-title">Extended Helm</div>
                <div class="features__list-item-text">
                    werf uses a built-in <code>helm</code> binary to implement a Helm-compatible deployment with additional features. It doesn't require to have <code>helm</code> separately installed. It provides you a descriptive and sharp <b>deploy logging</b>, fast <b>failures detection</b> during deploy process, secrets support and other extensions making deploy process <b>robust and reliable</b>.
                </div>
            </li>
            <li class="features__list-item features__list-item_special">
                <div class="features__list-item-title">Open Source</div>
                <div class="features__list-item-description">
                    <a href="https://github.com/werf/werf" target="_blank">Written in Go</a>
                </div>
            </li>
        </ul>
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

<div class="reliability">
    <div class="page__container">
        <div class="reliability__content">
            <div class="reliability__column">
                <div class="reliability__title">
                    werf is a mature,<br>
                    reliable tool you can trust
                </div>
                <a href="{{ site.baseurl }}/installation.html#all-changes-in-werf-go-through-all-stability-channels" class="page__btn page__btn_b page__btn_small page__btn_inline">
                    read about stability channels and release process
                </a>
            </div>
            <div class="reliability__column reliability__column_image">
                <div class="reliability__image"></div>
            </div>
        </div>
    </div>
</div>

<div class="community">
    <div class="page__container">
        <div class="community__content">
            <div class="community__title">Friendly and rapidly growing community</div>
            <div class="community__subtitle">werf’s developers are always in contact with the community.<br/> You can reach us in Twitter and Discourse.</div>
            <div class="community__btns">
                <a href="{{ site.social_links[page.lang].twitter }}" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_twitter"></span>
                    Join us in Twitter
                </a>
                <a href="https://community.flant.com/c/werf/6" rel="noopener noreferrer" target="_blank" class="page__btn page__btn_w community__btn">
                    <span class="page__icon page__icon_discourse"></span>
                    Join us in Discourse
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
                Detailed documentation
            </div>
            <div class="documentation__info-text">
                werf documentation comprises over 100 articles covering typical use cases (getting started, deploying to Kubernetes, CI/CD integration, and more), CLI, commands, and providing a thorough description of functions & architecture.
            </div>
        </div>
        <div class="documentation__btns">
            <a href="{{ site.baseurl }}/introduction.html" class="page__btn page__btn_b documentation__btn">
                Introduction
            </a>
            <a href="{{ site.baseurl }}/documentation/quickstart.html" class="page__btn page__btn_o documentation__btn">
                Quickstart
            </a>
            <a href="{{ site.baseurl }}/documentation/index.html" class="page__btn page__btn_o documentation__btn">
                Documentation
            </a>
        </div>
    </div>
</div>
