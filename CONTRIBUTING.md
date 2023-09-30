# Contributing to werf

werf is an Open Source project, and we are thrilled to develop and improve it in collaboration with the community.

## Feedback

The first thing we recommend is to check the existing [issues](https://github.com/werf/werf/issues), [discussion threads](https://github.com/werf/werf/discussions), and [documentation](https://werf.io/documentation/v1.2/) - there may already be a discussion or solution on your topic. If not, choose the appropriate way to address the issue on [the new issue form](https://github.com/werf/werf/issues/new/choose).

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

### Coding Conventions

- [Effective Go](https://golang.org/doc/effective_go.html).
- [Go's commenting conventions](http://blog.golang.org/godoc-documenting-go-code).

## Improving the documentation

The documentation is made with [Jekyll](https://jekyllrb.com/) and contained within `./docs`. See the [docs DEVELOPMENT.md](./docs/DEVELOPMENT.md) for information about developing process.
