# Changelog

### [1.2.63](https://www.github.com/werf/werf/compare/v1.2.62...v1.2.63) (2022-02-04)


### Bug Fixes

* **bundles:** remove incompatibility with helm 3 ([ba1e5b4](https://www.github.com/werf/werf/commit/ba1e5b493c68305617b7cfb25ca77c24935f4e45))

### [1.2.62](https://www.github.com/werf/werf/compare/v1.2.61...v1.2.62) (2022-02-03)


### Bug Fixes

* **slugify:** limited slug is not idempotent ([4b06e8a](https://www.github.com/werf/werf/commit/4b06e8a753d107f655e4376f9e9297a5be8cb82e))

### [1.2.61](https://www.github.com/werf/werf/compare/v1.2.60...v1.2.61) (2022-02-02)


### Features

* **images-imports:** added reference docs ([8d6ef61](https://www.github.com/werf/werf/commit/8d6ef617e669bdb89f5f20367657dc4fdcd9a8f2))

### [1.2.60](https://www.github.com/werf/werf/compare/v1.2.59...v1.2.60) (2022-02-01)


### Features

* **config:** dependencies directive parser ([3eb94e4](https://www.github.com/werf/werf/commit/3eb94e4efcea50580177a28b5d0b782d215fabbe))
* **images-dependencies:** implement images dependencies for dockerfile builder ([f8b0204](https://www.github.com/werf/werf/commit/f8b02046f495f9753e2b878832f3f5697fadb96a))
* **images-dependencies:** stapel deps configuration for dependencies stage ([30f06fb](https://www.github.com/werf/werf/commit/30f06fb0537c508e929b9a6a2adba028982695a8))
* **images-imports:** dependencies directive parser ([0fc45d5](https://www.github.com/werf/werf/commit/0fc45d5996c709f16cee153a032884a3433e7197))
* **images-imports:** respect dependencies during build ([4adb6a3](https://www.github.com/werf/werf/commit/4adb6a39923ec778dfc74aec3abce9017e93eab5))


### Bug Fixes

* **dockerfile:** validate base image resolved to non-empty image ([e6f90c1](https://www.github.com/werf/werf/commit/e6f90c1dc78ff2c979ffaece11983d6f1c01a156))
* **images-dependencies:** forbid after/before for dockerfile deps ([38df0c7](https://www.github.com/werf/werf/commit/38df0c7527d6c2ae2d4e31a9c97b613a409a9e6c))
* **images-imports:** added import type=ImageID into validation ([b58eb07](https://www.github.com/werf/werf/commit/b58eb07db429e89b72cecae2d50162872a2786ca))

### [1.2.59](https://www.github.com/werf/werf/compare/v1.2.58...v1.2.59) (2022-01-27)


### Features

* **images-dependencies:** implement images dependencies for stapel builder ([5d5f144](https://www.github.com/werf/werf/commit/5d5f1449dd0791b4f89e60e7f454c13b85c52d7a))
* **images-dependencies:** introduce basic image dependencies configuration structs ([da36104](https://www.github.com/werf/werf/commit/da36104fdf9aa5d2223d1bde0464e4af582eb88a))
* **images-dependencies:** introduce basic image dependencies configuration structs (fix) ([1ef7073](https://www.github.com/werf/werf/commit/1ef7073e89a637dd6b1095f7cd2e696ab35ae1a7))
* **images-dependencies:** rename imports to dependencies ([725fbc9](https://www.github.com/werf/werf/commit/725fbc94eb9b8e863d7ceae81ad8e5d15db6114e))


### Bug Fixes

* 'werf helm get-release' command panic ([bc52c8e](https://www.github.com/werf/werf/commit/bc52c8e1ed7c54e0ee1118671a1ba6fa9c47a318))
* **build:** multi-stage does not work properly with build args ([2b59c76](https://www.github.com/werf/werf/commit/2b59c76140b714ce3a87ff8fa0a7593e5a2d1097))
* **quay:** ignore TAG_EXPIRED broken tags ([c302c05](https://www.github.com/werf/werf/commit/c302c05fa607a86abf13027ff6f25a272155053e))

### [1.2.58](https://www.github.com/werf/werf/compare/v1.2.57...v1.2.58) (2022-01-21)


### Features

* **build:** commit info in build container ([f1e3372](https://www.github.com/werf/werf/commit/f1e3372ee2da6f66f662b797183c47772dff63eb))
* **gc:** run host garbage collection in background ([29a1ea5](https://www.github.com/werf/werf/commit/29a1ea5a0129f6db19f4023a5e6459842bff4b90))
* **werf.yaml:** support [[ namespace ]] param in the helm-release template ([82d54e9](https://www.github.com/werf/werf/commit/82d54e9d46d55828649a3bb3f0bc05cb513fd312))

### [1.2.57](https://www.github.com/werf/werf/compare/v1.2.56...v1.2.57) (2022-01-19)


### Bug Fixes

* **build:** virtual merge commits and inconsistent build cache ([7372992](https://www.github.com/werf/werf/commit/73729924905ef03a199a5d2fb26cccded5f6c69a))
* **git:** fast, ad-hoc fix, return exec.ExitError from gitCmd.Run() ([d737d8b](https://www.github.com/werf/werf/commit/d737d8bbfd216ad9ea068550e625a0b8c1f0c570))
* **git:** git warnings sometimes break werf ([0a50961](https://www.github.com/werf/werf/commit/0a509618bf8ba3734b7c942486e87895200ead6d))

### [1.2.56](https://www.github.com/werf/werf/compare/v1.2.55...v1.2.56) (2022-01-18)


### Features

* **build:** expose commit info in werf templates ([4c2b33a](https://www.github.com/werf/werf/commit/4c2b33af0dde4bf88c233c61f1381c3a7f800e5c))


### Bug Fixes

* **dependencies:** update deps, incompatible image-spec ([4518b58](https://www.github.com/werf/werf/commit/4518b5825dfd0ce80c48a5c11fe0ff0898c7cdbd))

### [1.2.55](https://www.github.com/werf/werf/compare/v1.2.54...v1.2.55) (2021-12-29)


### Features

* Added login and logout cli commands for container registry ([0b7e147](https://www.github.com/werf/werf/commit/0b7e14716145af54a7790c1303c48802cbc19faa))

### Docs

* Buildah articles & run in container (#4043). Correcting & translating Buildah and Run in Kubernetes articles to russian.


### [1.2.54](https://www.github.com/werf/werf/compare/v1.2.53...v1.2.54) (2021-12-24)


### Bug Fixes

* parse git versions without patch or minor version ([17a20be](https://www.github.com/werf/werf/commit/17a20be81975240063bcdb004801274639321b33))
* warning in `git version` break werf ([266bad0](https://www.github.com/werf/werf/commit/266bad0acb4ee614bc82ac77819cdf56e9f74123))

### [1.2.53](https://www.github.com/werf/werf/compare/v1.2.52...v1.2.53) (2021-12-17)


### Bug Fixes

* Add missing WERF_TIMEOUT variable for --timeout param ([672d379](https://www.github.com/werf/werf/commit/672d37952a51551d6d72305704476da1f89f5276))

### [1.2.52](https://www.github.com/werf/werf/compare/v1.2.51...v1.2.52) (2021-12-16)


### Features

* **multiwerf:** print multiwerf deprecation warning if multiwerf outdated ([12d0f55](https://www.github.com/werf/werf/commit/12d0f55357cf0efcdc05db23835f997357b99eb6))


### Bug Fixes

* **harbor:** detect usage of harbor without --repo-container-registry=harbor option ([a3843f9](https://www.github.com/werf/werf/commit/a3843f9ba34088e8c4c6cf521e274421cf1c6059))

### [1.2.51](https://www.github.com/werf/werf/compare/v1.2.50...v1.2.51) (2021-12-10)


### Bug Fixes

* **buildah:** do not use ignore_chown_errors option for overlay storage driver ([299a33e](https://www.github.com/werf/werf/commit/299a33e858d2ddfff3a458ea425af8db79e69c95))

### [1.2.50](https://www.github.com/werf/werf/compare/v1.2.49...v1.2.50) (2021-12-10)


### Features

* **buildah:** support autodetection of native mode for overlayfs ([7858360](https://www.github.com/werf/werf/commit/7858360a502087cecafebf026523d4a4033a3da6))


### Bug Fixes

* **buildah:** Buildah mode autodetection ([80b9e90](https://www.github.com/werf/werf/commit/80b9e903b61c391dacfaee674f953b4bd591c443))

### [1.2.49](https://www.github.com/werf/werf/compare/v1.2.48...v1.2.49) (2021-12-09)


### Bug Fixes

* **buildah:** pass default registries.conf to native buildah ([ca2995a](https://www.github.com/werf/werf/commit/ca2995ade2ed6d1282beb3dd2da032e1fcdbf06c))

### [1.2.48](https://www.github.com/werf/werf/compare/v1.2.47...v1.2.48) (2021-12-09)


### Features

* **buildah:** added new official werf images:
  * ghcr.io/werf/werf:1.2-{alpha|beta|ea|stable}-{alpine|ubuntu|centos|fedora};
  * ghcr.io/werf/werf:1.2-{alpha|beta|ea|stable} (same as ghcr.io/werf/werf:1.2-{alpha|beta|ea|stable}-alpine);
* **buildah:** native OCI rootless mode; vfs storage driver; bugfixes ([58e92a2](https://www.github.com/werf/werf/commit/58e92a28b65f2492b5388582740bcb18f3e26068)).
* **buildah:** improve docs about running werf in containers.


### Bug Fixes

* **cleanup:** do not use stages-storage-cache when getting all stages list ([7e9651b](https://www.github.com/werf/werf/commit/7e9651ba0a78319bb270fb8cbfd8614e08bc56fb))
* **deploy:** status-progress-period and hooks-status-progress-period params fix ([2522b25](https://www.github.com/werf/werf/commit/2522b25e6b27488bfb5beae98fb1a379571f5202))

### [1.2.47](https://www.github.com/werf/werf/compare/v1.2.46...v1.2.47) (2021-12-03)


### Docs

* New docs for running werf in container (experimental): [https://werf.io/documentation/v1.2.46/advanced/ci_cd/run_in_container/run_in_docker_container.html](https://werf.io/documentation/v1.2.46/advanced/ci_cd/run_in_container/run_in_docker_container.html).

### Bug Fixes

* **cleanup:** ignore harbor "unsupported 404 status code" errors ([adf60a0](https://www.github.com/werf/werf/commit/adf60a02006ff27c6f610451e63f7a74db765c18))

### [1.2.46](https://www.github.com/werf/werf/compare/v1.2.45...v1.2.46) (2021-12-02)


### Features

* **docs**: added docs for running werf in container in experimental mode.

### [1.2.45](https://www.github.com/werf/werf/compare/v1.2.44...v1.2.45) (2021-12-02)


### Features

* **buildah:** publish initial werf image with compiled werf binary and buildah environment ([20dde28](https://www.github.com/werf/werf/commit/20dde28c4d182469e7c584a9b4e0a959686504f4))
* **buildah:** working native-rootless buildah mode inside docker container ([ed4fa0a](https://www.github.com/werf/werf/commit/ed4fa0af4c92da7a6daa8cd27b23eb5bf21121a9))


### Bug Fixes

* panic when docker image inspect has failed with unexpected error ([6011721](https://www.github.com/werf/werf/commit/60117213f3dbb1c60b971568a4eb897a573d6e90))

### [1.2.44](https://www.github.com/werf/werf/compare/v1.2.43...v1.2.44) (2021-11-26)


### Features

* **buildah:** publish initial werf image with compiled werf binary and buildah environment ([20dde28](https://www.github.com/werf/werf/commit/20dde28c4d182469e7c584a9b4e0a959686504f4))
* **buildah:** working native-rootless buildah mode inside docker container ([ed4fa0a](https://www.github.com/werf/werf/commit/ed4fa0af4c92da7a6daa8cd27b23eb5bf21121a9))


### Bug Fixes

* panic when docker image inspect has failed with unexpected error ([6011721](https://www.github.com/werf/werf/commit/60117213f3dbb1c60b971568a4eb897a573d6e90))

### [1.2.43](https://www.github.com/werf/werf/compare/v1.2.42...v1.2.43) (2021-11-25)


### Bug Fixes

* create new release ([c7e314c](https://www.github.com/werf/werf/commit/c7e314cf6eebedd5b1f9db8bf766cbf1efd2480c))

### [1.2.42](https://www.github.com/werf/werf/compare/v1.2.41...v1.2.42) (2021-11-24)


### Bug Fixes

* **host_cleanup:** getting true dangling images ([22949ca](https://www.github.com/werf/werf/commit/22949cab4cb0251831ee6111d7861b85adea8db8))

### [1.2.41](https://www.github.com/werf/werf/compare/v1.2.40...v1.2.41) (2021-11-18)


### Bug Fixes

* fix(deploy): fix broken 3 way merge cases: https://github.com/werf/werf/issues/3461 and https://github.com/werf/werf/issues/3462. Upstream helm issue: https://github.com/helm/helm/issues/10363.

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
   * Enable buildah mode with `WERF_CONTAINER_RUNTIME_BUILDAH=auto|native-rootless|docker-with-fuse` environment variable:
        * `native-rootless` mode uses local storage and runs only under Linux.
        * `docker-with-fuse` mode runs buildah inside docker enabling crossplatform buildah support. This mode could be changed later to use podman instead of docker server.

### Bug Fixes

* spelling ([994af88](https://www.github.com/werf/werf/commit/994af880a8e00d52437227a82e52d0b184a17ae0))

### [1.2.26](https://www.github.com/werf/werf/compare/v1.2.25...v1.2.26) (2021-10-08)


### Features

* Completed first step of buildah adoption: allow building of dockerfiles with buildah on any supported by the werf platform (linux, windows and macos).
   * Enable buildah mode with `WERF_CONTAINER_RUNTIME_BUILDAH=auto|native-rootless|docker-with-fuse` environment variable:
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
