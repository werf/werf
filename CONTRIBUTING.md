# Contributing to werf

werf is an Open Source project, and we are thrilled to develop and improve it in collaboration with the community.

## Before you start

For any significant change, open the issue first (or comment on an existing one) to discuss with the maintainers on what to do and how. When the solution is agreed upon, you can proceed with implementation and open a pull request.

For small changes, such as few lines bugfixes or documentation improvements, feel free to open a pull request directly.

For easy first issues, check the [good first issue](https://github.com/werf/werf/issues?q=is%3Aissue%20state%3Aopen%20label%3A%22good%20first%20issue%22) tag.

You can also check the existing [issues](https://github.com/werf/werf/issues), [discussion threads](https://github.com/werf/werf/discussions), and [documentation](https://werf.io/docs/v3/) — there may already be a discussion or solution on your topic. If not, choose the appropriate way to address the issue on [the new issue form](https://github.com/werf/werf/issues/new/choose).

## Contributing code

1. [Fork the project](https://github.com/werf/werf/fork).
2. Clone the project:

   ```shell
   git clone https://github.com/[GITHUB_USERNAME]/werf
   ```

3. Prepare an environment. To build and run werf locally, you'll need to _at least_ have the following installed:

   - [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) 2.18.0+
   - [Go](https://golang.org/doc/install) 1.11.4+

      **Important: Go Toolchain Configuration**

      To prevent unwanted modifications to the `go.mod` file (specifically the automatic addition of a `toolchain` directive), you must set the following environment variable:

      ```shell
      export GOTOOLCHAIN=local
      ```

      We recommend adding this setting to your shell profile (e.g., `.bashrc`, `.zshrc`, or equivalent) to make it persistent across sessions:

      ```shell
      echo 'export GOTOOLCHAIN=local' >> ~/.bashrc
      source ~/.bashrc
      ```

      This configuration ensures Go uses only the locally installed version of the tools instead of automatically selecting and potentially adding a toolchain version to `go.mod`, preventing unintended changes to the file.

   - [Docker](https://docs.docker.com/get-docker/)
   - [ginkgo](https://onsi.github.io/ginkgo/#installing-ginkgo) (testing framework required to run tests)
   - [go-task](https://taskfile.dev/installation/) (build tool to run common workflows)
     - To skip confirmation prompts when running tasks, use the `--yes` flag:  
         ```shell
         task --yes taskname
         ```
         Useful for automation or when you're sure the task should run without manual confirmation.
         
      To install dependencies, use the following task:

   - `task deps:install:all`

   Additionally, to build the `werf` binary, you need to install the `libbtrfs-dev` package.

4. Make your changes.
5. Run `task` with no arguments to run all essential checks (build, lint, format, quick tests, and others).

   You can also check out all available tasks with `task -l`.

6. Testing:
   1. Setup testing environment:
      ```shell
      task test:setup:environment
      ```
   2. Run tests:
      ```shell
      task test:unit
      task test:integration
      task test:e2e
      ```
   3. Cleanup testing environment:
      ```shell
      task test:cleanup:environment
      ```
   4.  Do manual testing (if needed).

7. Commit your changes. See [Conventions](#conventions) for the commit message format. The commit must be signed off (`--signoff`) as an acknowledgment of the [DCO](https://developercertificate.org/).
8. Push and open a pull request.

### Test resource cleanup

Image-building suites register `SuiteData.SetupProjectCleanup()` after their
project, stubs and temporary-directory setup. It defers local image cleanup
until content-check containers have been removed, while the
test environment and temporary directory still exist. Cleanup has its own bounded
Ginkgo context; failures to remove selected references fail the spec.

Use `SuiteData.SetContainerBackendMode(mode)` before using a backend (and before
any availability check that can skip the spec). It sets `WERF_BUILDAH_MODE` and
records the mode for cleanup. In a mixed spec, register each backend, including
an initial Docker build; a mode passed only in a child command's environment is
not visible to the hook. An ordinary spec with no recorded modes and an empty or
`docker` setting cleans Docker only. With another inherited setting and no record,
the hook conservatively checks both stores, as it does for an unrecognized mode.
Mode records do not carry over to the next spec.

Selection requires the exact `werf` project label and a project repository,
`WERF_REPO`, `WERF_FINAL_REPO`, or a repository explicitly recorded in
`SuiteData.CleanupRepositories`. Suites using command-line repository overrides
must record those exact repositories. Foreign aliases are retained; removal does
not force-delete containers or prune unselected image ancestors.

If all removals in a Docker cleanup pass succeed, cleanup inspects the
original image IDs of the project tags in a batch. It removes only IDs with
no remaining tags or repository digests. If a removal fails, cleanup skips
inspection and lists the project images again. An unexpected inspect error or
response also triggers a full listing; the reason is recorded in the spec log.
A final full listing still verifies that no project references remain.

On Linux, the test executable is re-executed as a cleanup worker before Ginkgo
starts, with a JSON request in `_WERF_TEST_CLEANUP_PROJECT`. The worker opens the
native storage once for listing, removal and verification. Non-root callers use
`buildah unshare`; the parent owns its process group and kills it on cancellation.
Docker cleanup runs for a selected Docker store when its CLI is on PATH. Selected
Buildah cleanup requires Linux and the Buildah CLI on PATH. An installed runtime
that fails to initialize rootless execution, list images or remove them is an
error, not a successful cleanup.

On the first image-in-use error for each reference, the Buildah worker reports the
reference and the storage error (including its container ID) to the spec log.
If the inventory is unchanged, it retries only when every remaining reference
returned an image-in-use error; any other error or a removal that had no effect
fails cleanup instead. The 30-second retry window starts at the first observed
conflict, not the start of cleanup. A changed inventory starts a fresh window on
the next conflict. The window is checked between passes; the enclosing spec
deadline still bounds the operation.

This cleanup does not delete remote registry data, shared base images, BuildKit
cache, or shared werf caches. In particular, project build-dir mounts under
`$WERF_HOME/shared_context/mounts/projects/<project>` and manifest cache under
`$WERF_HOME/local_cache/manifests` are not removed by this helper. They require
separate retention and ownership handling on persistent runners.
Runner shutdown or SIGKILL can prevent all test and
workflow cleanup from running. Persistent-runner operators must separately manage
cache retention and reconcile abandoned job environments only after confirming
their owning jobs have finished. Do not run global image or volume prune alongside
active tests on a shared host.


## Conventions

### Commit message

Each commit message consists of a **header** and a [**body**](#body). The header has a special format that includes a [**type**](#type), a [**scope**](#scope) and a [**subject**](#subject):

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
```

#### Type

Must be one of the following:

- **feat**: new features or capabilities that enhance the user's experience.
- **fix**: bug fixes that enhance the user's experience.
- **refactor**: a code changes that neither fixes a bug nor adds a feature.
- **docs**: updates or improvements to documentation.
- **test**: additions or corrections to tests.
- **chore**: updates that don't fit into other types.

#### Scope

Scope indicates the area of the project affected by the changes. The scope can consist of a top-level scope, which broadly categorizes the changes, and can optionally include nested scopes that provide further detail.

Supported scopes are the following:

```
# The end-user functionalities, aiming to streamline and optimize user experiences.
- giterminism
- build
  - stapel
  - dockerfile
  - docker
  - buildah
  - tagging
  - stages
- deploy
  - values
  - dependencies
  - secrets
  - templates
  - tracking
  - resource-order
  - resource-lifecycle
  - plan
- bundle
- cleanup
- host-cleanup
- run
- kube-run
- compose
- ci-env

# Maintaining, improving code quality and development workflow.
- ci
- release
- dev
- deps
```

In the header, multiple and nested scopes are separated by commas, from the broadest to the most specific: `fix(build, stapel, import): ...`.

#### Subject

The subject contains a succinct description of the change:

- use the imperative, present tense: "change" not "changed" nor "changes"
- don't capitalize the first letter
- no dot (.) at the end
- for **feat** and **fix**, describe the change from the user's side — the symptom that goes away, or what becomes possible. Release notes in `CHANGELOG.md` are generated from these subjects verbatim, and their reader has never seen the code, so internal names and the mechanism belong in the **body**. Prefer "preserve files imported into symlinked directories" over "resolve symlinks in the import target path".
- for **refactor**, **test** and **chore**, the reader is a developer, so the mechanism is the right subject.

#### Body

Just as in the **subject**, use the imperative, present tense: "change" not "changed" nor "changes".
The body should include the motivation for the change and contrast this with previous behavior.

### Branch name

Each branch name consists of a [**type**](#type), [**scope**](#scope), and a [**short-description**](#short-description):

```
<type>/<scope>/<short-description>
```

When naming branches, only the top-level scope should be used. Multiple or nested scopes are not allowed in branch names, ensuring that each branch is clearly associated with a broad area of the project.

#### Short description

A concise, hyphen-separated phrase in kebab-case that clearly describes the main focus of the branch.

### Pull request name

Each pull request title should clearly reflect the changes introduced, adhering to [**the header format** of a commit message](#commit-message), typically mirroring the main commit's text in the PR.

### Code style

See [CODESTYLE.md](CODESTYLE.md).

## Improving the documentation

The documentation is created using [Jekyll](https://jekyllrb.com/) and is located in `./docs`. Please refer to the [docs DEVELOPMENT.md](./docs/DEVELOPMENT.md) for information about the development process.
