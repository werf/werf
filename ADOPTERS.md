# Adopters

Here you can find a list of organisations that use werf in their 
environments. The list has been created based on publicly available 
information and/or information from representatives of organisations
who expressed interest in sharing these details on GitHub.

## werf adopters

| Org/project | Contact (if any) | Description (how they use werf, whether it's in production, etc.) |
| ------------ | ------- | ------------------ |
| [Flant](https://flant.ru/) | @ilya-lesikov | A DevOps service and solution provider uses werf in production: a) for implementing the software delivery process across dozens of various customers, b) as part of the [Deckhouse Kubernetes Platform](https://deckhouse.io/) (the foundation for Delivery Kit and Marketplace). |
| [Palark GmbH](https://palark.com/) | @shurup | A DevOps and SRE agency uses werf to implement CI/CD for various customers and for internal software development. |
| [KMD](https://www.kmd.net/) | - | The KMD's WorkZone EMI product uses werf for container deployment and packages, according to [this public documentation](https://docs.workzone.kmd.net/2026_2/en-us/Content/WZCtnr_Guide/Standard_werf_packages.htm) (as updated in July 2026). |
| [Condo](https://github.com/open-condo-software/condo) | - | An Open Source property management SaaS uses werf to build and deploy the code across all environments, including the production one (see [werf.yaml](https://github.com/open-condo-software/condo/blob/main/werf.yaml) and GitHub Actions [configuration](https://github.com/open-condo-software/condo/blob/main/.github/workflows/deploy_production.yaml)). |
| [GBH](https://gbh.tech/) | - | A technology consulting company uses werf for GitHub Action-based deployments to AWS EKS clusters (see [werf-deployment-action](https://github.com/gbh-tech/werf-deployment-action)) and in its internal tooling (see [envi](https://github.com/gbh-tech/envi)). |
| [A-listware LTD](https://a-listware.com/) | - | A software development company works with werf in CI/CD while offering managed IT services, according to [this website page](https://a-listware.com/services/managed-it). |
| [Movavi](https://movavi.com/) | - | A multimedia software developer lists werf as one of the required skills for their DevOps Team Lead role, according to [this vacancy](https://job.movavi.com/vacancies/devops-team-lead/) (available in September 2026). |
| [Skyro](https://www.skyro.io/) | - | A financial technologies group uses werf to deploy and manage services in AWS EKS, according to [this vacancy](https://www.linkedin.com/jobs/view/middle+-golang-developer-fluent-in-english-russian-at-skyro-4191533455/) (published in June 2025). |
| [SOFTSWISS](https://www.softswiss.com/) | - | A B2B iGaming provider lists werf as one of the nice-to-have skills for their Senior DevOps/System Engineer role, according to [this vacancy](https://www.linkedin.com/jobs/view/devops-system-engineer-platform-kubernetes-%E2%80%93-senior-at-softswiss-4410884220/) (published in May 2026). |
| [Wisebits Group](https://wisebits.com/) | - | An international IT holding lists werf as one of the required skills for their DevOps Engineer role, according to [this vacancy](https://bebee.com/cy/jobs/devops-engineer-wisebits-group-limassol--fj-2296208850) (published in August 2026). |
| [Koch](https://www.kochinc.com/) | - | An American multinational conglomerate corporation lists werf as one of the nice-to-have skills for their Lead Linux Engineer role, according to [this vacancy](https://www.linkedin.com/jobs/view/lead-linux-engineer-at-koch-4416795567/) (published in July 2026). |
| [3Play Media](https://www.3playmedia.com/) | - | A video accessibility and localization solution provider manages product deployments with werf, according to [this vacancy](https://www.linkedin.com/jobs/view/principal-front-end-engineer-growth-at-3play-media-4423041153/) (published in July 2026). |
| [FOLIO](https://folio.org/) | - | An Open Source platform for libraries migrated all its Docker builds to werf in January 2026, according to [this public ticket](https://folio-org.atlassian.net/browse/RANCHER-2731). |
| [Innowise Group](https://innowise.com/) | - | An international full-cycle software development company [mentions](https://innowise.com/hire-developers/devops/) werf in the core DevOps technologies it works with. |
| [Seznam](https://www.seznam.cz/) | @hamskerpir | A Czech Internet company uses werf for CI/CD in staging and production, with more details provided in [this article](https://dev.to/laplacedaemon/how-werf-streamlined-the-transition-to-kubernetes-and-accelerated-our-cicd-4f19). |
| [Raiffeisenbank Russia](https://www.raiffeisen.ru/) | - | A bank uses werf internally as part of its "builder for creating pipelines in GitLab", according to publicly available configurations in [The way of CI/CD project repo](https://github.com/Raiffeisen-DGTL/The-Way-of-CICD-Open-Source-Edition). |

## Add your use case

If your company/organisation/project relies on werf and doesn't mind
sharing  this fact with others, please open a relevant PR in this repo
that will add a line to the table above. Such a contribution will be
VERY helpful for the werf project on various levels, from motivating
the core developers to keep their work going to formally demonstrating
the actual usage and maturity of werf to a broader engineering
community!

