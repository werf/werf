---
title: Overview
permalink: usage/project_configuration/overview.html
---

werf follows the principles of the IaC (Infrastructure as Code) approach and encourages the user to store the project delivery configuration along with the application code in Git and to use external dependencies responsibly. This is accomplished by a mechanism called [giterminism]({{ "usage/project_configuration/giterminism.html" | true_relative_url }}).

A typical project configuration includes several files:

- werf.yaml;
- one or several Dockerfiles;
- Helm chart.

## werf.yaml

werf.yaml is the main configuration file of a project in werf. Its primary purpose is to bind build and deploy instructions.

### Build instructions

These instructions are defined for each application component. They can be in two formats:
- Dockerfiles that describe the project images.
- Stapel (an alternative assembly syntax).

> Refer to the ["Build"]({{ "usage/build/overview.html" | true_relative_url }}) section of the documentation for more details on the assembly configuration.

### Deploy instructions

These instructions are defined for the entire application (and all deployment environments) and should take the form of a Helm Chart.

> Refer to the ["Deploy"]({{ "usage/deploy/overview.html" | true_relative_url }}) section of the documentation for details on the deployment configuration.

## Example of a typical project configuration

```yaml
# werf.yaml
project: app
configVersion: 1
---
image: backend
context: backend
dockerfile: Dockerfile
---
image: frontend
context: frontend
dockerfile: Dockerfile
```

```shell
$ tree -a
.
├── .helm
│   ├── templates
│   │   ├── NOTES.txt
│   │   ├── _helpers.tpl
│   │   ├── deployment.yaml
│   │   ├── hpa.yaml
│   │   ├── ingress.yaml
│   │   ├── service.yaml
│   │   └── serviceaccount.yaml
│   └── values.yaml
├── backend
│   ├── Dockerfile
│   └── ...
├── frontend
│   ├── Dockerfile
│   └── ...
└── werf.yaml
```
