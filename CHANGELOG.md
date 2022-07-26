# Changelog

### [1.2.140](https://www.github.com/werf/werf/compare/v1.2.139...v1.2.140) (2022-07-26)


### Features

* **render:** print build log if failed ([8007487](https://www.github.com/werf/werf/commit/8007487fb168851d5b2aae957a3ac0156926f912))


### Bug Fixes

* **render:** live output in verbose/debug mode does not work properly ([5032780](https://www.github.com/werf/werf/commit/50327800cda0a2be39b8d68bdf70147965368426))

### [1.2.139](https://www.github.com/werf/werf/compare/v1.2.138...v1.2.139) (2022-07-25)


### Features

* **buildah:** auto Buildah Ulimits from current user ulimits ([28d4d28](https://www.github.com/werf/werf/commit/28d4d28b42a076f496d9227382a37dd24ef79027))
* **buildah:** configurable Ulimit via `$WERF_BUILDAH_ULIMIT` ([734963a](https://www.github.com/werf/werf/commit/734963a4301b120a3a8d1e09402ac98df79fc306))

### [1.2.138](https://www.github.com/werf/werf/compare/v1.2.137...v1.2.138) (2022-07-22)


### Bug Fixes

* **helm:** CleanupOnFail imrovements ([ec749a1](https://www.github.com/werf/werf/commit/ec749a1b480eb8087db902d7b3915e4bc5b8dab4))

### [1.2.137](https://www.github.com/werf/werf/compare/v1.2.136...v1.2.137) (2022-07-21)


### Bug Fixes

* **kubedog:** generic tracker: improve logging + few possible fixes ([b93b1fe](https://www.github.com/werf/werf/commit/b93b1fead3d571be70c56f1b35c9f8773f180f9f))

### [1.2.136](https://www.github.com/werf/werf/compare/v1.2.135...v1.2.136) (2022-07-21)


### Bug Fixes

* **stapel:** importing of symlink that points to directory does not work properly ([835260f](https://www.github.com/werf/werf/commit/835260f07bf43d79ccc0d5943907bbc2be32fdfd))
* **test:** init werf docker failed on macOS ([8306b57](https://www.github.com/werf/werf/commit/8306b572edadb4f8583c5d098457c21ba1e07c35))

### [1.2.135](https://www.github.com/werf/werf/compare/v1.2.134...v1.2.135) (2022-07-20)


### Bug Fixes

* **kubedog:** Generic tracker hangs if no list/watch access ([62b42b1](https://www.github.com/werf/werf/commit/62b42b18ea2d6856c828c76ff2aff1454115d4c5))

### [1.2.134](https://www.github.com/werf/werf/compare/v1.2.133...v1.2.134) (2022-07-19)


### Bug Fixes

* revert "feat: tpl performance improved" ([a8d4b58](https://www.github.com/werf/werf/commit/a8d4b58fd1052bb07f80f46892a2304de9955eef))

### [1.2.133](https://www.github.com/werf/werf/compare/v1.2.132...v1.2.133) (2022-07-19)


### Bug Fixes

* **helm:** skip delete if unmatched resource ownership metadata ([ab416ba](https://www.github.com/werf/werf/commit/ab416bab534d63382715ba30b861bb10c64ce56b))

### [1.2.132](https://www.github.com/werf/werf/compare/v1.2.131...v1.2.132) (2022-07-18)


### Features

* **helm:** werf.io/no-activity-timeout annotation ([7b84ea0](https://www.github.com/werf/werf/commit/7b84ea057b8acb7d060344b2893c10878b4987e8))
* **telemetry:** added CommandExited durationMs field ([7d7c71a](https://www.github.com/werf/werf/commit/7d7c71a4ad63de03ea34dd627c44b5d354f02c37))


### Bug Fixes

* **kubedog:** increase default NoActivityTimeout to 4min ([7a6aa6f](https://www.github.com/werf/werf/commit/7a6aa6f2b7ec3d19fe55bd56197209b261b414b1))

### [1.2.131](https://www.github.com/werf/werf/compare/v1.2.130...v1.2.131) (2022-07-18)


### Bug Fixes

* **kubedog:** 3way merge patches had missing fields ([870c4e3](https://www.github.com/werf/werf/commit/870c4e3abfb1e315c215221c28f64be3cbed3d38))

### [1.2.130](https://www.github.com/werf/werf/compare/v1.2.129...v1.2.130) (2022-07-18)


### Features

* **kubedog:** improve progress status + fix Job Duration not changed ([a8ce29b](https://www.github.com/werf/werf/commit/a8ce29b50a6e91de33281841a67fa6e02079afeb))

### [1.2.129](https://www.github.com/werf/werf/compare/v1.2.128...v1.2.129) (2022-07-15)


### Features

* **custom-tag:** add custom tags params for converge and bundle-publish ([4cc2802](https://www.github.com/werf/werf/commit/4cc2802e02c17878f5589b49bcc032293f8c968e))
* **kubedog:** show Ready resources only once ([d8281b8](https://www.github.com/werf/werf/commit/d8281b86f04c2a085507349e394235eaacc1c2bd))

### [1.2.128](https://www.github.com/werf/werf/compare/v1.2.127...v1.2.128) (2022-07-14)


### Features

* **telemetry:** anonimized cli options usage, exit-code event, tune timeouts ([3402424](https://www.github.com/werf/werf/commit/340242453ceb906d814dcfb31d76ab4b95318e1d))
* **telemetry:** fix telemetry lags; telemetry logs; ignore commands list ([9a2310f](https://www.github.com/werf/werf/commit/9a2310f97941e651730fd5f72823580e3210d983))

### [1.2.127](https://www.github.com/werf/werf/compare/v1.2.126...v1.2.127) (2022-07-13)


### Features

* **helm:** skip templating if plain text with no templates passed to `tpl` ([c1e00e4](https://www.github.com/werf/werf/commit/c1e00e4563f1d0dd3d7df9e6768cb884657daa5a))

### [1.2.126](https://www.github.com/werf/werf/compare/v1.2.125...v1.2.126) (2022-07-13)


### Bug Fixes

* **bundles:** --secret-values option for bundle-render command ([f722ec9](https://www.github.com/werf/werf/commit/f722ec99d6594700e85b6d91acc4cb91030aed04))

### [1.2.125](https://www.github.com/werf/werf/compare/v1.2.124...v1.2.125) (2022-07-13)


### Features

* **helm:** `tpl` performance improved ([7422424](https://www.github.com/werf/werf/commit/7422424587221c6775829609e6d60067426e04d4))

### [1.2.124](https://www.github.com/werf/werf/compare/v1.2.123...v1.2.124) (2022-07-12)


### Features

* **helm:** split long templating errors over multiple lines ([6af7192](https://www.github.com/werf/werf/commit/6af71924b0621ddd09af741a3361739c3edafc0b))
* **kube-run:** --add-annotations/labels + pass service annotations ([ba2fc7c](https://www.github.com/werf/werf/commit/ba2fc7cf9b11128df3b4874e98ce144e053ccaa0))


### Bug Fixes

* **secrets:** compatible behaviour of secret values edit command with previous versions ([89cbef5](https://www.github.com/werf/werf/commit/89cbef50c9861c9e4468562267239506d4a8eaa5))

### [1.2.123](https://www.github.com/werf/werf/compare/v1.2.122...v1.2.123) (2022-07-11)


### Features

* **telemetry:** use new telemetry with updated schema and projectID ([cf784f7](https://www.github.com/werf/werf/commit/cf784f7c36979e53ace77fda72c96caa85c7d2fe))

### [1.2.122](https://www.github.com/werf/werf/compare/v1.2.121...v1.2.122) (2022-07-08)


### Bug Fixes

* remove LegacyStageImageContainer accidental debug messages ([e70d8b6](https://www.github.com/werf/werf/commit/e70d8b6f6740ad2d61d9089ecdb9dc4c855eeb52))

### [1.2.121](https://www.github.com/werf/werf/compare/v1.2.120...v1.2.121) (2022-07-06)


### Features

* **helm:** track Helm hooks of any kind ([86ba23f](https://www.github.com/werf/werf/commit/86ba23f94fd19a7d43f2d8e26a4b087e656c12b0))


### Bug Fixes

* **kubedog:** non-blocking mode didn't work ([0cc6882](https://www.github.com/werf/werf/commit/0cc68826ed57aff660ceae676f63ddcd6d23487a))

### [1.2.120](https://www.github.com/werf/werf/compare/v1.2.119...v1.2.120) (2022-07-05)


### Features

* **kubedog:** generic resources tracking ([93ed2e5](https://www.github.com/werf/werf/commit/93ed2e524436dda75289f6b34ff0a16bb963fb86))

### [1.2.119](https://www.github.com/werf/werf/compare/v1.2.118...v1.2.119) (2022-07-01)


### Features

* **telemetry:** experiments with opentelemetry, traces and clickhouse storage ([2e404a9](https://www.github.com/werf/werf/commit/2e404a963be263498244a473418c025f5646ef10))


### Bug Fixes

* **secrets:** panic and incorrect behaviour during secrets edit ([289400d](https://www.github.com/werf/werf/commit/289400d3aec3794dc310942fb81d0b3423843753))

### [1.2.118](https://www.github.com/werf/werf/compare/v1.2.117...v1.2.118) (2022-06-29)


### Features

* **telemetry:** basic telemetry client and local setup ([6dcbd3e](https://www.github.com/werf/werf/commit/6dcbd3e7d8d108636510a3f22752ddbd5031f31f))


### Bug Fixes

* **docker-instructions:** exactOptionValues option to fix docker-server backend options evaluation ([9b3dbf9](https://www.github.com/werf/werf/commit/9b3dbf97b5b7e9e8f3dde1b8966b43d106b08f03))
* **external-deps:** use Unstructured instead of builtin types ([afbb5b4](https://www.github.com/werf/werf/commit/afbb5b4f96b22e07c570a1c6a233b9addc93cebc))
* **git-worktree:** ignore existing locked service worktree when re-adding ([7775193](https://www.github.com/werf/werf/commit/7775193d1a7f01ebf93a28b361b10be47d146563))
* **submodules:** auto handle "commits not present" patch creation error ([91a829b](https://www.github.com/werf/werf/commit/91a829b341fe3da2064f8e3e25d3394783089c2e))

### [1.2.117](https://www.github.com/werf/werf/compare/v1.2.116...v1.2.117) (2022-06-21)


### Features

* **buildah:** $WERF_CONTAINERIZED will override in container detection ([5766e6a](https://www.github.com/werf/werf/commit/5766e6a0fa6bc5400190805e6708d70ae3b3f10a))
* **buildah:** container runtime autodetection ([695ae97](https://www.github.com/werf/werf/commit/695ae97dfc28c23763257b20223f6de3b8758abb))
* **secrets:** preserve comments, order and aliases in the secrets edit commands ([5bc6092](https://www.github.com/werf/werf/commit/5bc6092cf16346fa3e1984959b9571978ad1cc30))


### Bug Fixes

* **buildah:** improve whether we are in container detection ([532a002](https://www.github.com/werf/werf/commit/532a002494963a024d34fb550a8126dd7d4f4933))
* **host-cleanup:** do not remove v1.2 local storage images ([9702026](https://www.github.com/werf/werf/commit/9702026a2e8b9834520ce08ada399faf5ca8d51c))
* **host-cleanup:** host cleanup not working in buildah mode ([cb51e32](https://www.github.com/werf/werf/commit/cb51e3255601756a82143387f656e7d4670b0cab))
* **host-cleanup:** run host cleanup without docker-server in buildah mode ([f1b1403](https://www.github.com/werf/werf/commit/f1b140316c7868e8a7bc12a52531603449d7cff3))

### [1.2.116](https://www.github.com/werf/werf/compare/v1.2.115...v1.2.116) (2022-06-16)


### Features

* **external-deps:** external dependencies for release resources ([73e6bcc](https://www.github.com/werf/werf/commit/73e6bcca54570c97b398a035fb60ae9af85e894c))
* **external-deps:** external dependencies now available for `werf helm` ([c968c08](https://www.github.com/werf/werf/commit/c968c08ddc05bcd546ff07482a8a67447809d645))

### [1.2.115](https://www.github.com/werf/werf/compare/v1.2.114...v1.2.115) (2022-06-15)


### Bug Fixes

* **bundles:** cleanup --final-repo param usage in bundles ([4d77117](https://www.github.com/werf/werf/commit/4d771175eeee0eaa26689a869d83644e44da94c8))
* **docs:** add info about published rock-solid images ([9b09593](https://www.github.com/werf/werf/commit/9b09593dfa847aafff59b3e2865396d3807f07e3))
* **final-repo:** service values .Values.werf.repo should use --final-repo instead of --repo ([e0562f6](https://www.github.com/werf/werf/commit/e0562f6a6c1dc2662e02c8b30523affb6a2e2ce6))
* **helm:** fix werf panic and helm plugins with error codes ([a39a1a0](https://www.github.com/werf/werf/commit/a39a1a0b87c1fd96c8620aab5e0f34f12ab52996))

### [1.2.114](https://www.github.com/werf/werf/compare/v1.2.113...v1.2.114) (2022-06-10)


### Bug Fixes

* **custom-tags:** support custom tags for --final-repo images ([e785c87](https://www.github.com/werf/werf/commit/e785c87d79ff1fcf35eb766abb5462e019bc7403))
* **helm:** fix 'werf helm *' commands to correctly initialize namespace; fix output ([f7faaa7](https://www.github.com/werf/werf/commit/f7faaa7970798081c26e82d171beef3b6d1ebff4))

### [1.2.113](https://www.github.com/werf/werf/compare/v1.2.112...v1.2.113) (2022-06-08)


### Bug Fixes

* **helm:** `unable to recognize "": no matches for kind "..." in version "..."` errors when base64 kubeconfig used ([90678ec](https://www.github.com/werf/werf/commit/90678ecc132503f70d60e53ce79b9688f033e6a0))

### [1.2.112](https://www.github.com/werf/werf/compare/v1.2.111...v1.2.112) (2022-06-08)


### Bug Fixes

* **export-values:** propagate result of export-values to all parent charts Values ([12a0b54](https://www.github.com/werf/werf/commit/12a0b543f2245cbfdd55a6a75fb174711985cf43))

### [1.2.111](https://www.github.com/werf/werf/compare/v1.2.110...v1.2.111) (2022-06-07)


### Features

* **dismiss:** dont fail if no release found ([6f79a18](https://www.github.com/werf/werf/commit/6f79a186b747f8096611a2ad6440a35084f72699))


### Bug Fixes

* **dismiss:** --with-namespace created empty namespace if release already uninstalled ([7c1ab9b](https://www.github.com/werf/werf/commit/7c1ab9bc8ccdfabd468bcb2538d04eab37a8a5a1))
* **helm:** fix werf_secret_file not working in werf helm template command ([b2cec4b](https://www.github.com/werf/werf/commit/b2cec4b20a4f929b5ed729f6bf2ba271da5d2af7))
* **helm:** plugins positional arguments not passed properly ([98f9003](https://www.github.com/werf/werf/commit/98f90034ec8fcb059acf57c527d250aa49d99983))

### [1.2.110](https://www.github.com/werf/werf/compare/v1.2.109...v1.2.110) (2022-06-03)


### Features

* deploy in multiple stages; improve 3way merge ([9a8d3ee](https://www.github.com/werf/werf/commit/9a8d3ee4fb261bfa231aa76748667f45aa3afe9d))

### [1.2.109](https://www.github.com/werf/werf/compare/v1.2.108...v1.2.109) (2022-06-02)


### Bug Fixes

* **kube-run:** --copy-from skipped if command failed ([8f595ec](https://www.github.com/werf/werf/commit/8f595ec87a8673aa2f00a433cc50248cc20fdc80))
* **kube-run:** better log message when command failed ([6551c8e](https://www.github.com/werf/werf/commit/6551c8e30c08ea585e61dd7d816844f845c08052))

### [1.2.108](https://www.github.com/werf/werf/compare/v1.2.107...v1.2.108) (2022-06-01)


### Features

* **buildah:** update buildah to v1.26.1 ([bf1f2d0](https://www.github.com/werf/werf/commit/bf1f2d0bcba888ed62cea3497ca557b4a6d8cf99))


### Bug Fixes

* **buildah:** buildah Dockerfile builder was not using layers cache ([8d9326d](https://www.github.com/werf/werf/commit/8d9326d7c8f325efce01765c64f8e907deaa6069))
* **dockerfile:** support RUN with --mount from another stage ([ebd544a](https://www.github.com/werf/werf/commit/ebd544a302b66d7210eea69179ac829a7b2abd02))
* **helm:** fix 'error preparing chart dependencies... file exists' ([3f32bf0](https://www.github.com/werf/werf/commit/3f32bf08125daac73fbc5905f726976e3b814d04))

### [1.2.107](https://www.github.com/werf/werf/compare/v1.2.106...v1.2.107) (2022-05-27)


### Bug Fixes

* **cache-repo:** panic when using cache repo and fromImage directive ([3ceb622](https://www.github.com/werf/werf/commit/3ceb622860d7bffe3e66db7e353faad87b04a6f5))
* **cache-repo:** panic when using cache-repo and building images existing in cache ([1c97593](https://www.github.com/werf/werf/commit/1c97593dbd0da919cae9f3825f333fd2d863094c))

### [1.2.106](https://www.github.com/werf/werf/compare/v1.2.105...v1.2.106) (2022-05-25)


### Features

* **bundles:** --secret-values option for werf-bundle-apply command ([2daea2b](https://www.github.com/werf/werf/commit/2daea2b52c9f10806b093ed770b3cd6f1b28b296))
* **cleanup:** optimize cleanup deployed resources images scanning regarding Jobs ([b7edaa3](https://www.github.com/werf/werf/commit/b7edaa3290132a644689c4cd4af4565112976f05))
* **docs:** New article about resources adoption ([5ab8f26](https://www.github.com/werf/werf/commit/5ab8f2604e432eff5137d1151c78521d417e3002))


### Bug Fixes

* **cleanup:** fix cleanup not using in-cluster kube config when using in-cluster mode ([967a6aa](https://www.github.com/werf/werf/commit/967a6aa649c41752945987620e8f25b1d884fcff))
* **render:** support for --kube-context param when --validate option used ([91869a8](https://www.github.com/werf/werf/commit/91869a8379035994e32d582c91647e68d34590c4))

### [1.2.105](https://www.github.com/werf/werf/compare/v1.2.104...v1.2.105) (2022-05-23)


### Bug Fixes

* **post-renderer:** fix null value validation panic in annotations and labels ([5d80460](https://www.github.com/werf/werf/commit/5d80460a3bf3d86927dde49535b10eb1a3492c1b))

### [1.2.104](https://www.github.com/werf/werf/compare/v1.2.103...v1.2.104) (2022-05-19)


### Features

* **cross-platform-builds:** basic support of --platform=OS/ARCH[/VARIANT] parameter for buildah builder ([276fc0f](https://www.github.com/werf/werf/commit/276fc0ff1bf1e2137d490f8018ebae78ece4fa66))


### Bug Fixes

* **migrate2to3:** new target namespace not respected in new Release ([985e241](https://www.github.com/werf/werf/commit/985e241384779fba176d08ff7401aa397f6813fb))
* warning message misspeling fix ([15c2dbb](https://www.github.com/werf/werf/commit/15c2dbba532df8ce3afa0e02ed1a40d8dfa4d612))

### [1.2.103](https://www.github.com/werf/werf/compare/v1.2.102...v1.2.103) (2022-05-18)


### Bug Fixes

* **git:** fix error "unable to clone repo: reference delta not found" ([1733ccd](https://www.github.com/werf/werf/commit/1733ccd28910cdb6156fc46bf07c909aa51468bf))
* **helm:** prevent bug with pre-upgrade helm hooks, which was used from the previous release revision ([18570d3](https://www.github.com/werf/werf/commit/18570d33a76a262562851f44bc4ffbdfc0b9e95b))
* **post-renderer:** non-strict labels and annotations validation in werf's post-renderer ([18dd510](https://www.github.com/werf/werf/commit/18dd510bd33f708bc6385b1a0513ee5e0f38079b))

### [1.2.102](https://www.github.com/werf/werf/compare/v1.2.101...v1.2.102) (2022-05-18)


### Features

* **kube-run:** --copy-from-file and --copy-from-dir opts ([dcfa982](https://www.github.com/werf/werf/commit/dcfa9828ae65ca253416439a2623a2fb21fa2d68))
* **kube-run:** add --copy-to; replace --copy-from-[file|dir] with --copy-from ([231ccbc](https://www.github.com/werf/werf/commit/231ccbc8e228597c7d0fde0c78fedb27adb67fec))


### Bug Fixes

* **kube-run:** ignore image CMD ([98bfc7e](https://www.github.com/werf/werf/commit/98bfc7efd1acc2ceab183e8418d041ab0ac68904))

### [1.2.101](https://www.github.com/werf/werf/compare/v1.2.100...v1.2.101) (2022-05-16)


### Features

* **stapel-to-buildah:** allow buildah to build stapel images with shell builder ([27a1d49](https://www.github.com/werf/werf/commit/27a1d497af3e6d70181aa923a949c803318bc73b))


### Bug Fixes

* panic when --cache-repo used ([ec2ed93](https://www.github.com/werf/werf/commit/ec2ed932d128086bc3f61198486c3f379f1948fc))
* panic when --secondary-repo or --cache-repo used ([c59f1f9](https://www.github.com/werf/werf/commit/c59f1f9102a0fba364ecaee430a662c9a68c9e6d))
* **stapel-to-buildah:** fix cleanup parent-id issue for images built with buildah ([56e90e2](https://www.github.com/werf/werf/commit/56e90e25d67daaa5edc46818880a11f1e477b58e))

### [1.2.100](https://www.github.com/werf/werf/compare/v1.2.99...v1.2.100) (2022-05-13)


### Bug Fixes

* **imports:** recursive copying issues ([9351c25](https://www.github.com/werf/werf/commit/9351c25fe3aa8ef0842c23624b5d3a6a4b742e0d))
* switch to actions/checkout@v3 ([ba3ac8e](https://www.github.com/werf/werf/commit/ba3ac8ecbd4a6201ec8eb5c23bb6642bd25e7c4d))

### [1.2.99](https://www.github.com/werf/werf/compare/v1.2.98...v1.2.99) (2022-05-11)


### Bug Fixes

* **helm-for-werf:** detailed error message for "current release manifest contains removed kubernetes api(s) ..." error ([8e8e5df](https://www.github.com/werf/werf/commit/8e8e5df89c991e81149bf0778699da853ab45199))
* **stapel-to-buildah:** added missing ssh-auth-sock and commit related envs, labels and volumes ([3835e62](https://www.github.com/werf/werf/commit/3835e62235e55b7253f35d64c43d99436488abf2))

### [1.2.98](https://www.github.com/werf/werf/compare/v1.2.97...v1.2.98) (2022-05-06)


### Features

* **docs:** added info about deploying bundles as helm chart dependencies ([188ec71](https://www.github.com/werf/werf/commit/188ec71bcf65b3bcf324369355afcd5a7d83442b))
* **docs:** werf cheat sheet ([091383e](https://www.github.com/werf/werf/commit/091383e091809c8bb3afcadebc097ec92c51270c))
* **stapel-to-buildah:** run user stages commands in the script using sh or bash ([e9aa1d4](https://www.github.com/werf/werf/commit/e9aa1d444a00cab465d3797a31388c1f6e48e6c8))

### [1.2.97](https://www.github.com/werf/werf/compare/v1.2.96...v1.2.97) (2022-05-06)


### Bug Fixes

* **helm:** fix export-values in subcharts case, improve broken 3wm case handling ([bf04268](https://www.github.com/werf/werf/commit/bf0426800e0778481fa8d6f1d537bb2887d1d33c))

### [1.2.96](https://www.github.com/werf/werf/compare/v1.2.95...v1.2.96) (2022-05-05)


### Features

* update helm v3.8.1 to v3.8.2 ([7f4e6b7](https://www.github.com/werf/werf/commit/7f4e6b71e5899a0868338ec185fd98e59d6c4e26))


### Bug Fixes

* **helm:** solved broken 3 way merge case when pre-upgrade hook fails ([a4610e3](https://www.github.com/werf/werf/commit/a4610e3e8c930b49ef215a3ca073e3074bd74e8b))

### [1.2.95](https://www.github.com/werf/werf/compare/v1.2.94...v1.2.95) (2022-05-04)


### Bug Fixes

* anchors support for extra annotations and labels post-renderer ([b8211a9](https://www.github.com/werf/werf/commit/b8211a995d85c59a29bcc0bc644af97dcf0949cb))

### [1.2.94](https://www.github.com/werf/werf/compare/v1.2.93...v1.2.94) (2022-04-28)


### Bug Fixes

* broken export-values ([ec1c0e4](https://www.github.com/werf/werf/commit/ec1c0e4631c49e2cf4530cac953cc984fe2f035b))

### [1.2.93](https://www.github.com/werf/werf/compare/v1.2.92...v1.2.93) (2022-04-28)


### Features

* **bundle:** implement 'bundle copy' command ([16dbd2e](https://www.github.com/werf/werf/commit/16dbd2e8e9a2f3e15bb85f598a009b479a636188))
* **cleanup:** add cleanup.keepBuiltWithinLastNHours directive in werf.yaml ([aabfcea](https://www.github.com/werf/werf/commit/aabfcea7172d6241906e207578009645abe5daec))
* **cleanup:** disable cleanup policies in werf.yaml ([c293f3d](https://www.github.com/werf/werf/commit/c293f3d8b8509f00615edd772e89065c704298e2))

### [1.2.92](https://www.github.com/werf/werf/compare/v1.2.91...v1.2.92) (2022-04-27)


### Features

* **bundle:** implement 'bundle copy' command ([92122e7](https://www.github.com/werf/werf/commit/92122e7fa9cdd8f1710555b2df6bddb1cd7f1cdc))
* support --show-only|-s helm-style render option + export-values chaining ([e9e3b86](https://www.github.com/werf/werf/commit/e9e3b86549ad078b9db6686e0984a76d3cc10271))


### Bug Fixes

* **render:** manifests keys sort order not preserved after rendering ([469ce7a](https://www.github.com/werf/werf/commit/469ce7af96b3442f29ad4106d056c75d7bb6b312))

### [1.2.91](https://www.github.com/werf/werf/compare/v1.2.90...v1.2.91) (2022-04-22)


### Bug Fixes

* **buildah_backend:** bump copyrec, fix broken windows build ([cefeb72](https://www.github.com/werf/werf/commit/cefeb72c7b3d2e4261df3c82bf2587cce214384c))
* git worktree switch invalidation loop ([e698bff](https://www.github.com/werf/werf/commit/e698bff8f547f35043cba3057f629fa0b7a9ad7e))

### [1.2.90](https://www.github.com/werf/werf/compare/v1.2.89...v1.2.90) (2022-04-21)


### Features

* **kube-run:** pod and container name can be used in --overrides ([686b402](https://www.github.com/werf/werf/commit/686b40293734609a1b2ad9523f0a74a392717411))
* **kube-run:** set --overrides-type=strategic for better merges ([9f222a5](https://www.github.com/werf/werf/commit/9f222a547741335458f9d33b72d4f1ec195cb858))


### Bug Fixes

* .helm/Chart.yaml chart name redefines project name from werf.yaml ([cda82f7](https://www.github.com/werf/werf/commit/cda82f729dd7bf3d3f73ab91c01ea1a0b5e0a9b3))
* **build:** cleanup orphan build containers on ctrl-c or gitlab cancel ([8702efa](https://www.github.com/werf/werf/commit/8702efaf5cae11da9c60e174397a280fa9537ba2))
* **deploy:** do not print secret values in debug mode by default ([44be01a](https://www.github.com/werf/werf/commit/44be01a3461d471b16757954ce9617d7cd3f317f))

### [1.2.89](https://www.github.com/werf/werf/compare/v1.2.88...v1.2.89) (2022-04-19)


### Bug Fixes

* **deploy:** remove server-dry-run helm extension to prevent possible bug ([f77a8c0](https://www.github.com/werf/werf/commit/f77a8c00d608f5e0447f9d2f15182e68799ca0da))

### [1.2.88](https://www.github.com/werf/werf/compare/v1.2.87...v1.2.88) (2022-04-15)


### Features

* **custom-tags:** add %image_content_based_tag% shortcut ([efd1072](https://www.github.com/werf/werf/commit/efd1072d1959aac824830a2128a49e47f4efb615))
* **export:** add %image_content_based_tag% shortcut ([7122ee9](https://www.github.com/werf/werf/commit/7122ee97863b4b1f0fc57415e0879b035df607f1))
* **stapel-to-buildah:** git archive stage implementation ([328b033](https://www.github.com/werf/werf/commit/328b0338f586636ba6eee6a361aab58c787d083b))
* **stapel-to-buildah:** implemented dependencies checksum using buildah container backend ([9596f6d](https://www.github.com/werf/werf/commit/9596f6db61db4da88c3ff58ebafffacde871191d))
* **stapel-to-buildah:** support git patches related stages ([79f71c1](https://www.github.com/werf/werf/commit/79f71c1890aba92ea9130e55afb5e5df87539815))


### Bug Fixes

* **kube-run:** didn't work in Native Buildah mode ([db1fec6](https://www.github.com/werf/werf/commit/db1fec686351f360174f45876d2155986d35cd50))
* **tests:** fix ansible suite, change deprecated base image ([bdb6c9c](https://www.github.com/werf/werf/commit/bdb6c9c3b5d25ef07e5453c852f190b33c62d5ce))

### [1.2.87](https://www.github.com/werf/werf/compare/v1.2.86...v1.2.87) (2022-04-08)


### Bug Fixes

* **slugification:** release name can contain dots ([766610b](https://www.github.com/werf/werf/commit/766610b433e93fc643ada27bc7fc06b1e5b561a0))

### [1.2.86](https://www.github.com/werf/werf/compare/v1.2.85...v1.2.86) (2022-04-08)


### Bug Fixes

* **server-dry-run:** possible fix for 'unable to recognize ...: no matches for kind ... in version ...' (part 2) ([e053dad](https://www.github.com/werf/werf/commit/e053dadade90a614c336e2c9d868bb21c51b87a1))

### [1.2.85](https://www.github.com/werf/werf/compare/v1.2.84...v1.2.85) (2022-04-08)


### Features

* **stapel-to-buildah:** basic implementation of dependencies* stages ([9ead236](https://www.github.com/werf/werf/commit/9ead236edb0164c1b620c646085d1e114f73001c))


### Bug Fixes

* **buildah:** use crun instead of runc ([fbae777](https://www.github.com/werf/werf/commit/fbae77711aaae0c8ace8513626a7494348de8acc))
* **server-dry-run:** possible fix for 'unable to recognize ...: no matches for kind ... in version ...' ([5b13270](https://www.github.com/werf/werf/commit/5b132701d44ba6e6481e4410cfd608960afb8ab8))

### [1.2.84](https://www.github.com/werf/werf/compare/v1.2.83...v1.2.84) (2022-04-05)


### Bug Fixes

* **slugification:** kubernetes namespace and release name cannot contain dots ([e22eecb](https://www.github.com/werf/werf/commit/e22eecb91ce04d5af10c83cbed21db6518008d9c))

### [1.2.83](https://www.github.com/werf/werf/compare/v1.2.82...v1.2.83) (2022-04-04)


### Bug Fixes

* **cleanup:** manage custom tags that do not have associated existent stages ([ef6efc3](https://www.github.com/werf/werf/commit/ef6efc3718a61248d08ee8230c347245d073ff0d))
* ignoring broken config in container registry ([50ed5c7](https://www.github.com/werf/werf/commit/50ed5c7e1fe1b6c6767162e9c0be2714cac83518))

### [1.2.82](https://www.github.com/werf/werf/compare/v1.2.81...v1.2.82) (2022-04-01)


### Bug Fixes

* **dependencies:** broken imports checksum when files names contain spaces ([57ea901](https://www.github.com/werf/werf/commit/57ea901900710a40dae90065e7687241b68f68b4))

### [1.2.81](https://www.github.com/werf/werf/compare/v1.2.80...v1.2.81) (2022-04-01)


### Features

* **stapel-to-buildah:** support user stages and mounts ([da55b2a](https://www.github.com/werf/werf/commit/da55b2ae823c16ed3e2ea62cbe34e830f2a78d05))


### Bug Fixes

* **cleanup:** fail on getting manifests for some custom tag metadata ([90a3767](https://www.github.com/werf/werf/commit/90a3767c6847bc8782a0302c0fa58171811ad1e1))
* **stapel-to-buildah:** working build of 'from' stage ([91527db](https://www.github.com/werf/werf/commit/91527db132ff822428a97a7a05fa6427c3150f04))

### [1.2.80](https://www.github.com/werf/werf/compare/v1.2.79...v1.2.80) (2022-03-30)


### Features

* **kube-run:** add --kube-config-base64 ([a32cd4f](https://www.github.com/werf/werf/commit/a32cd4fd0c0c735b9c2a9b66548734da078ba6ce))
* **kubectl:** add --tmp-dir, --home-dir, --kubeconfig-base64 ([cddc6b6](https://www.github.com/werf/werf/commit/cddc6b620f380f1407ecb268a80dbef914ac0373))
* **stapel-to-buildah:** implement 'from' stage ([7cc7d71](https://www.github.com/werf/werf/commit/7cc7d71ccc2ec03aa1f039610d7d7011d6b1e1f2))


### Bug Fixes

* **kube-run:** broken --docker-config ([60b74b8](https://www.github.com/werf/werf/commit/60b74b8ea302aa9322c11745a07fb89e7404cdcc))

### [1.2.79](https://www.github.com/werf/werf/compare/v1.2.78...v1.2.79) (2022-03-23)


### Features

* **kube-run:** --auto-pull-secret provides private registry access for pod ([d94104f](https://www.github.com/werf/werf/commit/d94104f4ef88200f77ce0b4a4eefe9a6c4f5d0b6))
* **kube-run:** add --kube-config, fix --kube-context opts ([8014d98](https://www.github.com/werf/werf/commit/8014d982bc62c62e5fc031bc99ad32dbf70ef2c5))
* **kubectl:** respect a few global $WERF_* env vars ([a2d523e](https://www.github.com/werf/werf/commit/a2d523eeb29ceaa2c9ae7d92b8f12ecde1a8609a))


### Bug Fixes

* **cleanup:** fail when no kubernetes configs available and no --without-kube option specified ([14de74f](https://www.github.com/werf/werf/commit/14de74f57d85bc79a6d5408c70ce183d25554c4a))
* **docs:** update cli reference ([7f65ca2](https://www.github.com/werf/werf/commit/7f65ca2d62fd448954af8739a04c6f8400d982e4))
* **docs:** update cli reference ([ad3a705](https://www.github.com/werf/werf/commit/ad3a705bc662c6bfec9f4e69a2896fc7961cf36c))
* **docs:** update cli reference ([588eb2d](https://www.github.com/werf/werf/commit/588eb2d5aea35d0c6e3bece766f34c7b83232bfc))
* **kube-run:** temporarily disable --kube-config* opts ([352a0bd](https://www.github.com/werf/werf/commit/352a0bd883438723314671c24bbae18c68113657))
* **server-dry-run:** fix "admission webhook ... does not support dry-run" ([5b118f4](https://www.github.com/werf/werf/commit/5b118f406256b16b68a78044e79d18fc1e952c6a))

### [1.2.78](https://www.github.com/werf/werf/compare/v1.2.77...v1.2.78) (2022-03-21)


### Features

* kubectl exposed via `werf kubectl` ([c6435d8](https://www.github.com/werf/werf/commit/c6435d82a954076a7b76abca342ef78b7419dc44))
* new command `werf kube-run` ([3da4449](https://www.github.com/werf/werf/commit/3da444970dc1c92f36c9747849804991df343d6f))


### Bug Fixes

* "unable to get docker system info" error when container runtime not used ([93b6f5a](https://www.github.com/werf/werf/commit/93b6f5a0cd07865d208049ac1455a9da1386ad78))
* **build:** do not store images into final repo when --skip-build is set ([69e1bb0](https://www.github.com/werf/werf/commit/69e1bb0783700586fc902f4367a283fcdf391b6f))
* context extraction error ([d19cfb6](https://www.github.com/werf/werf/commit/d19cfb6f6f60f4ca5ea952cd794c5ac9d439dbaf))
* **deploy:** fix --set-file giving []uint{} array intead of string ([aa3aa4e](https://www.github.com/werf/werf/commit/aa3aa4ea2e842ac52b4f59aaccf2e8a938c6a19f))
* temporarily disable broken server-dry-run ([e648787](https://www.github.com/werf/werf/commit/e648787f499a24529ea91295574cdbc8bc16fa00))

### [1.2.77](https://www.github.com/werf/werf/compare/v1.2.76...v1.2.77) (2022-03-17)


### Features

* **build:** speeding up with runtime caching for meta images ([7ea0a4c](https://www.github.com/werf/werf/commit/7ea0a4cd92427567b34734e036c66821b37d2a14))
* **build:** speeding up with runtime caching for stages ([a13a7b0](https://www.github.com/werf/werf/commit/a13a7b0e610c06db072133931208f0bbc644b127))
* **cleanup/purge:** speeding up with runtime caching for stages ([cbb31b2](https://www.github.com/werf/werf/commit/cbb31b29778a94bd0a900eee021a4967c626217b))


### Bug Fixes

* **purge:** fix final repo stages deletion ([11ed6f7](https://www.github.com/werf/werf/commit/11ed6f7fcb61b6b7e5481eadc1e0b60a7b189072))

### [1.2.76](https://www.github.com/werf/werf/compare/v1.2.75...v1.2.76) (2022-03-15)


### Bug Fixes

* default ~/.ssh/id_rsa key not loaded ([2c186fe](https://www.github.com/werf/werf/commit/2c186fe7405422d5f37be1f37eaedcebbcfb7ae4))

### [1.2.75](https://www.github.com/werf/werf/compare/v1.2.74...v1.2.75) (2022-03-14)


### Features

* **ssh-key:** support passphrases for --ssh-key options ([9ed3c96](https://www.github.com/werf/werf/commit/9ed3c9629a167caceef75c347b1eb288947b3ed5))


### Bug Fixes

* broken --ssh-key option ([c389259](https://www.github.com/werf/werf/commit/c38925991b3b78dfadf3ca11b837abc5bcf15ecd))
* server dry-run validation breaks helm 2 to 3 transition ([0450171](https://www.github.com/werf/werf/commit/0450171502fb49d271ef09f5f90144cfc7a9ce89))

### [1.2.74](https://www.github.com/werf/werf/compare/v1.2.73...v1.2.74) (2022-03-10)


### Bug Fixes

* **deploy:** fix server side validation false positive failure case ([b64b8bb](https://www.github.com/werf/werf/commit/b64b8bb1af715f0ea868618fbc7b889a0b0ffe01))

### [1.2.73](https://www.github.com/werf/werf/compare/v1.2.72...v1.2.73) (2022-03-04)


### Features

* **deploy:** support server side validation in converge/dismiss commands ([6df39c9](https://www.github.com/werf/werf/commit/6df39c956a25ccd81fea8bc4cde12443f8207e06))
* update to helm v3.8.0 ([6fff511](https://www.github.com/werf/werf/commit/6fff511f8646baae15a1ee6ee1bf1ee6c1e43c36))

### [1.2.72](https://www.github.com/werf/werf/compare/v1.2.71...v1.2.72) (2022-02-25)


### Bug Fixes

* **buildah:** support Dockerfile builder target param to build specific stage ([44bc718](https://www.github.com/werf/werf/commit/44bc71810b79379132555bc4bfe022f7863fbabb))

### [1.2.71](https://www.github.com/werf/werf/compare/v1.2.70...v1.2.71) (2022-02-24)


### Features

* **cleanup:** optimization of cleaning images which are used when importing ([1b82a47](https://www.github.com/werf/werf/commit/1b82a476af8d97f4e4abd0852fc4635ad1aac63d))


### Bug Fixes

* add werf-cleanup command warning when no kube configs available ([e87261b](https://www.github.com/werf/werf/commit/e87261bcfb39cebe4a6d707641579305c3ed84b6))
* WERF_KUBE_CONFIG and WERF_KUBECONFIG environment variables not working ([b0615b0](https://www.github.com/werf/werf/commit/b0615b0ab583a9788d39c786a31cc251336fe945))

### [1.2.70](https://www.github.com/werf/werf/compare/v1.2.69...v1.2.70) (2022-02-21)


### Bug Fixes

* **helm:** don't add annotations and labels to *List Kinds ([4f2d029](https://www.github.com/werf/werf/commit/4f2d02906f64d9e13760b58f1ecbe4377057357a))
* panic when auto host cleanup runs in some werf commands ([a7064ff](https://www.github.com/werf/werf/commit/a7064ff89a4104853e84033ad851dd52aa9f2667))

### [1.2.69](https://www.github.com/werf/werf/compare/v1.2.68...v1.2.69) (2022-02-18)


### Bug Fixes

* possible error during worktree switch procedure due to lost error handling ([82b1770](https://www.github.com/werf/werf/commit/82b177070086085a872e566cbf49ea2696f74604))

### [1.2.68](https://www.github.com/werf/werf/compare/v1.2.67...v1.2.68) (2022-02-18)


### Features

* **cleanup:** cleaning up artifacts by git history-based policy as well as images ([04404a3](https://www.github.com/werf/werf/commit/04404a3a6b7d1c59cd93fe9e8cfe9e6a8158be16))


### Bug Fixes

* **build:** werf does not reset stages storage cache when import source image not found ([262412a](https://www.github.com/werf/werf/commit/262412ad3c908d900970cf428a1a454fc82e703c))
* host-cleanup procedure not running in gitlab-ci ([a78df7c](https://www.github.com/werf/werf/commit/a78df7c27d71b3d8be0efd0412fe02a19b0a3068))
* **host-cleanup:** host cleanup not working without --docker-server-storage-path option ([dfa159c](https://www.github.com/werf/werf/commit/dfa159cba16be7ecf0ca73764f9b57e0d5f3f38c))
* more correct handling of storage.ErrBrokenImage ([fbbdd54](https://www.github.com/werf/werf/commit/fbbdd5471eaae770cdf5976bbd54979459d9b2af))

### [1.2.67](https://www.github.com/werf/werf/compare/v1.2.66...v1.2.67) (2022-02-15)


### Features

* **bundle:** new command "werf bundle render" ([ad0181e](https://www.github.com/werf/werf/commit/ad0181ea0adc787c81dd9a8047f3ef3e825801b5))


### Bug Fixes

* "unable to switch worktree" in gitlab ([fe6c2d4](https://www.github.com/werf/werf/commit/fe6c2d49a45ea8cd8b7c343b3c7eee51fb05ad42))

### [1.2.66](https://www.github.com/werf/werf/compare/v1.2.65...v1.2.66) (2022-02-14)


### Features

* **config:** dependency graph for werf.yaml images ([403bcb0](https://www.github.com/werf/werf/commit/403bcb05174a8d3fa199d0453b16e009ff51fc72))


### Bug Fixes

* **cleanup:** odd warning message with a nonexistent tag ([e873376](https://www.github.com/werf/werf/commit/e873376bc5a8c355e15e9cdf7feb444bf392eb43))

### [1.2.65](https://www.github.com/werf/werf/compare/v1.2.64...v1.2.65) (2022-02-08)


### Bug Fixes

* **dev-mode:** dev branch breaking on complex merge conflicts ([a628ce6](https://www.github.com/werf/werf/commit/a628ce60bad29bf080ee98efa6586f95f444e680))

### [1.2.64](https://www.github.com/werf/werf/compare/v1.2.63...v1.2.64) (2022-02-07)


### Features

* **buildah:** update buildah subsystem to v1.24.1 ([f0f3816](https://www.github.com/werf/werf/commit/f0f38165689662984b3232ea710d66d934f22da1))
* **dev-mode:** less rebuilds due to better cache handling ([34df9d2](https://www.github.com/werf/werf/commit/34df9d279ef9ed6c3676588ac91fe526f919588f))

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
