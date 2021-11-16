# Changelog

### [1.2.40](https://www.github.com/werf/werf/compare/v1.2.39...v1.2.40) (2021-11-16)


### Bug Fixes

* **bundles:** fix werf-bundle-publish command error when --tag contains underscore chars ([03e8f88](https://www.github.com/werf/werf/commit/03e8f88ffc05a31d557df11a02d4a1d6df1300e3))

### [1.2.39](https://www.github.com/werf/werf/compare/v1.2.38...v1.2.39) (2021-11-09)


### Bug Fixes

* **buildah:** support --build-arg arguments defined in the werf.yaml ([8a2081e](https://www.github.com/werf/werf/commit/8a2081ef4499bba507fd66198370353460f04992))

### [1.2.38](https://www.github.com/werf/werf/compare/v1.2.37...v1.2.38) (2021-11-02)


### Features

* **bundles:** update helm to 3.7.1, provide compatibility with old published bundles ([9dc215c](https://www.github.com/werf/werf/commit/9dc215c9b6e5d6b263ec31346809a7dc5325e81e))

### [1.2.37](https://www.github.com/werf/werf/compare/v1.2.36...v1.2.37) (2021-10-29)


### Features

* **helm:** werf-helm-* commands now fully support --post-renderer param ([eb8208e](https://www.github.com/werf/werf/commit/eb8208efbcae5e0b9779a79551f43c22e4e3fc0c))

### [1.2.36](https://www.github.com/werf/werf/compare/v1.2.35...v1.2.36) (2021-10-21)


### Bug Fixes

* **deploy:** fix dismiss command fails with "panic: close of closed channel" ([b9b064c](https://www.github.com/werf/werf/commit/b9b064c066f48b75e5852c7401589396e0b85a0a))

### [1.2.35](https://www.github.com/werf/werf/compare/v1.2.34...v1.2.35) (2021-10-20)


### Features

* **buildah:** communication with insecure registries ([e0502c2](https://www.github.com/werf/werf/commit/e0502c28b37939bc86362f8fa644f007a71b4421))


### Bug Fixes

* **cleanup:** panic: runtime error: invalid memory address or nil pointer dereference ([9024c5c](https://www.github.com/werf/werf/commit/9024c5c393ae07901766fa46065e4ce9ea2239e3))

### [1.2.34](https://www.github.com/werf/werf/compare/v1.2.33...v1.2.34) (2021-10-19)


### Bug Fixes

* **stapel:** add patch to update ssl certs in the old stapel image ([76fb6c8](https://www.github.com/werf/werf/commit/76fb6c81aecba5fdb6e6cba6fbefba4c870b2329))
* **stapel:** build omnibus packages with /.werf/stapel toolchain ([cc86423](https://www.github.com/werf/werf/commit/cc864232acba0dd8b4c7d37922c5f2c646b5f991))

### [1.2.33](https://www.github.com/werf/werf/compare/v1.2.32...v1.2.33) (2021-10-19)


### Bug Fixes

* **build:** commit object or packfile not found after git GC ([b7a6bba](https://www.github.com/werf/werf/commit/b7a6bba1348016df9075ca7d75d99597f2e8b45c))

### [1.2.32](https://www.github.com/werf/werf/compare/v1.2.31...v1.2.32) (2021-10-18)


### Bug Fixes

* **custom tags:** --use-custom-tag with an image name not work properly ([89807af](https://www.github.com/werf/werf/commit/89807afea5b8d379948877c9e271aa3b1b3ec31a))

### [1.2.31](https://www.github.com/werf/werf/compare/v1.2.30...v1.2.31) (2021-10-18)


### Bug Fixes

* **stapel:** disable python 2 deprecation warning in ansible builder ([00d9834](https://www.github.com/werf/werf/commit/00d9834be5cfdaa047872704ada598fb9e2053c9))
* **stapel:** rebuild stapel to fix expired certificates ([1da0d43](https://www.github.com/werf/werf/commit/1da0d432e624c1b25e69d3ee2ac04ba9f38cd962))

### [1.2.30](https://www.github.com/werf/werf/compare/v1.2.29...v1.2.30) (2021-10-14)


### Bug Fixes

* **deploy:** WERF_SET_DOCKER_CONFIG_VALUE not working ([b850301](https://www.github.com/werf/werf/commit/b850301d16a2dd9c05c223eebe745de95cdf9db8))

### [1.2.29](https://www.github.com/werf/werf/compare/v1.2.28...v1.2.29) (2021-10-12)


### Bug Fixes

* **deploy:** possible fix for hanging werf-dismiss ([4ea7915](https://www.github.com/werf/werf/commit/4ea79155e6f070af2909db5602db32c110201ef9))
* WERF_SET_DOCKER_CONFIG_VALUE env variable collision with --set param ([30177b4](https://www.github.com/werf/werf/commit/30177b4f4210cbde6e531091f396f7faa187f808))

### [1.2.28](https://www.github.com/werf/werf/compare/v1.2.27...v1.2.28) (2021-10-11)


### Features

#### Alias tags support [#3706](https://github.com/werf/werf/pull/3706)

* The option `--add-custom-tag=TAG_FORMAT` sets tag aliases for the content-based tag of each image (can be used multiple times).
* The option `--use-custom-tag=TAG_FORMAT` allows using tag alias in helm templates instead of an image content-based tag (NOT RECOMMENDED).
* If there is more than one image in the werf config it is necessary to use the image name shortcut `%image%` or `%image_slug%` in the tag format (e.g. `$WERF_ADD_CUSTOM_TAG_1="%image%-tag1"`, `$WERF_ADD_CUSTOM_TAG_2="%image%-tag2"`).
* For cleaning custom tags and associated content-based tag are treated as one:
  * The cleanup command deletes/keeps all tags following the cleaning policies for content-based tags.
  * The cleanup command keeps all when any tag is used in k8s.
* By default, alias tags are not allowed by giterminism, and it is necessary to use [werf-giterminism.yaml](https://werf.io/documentation/v1.2/reference/werf_giterminism_yaml.html) to activate options:
    ```yaml
    giterminismConfigVersion: 1 
    cli:
      allowCustomTags: true
    ```

### Bug Fixes

* final repo options not set for get-autogenerated-values command ([ff70054](https://www.github.com/werf/werf/commit/ff700543c873d2e7a74928be2d4c748b176624eb))
* **host-cleanup:** "permission denied" errors, do not wipe git-patches on every run ([2840427](https://www.github.com/werf/werf/commit/2840427158340ec1d2db8a90597c47facf489120))

### [1.2.27](https://www.github.com/werf/werf/compare/v1.2.25...v1.2.27) (2021-10-11)


### Features

* Completed first step of buildah adoption: allow building of dockerfiles with buildah on any supported by the werf platform (linux, windows and macos).
   * Enable buildah mode with `WERF_BUILDAH_CONTAINER_RUNTIME=auto|native-rootless|docker-with-fuse` environment variable:
        * `native-rootless` mode uses local storage and runs only under Linux.
        * `docker-with-fuse` mode runs buildah inside docker enabling crossplatform buildah support. This mode could be changed later to use podman instead of docker server.

### Bug Fixes

* spelling ([994af88](https://www.github.com/werf/werf/commit/994af880a8e00d52437227a82e52d0b184a17ae0))

### [1.2.26](https://www.github.com/werf/werf/compare/v1.2.25...v1.2.26) (2021-10-08)


### Features

* Completed first step of buildah adoption: allow building of dockerfiles with buildah on any supported by the werf platform (linux, windows and macos).
   * Enable buildah mode with `WERF_BUILDAH_CONTAINER_RUNTIME=auto|native-rootless|docker-with-fuse` environment variable:
        * `native-rootless` mode uses local storage and runs only under Linux.
        * `docker-with-fuse` mode runs buildah inside docker enabling crossplatform buildah support. This mode could be changed later to use podman instead of docker server.

### Bug Fixes

* spelling ([994af88](https://www.github.com/werf/werf/commit/994af880a8e00d52437227a82e52d0b184a17ae0))

### [1.2.25](https://www.github.com/werf/werf/compare/v1.2.24...v1.2.25) (2021-10-07)


### Bug Fixes

* **cleanup:** fix "should reset storage cache" error during werf-cleanup and werf-purge ([dd43b68](https://www.github.com/werf/werf/commit/dd43b6839b2a28ce2b4273f77c47f2fa0994969e))

### [1.2.24](https://www.github.com/werf/werf/compare/v1.2.23...v1.2.24) (2021-10-04)


### Bug Fixes

* **dev:** deletion of untracked files not taken into account ([c67a956](https://www.github.com/werf/werf/commit/c67a956b28c7e4e79d8fa49d27e2353fdf462a59))
* **dev:** submodule changes may not be taken into account ([f3b2fab](https://www.github.com/werf/werf/commit/f3b2fabbf169420e68ecad04fcf4834694f79c29))

### [1.2.23](https://www.github.com/werf/werf/compare/v1.2.22...v1.2.23) (2021-09-23)


### Bug Fixes

* panic in dismiss command, helm regsitry client initialization failure ([6a2e159](https://www.github.com/werf/werf/commit/6a2e159b8e54c9631c7eb7bc175dcd9b0ec37305))

### [1.2.22](https://www.github.com/werf/werf/compare/v1.2.21...v1.2.22) (2021-09-23)


### Bug Fixes

* sharing not thread safe go-git tree and storer ([1e2755b](https://www.github.com/werf/werf/commit/1e2755be30c7d1d0c5f757adf4fafa65b880e353))

### [1.2.21](https://www.github.com/werf/werf/compare/v1.2.19...v1.2.21) (2021-09-21)


### Bug Fixes

* **stapel:** changes in directories of import.include/excludePaths not triggered import ([f9043c3](https://www.github.com/werf/werf/commit/f9043c3316d41119e4f51ace7319a2ef2b92c5c3))

### [1.2.19](https://www.github.com/werf/werf/compare/v1.2.18+fix1...v1.2.19) (2021-09-17)


### Bug Fixes

* broken release building script, fix site installation instructions ([3dc31f2](https://www.github.com/werf/werf/commit/3dc31f2e4811084b0df93f017f832413c315740e))
* **dev:** fail on retry of a command with a deleted file ([d286821](https://www.github.com/werf/werf/commit/d28682109d096bffba1e4ba78c63405d2baaf84d))
* **dev:** special characters not handled properly ([1e72887](https://www.github.com/werf/werf/commit/1e72887d20119f8268a20b1fe84a869741416321))
* force disable CGO for release builds ([7a69ca7](https://www.github.com/werf/werf/commit/7a69ca736c457dd046d10b2fa43b8f2e296f143f))
* separate --insecure-helm-dependencies flag ([b2beb3a](https://www.github.com/werf/werf/commit/b2beb3ad94fdd560b4021fde40d487169203cefd))
