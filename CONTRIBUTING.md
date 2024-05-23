# Contributing to werf

werf is an Open Source project, and we are thrilled to develop and improve it in collaboration with the community.

## Feedback

The first thing we recommend is to check the existing [issues](https://github.com/werf/werf/issues), [discussion threads](https://github.com/werf/werf/discussions), and [documentation](https://werf.io/docs/v2/) - there may already be a discussion or solution on your topic. If not, choose the appropriate way to address the issue on [the new issue form](https://github.com/werf/werf/issues/new/choose).

## Contributing code

1. [Fork the project](https://github.com/werf/werf/fork).
2. Clone the project:
     
   ```shell
   git clone https://github.com/[GITHUB_USERNAME]/werf
   ```

3. Prepare an environment. To build and run werf locally, you'll need to _at least_ have the following installed:

   - [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) 2.18.0+
   - [Go](https://golang.org/doc/install) 1.11.4+
   - [Docker](https://docs.docker.com/get-docker/)
   - [go-task](https://taskfile.dev/installation/) (build tool to run common workflows)
   - [ginkgo](https://onsi.github.io/ginkgo/#installing-ginkgo) (testing framework required to run tests)

4. Make changes.
5. Build werf:

   ```shell
   task build # The built werf binary will be available in the bin directory.
   ```
   
6. Do manual testing.
7. Run tests:

   ```shell
   task test:unit
   task test:integration
   task test:e2e
   ```
   
8. Format and lint your code: 

   ```shell
   task format lint
   ```
   
9. Commit changes:

   - Follow [The Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/).
   - Sign off every commit you contributed as an acknowledgment of the [DCO](https://developercertificate.org/).

10. Push commits.
11. Create a pull request.

## Conventions

### Commit message

Each commit message consists of a **header** and a [**body**](#body). The header has a special format that includes a [**type**](#type), a [**scope**](#scope) and a [**subject**](#subject):

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
```

**Examples:**

  - _feat(api): bump resource version to v1beta1_
  - _feat(images): implement ImageLost phase_
  - _feat(images, vi): add PVC as a storage_
  - _fix(vm, vmmodel): fix unsupported type of model_
  - _refactor(core): rename 3rd party resources_
  - _docs(module): describe how to install module_
  - _chore(core): use alt linux as base image_
  - _chore: update go dependencies_

#### Type

Must be one of the following:

* **feat**: new features or capabilities that enhance the user's experience.
* **fix**: bug fixes that enhance the user's experience.
* **refactor**: a code changes that neither fixes a bug nor adds a feature.
* **docs**: updates or improvements to documentation.
* **test**: additions or corrections to tests.
* **chore**: updates that don't fit into other types.

#### Scope

Scope indicates the area of the project affected by the changes. The scope can consist of a top-level scope, which broadly categorizes the changes, and can optionally include nested scopes that provide further detail.

Supported scopes are the following:

  ```
  # The end-user functionalities, aiming to streamline and optimize user experiences.
  # NOTE! The api scope should be omitted if a more specific sub-scope is used.
  - api
    - vm
      - vmop
      - vmbda
      - vmcpu
      - vmcpureq
      - vmmodel
      - vmip
      - vmipl
    - disks
      - vd
    - images
      - vi
      - cvi

  # Core mechanisms and low-level system functionalities.
  - core
    - api-service
    - vmi-router
    - kubevirt
    - kube-api-rewriter
    - cdi
    - dvcr

  # Integration with the Deckhouse.
  - module

  # User metrics, alerts, dashboards, and logs that provide insights into system performance and health.
  - observability

  # Maintaining, improving code quality and development workflow. 
  - ci
  - lint
  - format
  - gen
  - dev
  ```

#### Subject

The subject contains a succinct description of the change:

  - use the imperative, present tense: "change" not "changed" nor "changes"
  - don't capitalize the first letter
  - no dot (.) at the end

#### Body

Just as in the **subject**, use the imperative, present tense: "change" not "changed" nor "changes".
The body should include the motivation for the change and contrast this with previous behavior.

### Branch name

Each branch name consists of a [**type**](#type), [**scope**](#scope), and a [**short-description**](#short-description):

```
<type>/<scope>/<short-description>
```

When naming branches, only the top-level scope should be used. Multiple or nested scopes are not allowed in branch names, ensuring that each branch is clearly associated with a broad area of the project.

**Examples:**

  - _feat/disks/support-new-image-source_
  - _chore/ci/speed-up-builds_

#### Short description

A concise, hyphen-separated phrase in kebab-case that clearly describes the main focus of the branch.

### Pull request name

Each pull request title should clearly reflect the changes introduced, adhering to [**the header format** of a commit message](#commit-message), typically mirroring the main commit's text in the PR.

**Examples**

  - _feat(vm): add live migration capability_
  - _docs(api): update REST API documentation for clarity_
    
### Coding Conventions

- [Effective Go](https://golang.org/doc/effective_go.html).
- [Go's commenting conventions](http://blog.golang.org/godoc-documenting-go-code).

## Improving the documentation

The documentation is made with [Jekyll](https://jekyllrb.com/) and contained within `./docs`. See the [docs DEVELOPMENT.md](./docs/DEVELOPMENT.md) for information about developing process.
