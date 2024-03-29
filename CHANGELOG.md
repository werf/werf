# Changelog

### [1.2.301](https://www.github.com/werf/werf/compare/v1.2.300...v1.2.301) (2024-03-29)


### Features

* **deps:** mini bump all modules ([8802bd7](https://www.github.com/werf/werf/commit/8802bd7c9557a817249b56c5f3310cf8b9121db4))


### Bug Fixes

* **buildah:** "could not find netavark" error ([0503170](https://www.github.com/werf/werf/commit/0503170d918988bd0b32a74a89660668e9b9db5f))
* **buildah:** "unknown version specified" error ([d475c14](https://www.github.com/werf/werf/commit/d475c1422d62f1ed4fbff4d23ad003b3daecd74a))
* **buildah:** unable to find "pasta" binary ([858933f](https://www.github.com/werf/werf/commit/858933f6c5f66541af0dcb4fee452dd73cccf277))
* **deps:** update all direct dependencies ([48641a0](https://www.github.com/werf/werf/commit/48641a09010a39065a80df754e39dd73dda53e24))
* **deps:** update Helm to v3.14.3 ([049f682](https://www.github.com/werf/werf/commit/049f682d9710224ebb6f0c3ec46c2f402f2ae6c4))
* e2e converge tests after Nelm migration ([216d363](https://www.github.com/werf/werf/commit/216d3636eabe6d4b19701efc93cb3a2ba2191c2b))
* **nelm:** --status-progress-period=-1 panics ([aa152f6](https://www.github.com/werf/werf/commit/aa152f6b84803a16ad9f6df4b400b289d09a7876))
* **nelm:** don't show logs produced before current release ([7067534](https://www.github.com/werf/werf/commit/70675344d7a6db8f2421b18f3839254a48c25767))
* **nelm:** Jobs not failing on errors /2 ([688b760](https://www.github.com/werf/werf/commit/688b760ac21f4ac6851a87e3194a6da8a915fa47))
* **nelm:** tracking might hang with werf.io/track-termination-mode: "NonBlocking" ([c0fda6b](https://www.github.com/werf/werf/commit/c0fda6b4cd76207be5b145ddb5d00838be73d42d))
* **stapel:** copying suitable stage from secondary might break reproducibility of subsequent stages ([96dddad](https://www.github.com/werf/werf/commit/96dddad4713f15fe6f179a1b4d05490e1ad2a49e))
* **stapel:** discarding newly built image might break reproducibility of subsequent stages ([78d2905](https://www.github.com/werf/werf/commit/78d2905ca1efc52bf5608372d4bed38a9b40b65f))
* **stapel:** using suitable stage might break reproducibility of subsequent stages ([f0a618c](https://www.github.com/werf/werf/commit/f0a618c38d55a26f40c480aed4502d257dd8b101))

### [1.2.300](https://www.github.com/werf/werf/compare/v1.2.299...v1.2.300) (2024-03-21)


### Features

* make werf root command embeddable ([2e1c9d3](https://www.github.com/werf/werf/commit/2e1c9d3c88351a7ff7c816c5d0cdcf55b9a8c5f2))


### Bug Fixes

* **nelm:** Jobs not failing on errors /2 ([96793f0](https://www.github.com/werf/werf/commit/96793f0671c441f54f5ba98c7953b6d4c9909051))

### [1.2.299](https://www.github.com/werf/werf/compare/v1.2.298...v1.2.299) (2024-03-21)


### Bug Fixes

* **nelm:** Jobs not failing on errors ([ca069a2](https://www.github.com/werf/werf/commit/ca069a2500796ccb06ca3c01767200cfc7ba3550))

### [1.2.298](https://www.github.com/werf/werf/compare/v1.2.297...v1.2.298) (2024-03-20)


### Bug Fixes

* **nelm:** WERF_DIR ignored for chart path ([5acfaa3](https://www.github.com/werf/werf/commit/5acfaa3b1075a9dfa264ff2883e8d43d1f66e36f))

### [1.2.297](https://www.github.com/werf/werf/compare/v1.2.296...v1.2.297) (2024-03-19)


### Features

* **nelm:** improve log coloring ([cc5c397](https://www.github.com/werf/werf/commit/cc5c397a71ad324070dc2e6610ab508dd3947e00))
* **nelm:** remove excessive operations summary ([8a295df](https://www.github.com/werf/werf/commit/8a295dfa6522b9c7aaedbfd2ee6212ef297398b4))
* **nelm:** show NOTES.txt on release ([3c64bd9](https://www.github.com/werf/werf/commit/3c64bd9c78e2a2894aebb702a09df1bd79fa55ba))
* **nelm:** show release name/namespace on converge/plan start ([eb85b19](https://www.github.com/werf/werf/commit/eb85b19fd85affbaec584472b3dd097231d769f8))


### Bug Fixes

* **nelm:** panic with werf.io/replicas-on-creation ([49043e3](https://www.github.com/werf/werf/commit/49043e3849e01d56be80e645a5ddecbe7c65a785))
* **nelm:** removing child chart values with null ([9fcc639](https://www.github.com/werf/werf/commit/9fcc639cc02d9fa22035cbc9f518682b280a85e8))

### [1.2.296](https://www.github.com/werf/werf/compare/v1.2.295...v1.2.296) (2024-03-05)


### Features

* **nelm:** activate new deployment engine Nelm by default ([be67fc1](https://www.github.com/werf/werf/commit/be67fc134c864e39fecc3e15896db9d2a8084cf6))

### [1.2.295](https://www.github.com/werf/werf/compare/v1.2.294...v1.2.295) (2024-02-29)


### Bug Fixes

* **nelm:** deployment tracking hangs; old pods tracked ([583017e](https://www.github.com/werf/werf/commit/583017e0cc67cf8572f902cedc7444b62437d7f4))
* **ssh:** initialize ssh agent before working with git repo ([caf1422](https://www.github.com/werf/werf/commit/caf142255c5d65cfa405f94acc94889ac7332453))

### [1.2.294](https://www.github.com/werf/werf/compare/v1.2.293...v1.2.294) (2024-02-19)


### Bug Fixes

* **deploy:** dont strip non-whitespace after --- in manifests ([1b4c7f2](https://www.github.com/werf/werf/commit/1b4c7f2ac8598ec0d464a022e577fcb1d7138653))

### [1.2.293](https://www.github.com/werf/werf/compare/v1.2.292...v1.2.293) (2024-02-19)


### Bug Fixes

* more robust manifest splitting ([340c92b](https://www.github.com/werf/werf/commit/340c92b82a919c6ce5c1a320467071147f76ed1d))
* **nelm:** corrupted revision number in errors messages ([e95fb4f](https://www.github.com/werf/werf/commit/e95fb4fe65c421cd7282d796f294641297a7ee96))

### [1.2.292](https://www.github.com/werf/werf/compare/v1.2.291...v1.2.292) (2024-02-15)


### Features

* add --atomic, --timeout, --skip-dependencies-repo-refresh for "bundle apply" ([b29c290](https://www.github.com/werf/werf/commit/b29c29018c527813e6d82b70c1cd9da52d496412))

### [1.2.291](https://www.github.com/werf/werf/compare/v1.2.290...v1.2.291) (2024-02-15)


### Bug Fixes

* **build:** image metadata publication when git history based policy disabled ([11850a8](https://www.github.com/werf/werf/commit/11850a88d9367b0a40b4f752a88a8b7b5fbd6c69))
* **bundle:** support for --insecure-registry/--skip-tls-verify-registry options ([96123cb](https://www.github.com/werf/werf/commit/96123cbd3fd3741699e9a7514ffc5fda875de2dd))
* **cleanup:** git unshallow and sync with remote when git history based policy disabled ([d1078a0](https://www.github.com/werf/werf/commit/d1078a099e4657e7433d0c7bf1d0e2557cd5e8fe))
* **cleanup:** image metadata cleanup when git history based policy disabled ([3a709ea](https://www.github.com/werf/werf/commit/3a709ea53dd033d6fbe8e82147624b6fff8ac0eb))
* **nelm:** cycle error if pre+post hook in release ([8c40fdb](https://www.github.com/werf/werf/commit/8c40fdb80c6c3e309b6c166f4a50db1a27e354df))

### [1.2.290](https://www.github.com/werf/werf/compare/v1.2.289...v1.2.290) (2024-02-14)


### Features

* **nelm:** save plan graph on plan build error ([c931364](https://www.github.com/werf/werf/commit/c9313645e85eb196fa608b445c0f4e4a3be01e64))


### Bug Fixes

* **nelm:** panic on replicas null ([513700e](https://www.github.com/werf/werf/commit/513700e094cc99078aea35141f3b3fbbd27236a2))

### [1.2.289](https://www.github.com/werf/werf/compare/v1.2.288...v1.2.289) (2024-02-08)


### Bug Fixes

* --wait-for-jobs did not use Kubedog ([69fac2c](https://www.github.com/werf/werf/commit/69fac2c722a86855fe53cd473d8cdf08e9e34518))

### [1.2.288](https://www.github.com/werf/werf/compare/v1.2.287...v1.2.288) (2024-02-07)


### Bug Fixes

* **nelm:** controllers not Readying; no logs for STS/DS ([84b5054](https://www.github.com/werf/werf/commit/84b505446b85ac46d951bddc73271c86112d4d5c))

### [1.2.287](https://www.github.com/werf/werf/compare/v1.2.286...v1.2.287) (2024-02-02)


### Bug Fixes

* **cleanup:** issue with incorrect exclusion of ancestors for retained images ([#5956](https://www.github.com/werf/werf/issues/5956)) ([00ee0e1](https://www.github.com/werf/werf/commit/00ee0e1d19c3d5511c641c8ab8794d84a6f01655))
* **nelm:** resources always rendered in no-cluster mode ([289cf14](https://www.github.com/werf/werf/commit/289cf1456ef81f5dcb517ad70ea94283b3afc195))

### [1.2.286](https://www.github.com/werf/werf/compare/v1.2.285...v1.2.286) (2024-02-01)


### Bug Fixes

* dont try to create namespace if exists ([88c0610](https://www.github.com/werf/werf/commit/88c06104196da552fafc2a64c05dcf803d60b961))

### [1.2.285](https://www.github.com/werf/werf/compare/v1.2.284...v1.2.285) (2024-02-01)


### Bug Fixes

* **nelm:** pods not ready in progress report when they are ([89efd58](https://www.github.com/werf/werf/commit/89efd58329b0869b106ff27ce02b874f08d00a69))

### [1.2.284](https://www.github.com/werf/werf/compare/v1.2.283...v1.2.284) (2024-01-26)


### Features

* **nelm:** --auto-rollback support for converge ([5826586](https://www.github.com/werf/werf/commit/58265865a1eafad6e408a8baedea9f3269888134))


### Bug Fixes

* **nelm:** refactor error handling in converge ([fd50797](https://www.github.com/werf/werf/commit/fd5079742ee00335998e09ea2542435a1f3909db))
* **nelm:** refactor failure deploy plan in converge ([51aa423](https://www.github.com/werf/werf/commit/51aa4237eeeedb7df6f7719151f0edb4273af490))
* **nelm:** refactor pending release check in converge ([00e0c0f](https://www.github.com/werf/werf/commit/00e0c0fbc23dc766c2df9d83d4439f4aa57eb8cf))

### [1.2.283](https://www.github.com/werf/werf/compare/v1.2.282...v1.2.283) (2024-01-25)


### Features

* **compose:** use host docker binary instead of docker-compose ([#5944](https://www.github.com/werf/werf/issues/5944)) ([a82a665](https://www.github.com/werf/werf/commit/a82a665091cba4e0f64f8a11c4d3006496867c90))
* **nelm:** heavily improved resource tracking ([0a23121](https://www.github.com/werf/werf/commit/0a23121447ad25641e00726cb6caa13fda4df383))

### [1.2.282](https://www.github.com/werf/werf/compare/v1.2.281...v1.2.282) (2024-01-23)


### Bug Fixes

* **cleanup:** panic: runtime error: invalid memory address or nil pointer dereference ([#5937](https://www.github.com/werf/werf/issues/5937)) ([33465ab](https://www.github.com/werf/werf/commit/33465abb68e64a9b869cb16db8df56a6a3d0ea3d))

### [1.2.281](https://www.github.com/werf/werf/compare/v1.2.280...v1.2.281) (2024-01-18)


### Features

* **nelm:** --exit-code flag for werf plan ([a9dcf1b](https://www.github.com/werf/werf/commit/a9dcf1b27f4fd71ef68b47f3b635aa1237c4ca0c))


### Bug Fixes

* update helm, nelm modules ([90d07de](https://www.github.com/werf/werf/commit/90d07debe795bf05b8b2f2d3eb32a69eb42bb686))

### [1.2.280](https://www.github.com/werf/werf/compare/v1.2.279...v1.2.280) (2024-01-16)


### Bug Fixes

* disable dependabot ([76998e9](https://www.github.com/werf/werf/commit/76998e9b479e5ab7f523608713b2758545a01abd))
* switch to go 1.21 ([2624b32](https://www.github.com/werf/werf/commit/2624b32cbf49a8437aa5d1886960b76ade2e8295))

### [1.2.279](https://www.github.com/werf/werf/compare/v1.2.278...v1.2.279) (2024-01-15)


### Bug Fixes

* update all go modules ([6cd8ff5](https://www.github.com/werf/werf/commit/6cd8ff5195ed9b23f3cab436ce463f58ab42b207))
* update more modules, get rid of minio ([09a77d2](https://www.github.com/werf/werf/commit/09a77d2100d078b53c0f5fa0078001e8cb50b1a8))

### [1.2.278](https://www.github.com/werf/werf/compare/v1.2.277...v1.2.278) (2024-01-10)


### Bug Fixes

* **deps:** update go-git ([ff00de8](https://www.github.com/werf/werf/commit/ff00de8ea7ddd887bcfbec224fbc548b3f5e2f2a))
* **deps:** update helm-plugin-utils ([608941e](https://www.github.com/werf/werf/commit/608941eef3c8aae0bc1c8e8f6fd4010b1acde6c0))

### [1.2.277](https://www.github.com/werf/werf/compare/v1.2.276...v1.2.277) (2023-12-29)


### Bug Fixes

* add WERF_NELM var (same as WERF_EXPERIMENTAL_DEPLOY_ENGINE) ([455d293](https://www.github.com/werf/werf/commit/455d29377b47be07efb100ed3528ce9a1689e8d9))
* Nelm moved to separate repo ([be662c3](https://www.github.com/werf/werf/commit/be662c3577318b3ee145eafa7161e276d6ff00f7))

### [1.2.276](https://www.github.com/werf/werf/compare/v1.2.275...v1.2.276) (2023-12-29)


### Bug Fixes

* pre-collapsed gitlab section not usable ([#5897](https://www.github.com/werf/werf/issues/5897)) ([68040b1](https://www.github.com/werf/werf/commit/68040b17a35cd9888d07eabde414c050877715df))

### [1.2.275](https://www.github.com/werf/werf/compare/v1.2.274...v1.2.275) (2023-12-22)


### Bug Fixes

* **build:** files from git with spaces in path not removed on install/beforeSetup/setup patch-stages ([6c052db](https://www.github.com/werf/werf/commit/6c052db1c250cfb5cf5e0259849e15856700034a))

### [1.2.274](https://www.github.com/werf/werf/compare/v1.2.273...v1.2.274) (2023-12-21)


### Bug Fixes

* **buildah:** fix digest change on rebuild of install/beforeSetup/setup stage when using multiple git-mappings in the same image ([52db306](https://www.github.com/werf/werf/commit/52db30637f090b6a721a66d1bdab8c7c48ebe8ad))

### [1.2.273](https://www.github.com/werf/werf/compare/v1.2.272...v1.2.273) (2023-12-21)


### Bug Fixes

* **exp-engine:** "no matches for kind" when deploying CR for CRD from crds/ ([5c68186](https://www.github.com/werf/werf/commit/5c68186be8b10cdc33e7fef045be10df9a57dce5))

### [1.2.272](https://www.github.com/werf/werf/compare/v1.2.271...v1.2.272) (2023-12-19)


### Features

* improve generic resource tracking ([c366a72](https://www.github.com/werf/werf/commit/c366a722681fd7cf348ec58bb42384464e47592a))


### Bug Fixes

* **staged-dockerfile:** explanatory error message when building unsupported Dockerfile conf ([a884912](https://www.github.com/werf/werf/commit/a884912608bbc0b0f0521e31a61dc3bd2407e7d3))

### [1.2.271](https://www.github.com/werf/werf/compare/v1.2.270...v1.2.271) (2023-12-14)


### Bug Fixes

* **secrets:** empty secrets values yaml file results in werf error ([7fe6b2b](https://www.github.com/werf/werf/commit/7fe6b2b77eb7702e2e764e7592d3fad5d4696daf))

### [1.2.270](https://www.github.com/werf/werf/compare/v1.2.269...v1.2.270) (2023-12-04)


### Features

* **container:** implement generic Retry-After header handling ([#5867](https://www.github.com/werf/werf/issues/5867)) ([b2a6022](https://www.github.com/werf/werf/commit/b2a60225a4f8d06429eb57cc3fb8b3a34e2addd3))


### Bug Fixes

* **secrets:** fix double decryption error when decrypting a yaml with multiple references to the same anchor ([a7df8cc](https://www.github.com/werf/werf/commit/a7df8cc98eecb1732e540e07ae9ec28b617b6e1d))

### [1.2.269](https://www.github.com/werf/werf/compare/v1.2.268...v1.2.269) (2023-11-10)


### Bug Fixes

* **werf-in-image:** unable to update tuf meta ([#5846](https://www.github.com/werf/werf/issues/5846)) ([fd41be1](https://www.github.com/werf/werf/commit/fd41be173d885ced78dedda7bebca6f32cc0bda5))

### [1.2.268](https://www.github.com/werf/werf/compare/v1.2.267...v1.2.268) (2023-11-10)


### Features

* **logboek:** pre-collapse GitLab log section ([#5842](https://www.github.com/werf/werf/issues/5842)) ([2b68426](https://www.github.com/werf/werf/commit/2b68426a8f081c3e90eb39503a9dc25a2b96d899))
* **logs:** add --log-time and --log-time-format options ([#5844](https://www.github.com/werf/werf/issues/5844)) ([d83bd20](https://www.github.com/werf/werf/commit/d83bd20bd132c4ba1caa4d833dbe003535f5bf49))

### [1.2.267](https://www.github.com/werf/werf/compare/v1.2.266...v1.2.267) (2023-10-19)


### Bug Fixes

* add yq to werf images ([7d23125](https://www.github.com/werf/werf/commit/7d23125432ba84b78d09db7e6252365a1526bf2d))

### [1.2.266](https://www.github.com/werf/werf/compare/v1.2.265...v1.2.266) (2023-10-16)


### Bug Fixes

* **werf-in-image:** detected dubious ownership in git repository ([#5827](https://www.github.com/werf/werf/issues/5827)) ([2ae94b4](https://www.github.com/werf/werf/commit/2ae94b452ad83930aed44db4646b78a0d88ada31))

### [1.2.265](https://www.github.com/werf/werf/compare/v1.2.264...v1.2.265) (2023-10-13)


### Features

* option to disable registry login in "werf ci-env" ([170d068](https://www.github.com/werf/werf/commit/170d06833ee17c6a26654bcba07b99b5c61643fe))


### Bug Fixes

* forbid "werf run" in Buildah mode ([09fb5bd](https://www.github.com/werf/werf/commit/09fb5bd58a0570d595a66b7b99b5601d44e86ee0))

### [1.2.264](https://www.github.com/werf/werf/compare/v1.2.263...v1.2.264) (2023-10-13)


### Features

* add more tools to werf images; update werf images ([9ca24e7](https://www.github.com/werf/werf/commit/9ca24e77ff245a9452b2dbd44eb79d668ebec7dc))
* configurable giterminism config path ([b852a34](https://www.github.com/werf/werf/commit/b852a347713b98da87b49383765f39341988f005))


### Bug Fixes

* "bundle copy --from" fails after "ci-env" ([6c8ed39](https://www.github.com/werf/werf/commit/6c8ed397f965c0a249bd88269e51ce19b7e61cb8))
* **build:** initialize ondemandKubeInitializer for custom sync server in Kubernetes ([e08a901](https://www.github.com/werf/werf/commit/e08a9014c4c9d75444fac01004e774580d82bd32))
* tone down --loose-giterminism warning ([938a9ab](https://www.github.com/werf/werf/commit/938a9ab8c3199b30ebd8c397b33fa426fe060803))

### [1.2.263](https://www.github.com/werf/werf/compare/v1.2.262...v1.2.263) (2023-09-28)


### Bug Fixes

* **exp-engine:** panic connecting internal dependencies ([b881bed](https://www.github.com/werf/werf/commit/b881bedf875b3d7b35aa171144a905a5ae89778f))

### [1.2.262](https://www.github.com/werf/werf/compare/v1.2.261...v1.2.262) (2023-09-28)


### Bug Fixes

* **exp-engine:** .dot graph saved even if not requested ([60b31ee](https://www.github.com/werf/werf/commit/60b31eecd0698ae964f8e74f93e2d505decfaf99))
* **exp-engine:** dependencies not connecting to Apply operations ([06efd9a](https://www.github.com/werf/werf/commit/06efd9a72ddf357506a3935e6629c85e88d52067))

### [1.2.261](https://www.github.com/werf/werf/compare/v1.2.260...v1.2.261) (2023-09-27)


### Bug Fixes

* **exp-engine:** helm hooks with multiple pre/post conditions always skipped ([5d0db6d](https://www.github.com/werf/werf/commit/5d0db6d130ebbd101043c8adb900656b08dad85c))

### [1.2.260](https://www.github.com/werf/werf/compare/v1.2.259...v1.2.260) (2023-09-25)


### Features

* **exp-engine:** hide sensitive Secret's diff in werf plan ([27192fd](https://www.github.com/werf/werf/commit/27192fd0b2fd66fc25c2553c15f984bd0a53dacb))


### Bug Fixes

* **parallel:** cancel context for active tasks when task fails ([#5800](https://www.github.com/werf/werf/issues/5800)) ([1d062fd](https://www.github.com/werf/werf/commit/1d062fd54e125ff145ef7d113f4beab9f2fe3a9f))

### [1.2.259](https://www.github.com/werf/werf/compare/v1.2.258...v1.2.259) (2023-09-21)


### Bug Fixes

* don't add useless "name" label when creating release namespace ([9621115](https://www.github.com/werf/werf/commit/9621115593ef436e95086a13c6ad68daa9a0e6df))
* **exp-engine:** cosmetics ([37888c6](https://www.github.com/werf/werf/commit/37888c6951607bbf7e83e1d826c2c7a075ef3f45))
* **exp-engine:** steal managed fields from any manager with prefix "werf" ([73cfb21](https://www.github.com/werf/werf/commit/73cfb219e1b1f18b5861069b21048c41cfaa47b4))

### [1.2.258](https://www.github.com/werf/werf/compare/v1.2.257...v1.2.258) (2023-09-21)


### Bug Fixes

* **exp-engine:** divide by zero error if no deployable resources; field adoption didn't work if live resource up to date ([b2e7742](https://www.github.com/werf/werf/commit/b2e774289c8170e557f3e6ea70577a38c4eb4ec3))

### [1.2.257](https://www.github.com/werf/werf/compare/v1.2.256...v1.2.257) (2023-09-21)


### Bug Fixes

* **exp-engine:** replace strategic patch with merge patch ([ba7a9df](https://www.github.com/werf/werf/commit/ba7a9dfe62f565d9ff19be696352e28418c92227))

### [1.2.256](https://www.github.com/werf/werf/compare/v1.2.255...v1.2.256) (2023-09-20)


### Features

* **ci-env:** add ci.werf.io/tag automatic annotation ([#5789](https://www.github.com/werf/werf/issues/5789)) ([b36586f](https://www.github.com/werf/werf/commit/b36586f616d19cccf16bfedd7bd0ec4c4b077704))
* complete rework of managed fields handling + minor fixes ([2d41d32](https://www.github.com/werf/werf/commit/2d41d3258881ef0b467d017f9dabd7d1653c20e6))
* **exp-engine:** "werf plan" command, "werf converge" improvements ([f3a2b0b](https://www.github.com/werf/werf/commit/f3a2b0bec4e9de792a55f1c4e8ab1b3c9ca06766))
* **exp-engine:** complete rework of managed fields handling + ([2d41d32](https://www.github.com/werf/werf/commit/2d41d3258881ef0b467d017f9dabd7d1653c20e6))


### Bug Fixes

* hide CRDs create diffs from plan ([2d41d32](https://www.github.com/werf/werf/commit/2d41d3258881ef0b467d017f9dabd7d1653c20e6))
* remove "..." from log messages ([2d41d32](https://www.github.com/werf/werf/commit/2d41d3258881ef0b467d017f9dabd7d1653c20e6))
* show "insignificant changes" in plan if filtered resource diff is ([2d41d32](https://www.github.com/werf/werf/commit/2d41d3258881ef0b467d017f9dabd7d1653c20e6))
* unreliable deletion of resources removed from release due to race ([327f1fd](https://www.github.com/werf/werf/commit/327f1fde2aceb2a9be43191c9fdcf9779ab1bc84))

### [1.2.255](https://www.github.com/werf/werf/compare/v1.2.254...v1.2.255) (2023-09-13)


### Bug Fixes

* **exp-engine:** minor deploy engine improvements ([afa07ab](https://www.github.com/werf/werf/commit/afa07ab6c841cb71ff5b0842e0e1513b27902671))

### [1.2.254](https://www.github.com/werf/werf/compare/v1.2.253...v1.2.254) (2023-09-13)


### Features

* **exp-engine:** experimental deploy engine v2 with graph ([192e179](https://www.github.com/werf/werf/commit/192e1790cdcfd0df8e5839720bbe5c3aedd0b0bb))

### [1.2.253](https://www.github.com/werf/werf/compare/v1.2.252...v1.2.253) (2023-08-22)


### Bug Fixes

* **exp-engine:** parallel GET/dry-APPLY finishing before results received ([0237621](https://www.github.com/werf/werf/commit/0237621493a0204bc84631d1f84ab621f6f983d5))

### [1.2.252](https://www.github.com/werf/werf/compare/v1.2.251...v1.2.252) (2023-08-18)


### Bug Fixes

* cannot deep copy *annotation.AnnotationReplicasOnCreation ([475824a](https://www.github.com/werf/werf/commit/475824ad1a21a3f699dfa6c2a99aad9dec9c43eb))

### [1.2.251](https://www.github.com/werf/werf/compare/v1.2.250...v1.2.251) (2023-08-16)


### Bug Fixes

* **render:** missing WERF_KUBE_VERSION env ([810f5d3](https://www.github.com/werf/werf/commit/810f5d3abef33a19ec058281edfb57e18c54a993))

### [1.2.250](https://www.github.com/werf/werf/compare/v1.2.249...v1.2.250) (2023-08-16)


### Bug Fixes

* **render:** add --kube-version flag ([bc48dd0](https://www.github.com/werf/werf/commit/bc48dd07e8f74c183ea0efa836c144e720de6330))

### [1.2.249](https://www.github.com/werf/werf/compare/v1.2.248...v1.2.249) (2023-08-03)


### Bug Fixes

* **exp-engine:** major refactor: new Resource(s), Release, History, ResourcePreparer, KubeClient classes ([6b9dcb2](https://www.github.com/werf/werf/commit/6b9dcb2ee8b6d90d728be9f26411d889876930bd))

### [1.2.248](https://www.github.com/werf/werf/compare/v1.2.247...v1.2.248) (2023-07-24)


### Bug Fixes

* **exp-engine:** decouple deploy from ActionConfig ([a4b0850](https://www.github.com/werf/werf/commit/a4b0850b83e3203f8357c05db6a528ff2f239c7e))
* **exp-engine:** use new History api ([1ca6109](https://www.github.com/werf/werf/commit/1ca610904115c08f798b88597f458ea1b9c357be))
* **exp-engine:** use new Waiter API ([3361a3d](https://www.github.com/werf/werf/commit/3361a3d04ee139384318afa7143b4b444e4a3368))
* **exp-engine:** use updated Helm API ([e6c8c1a](https://www.github.com/werf/werf/commit/e6c8c1ab5d1dfb688ffe295b746aeda818458ba2))

### [1.2.247](https://www.github.com/werf/werf/compare/v1.2.246...v1.2.247) (2023-07-18)


### Bug Fixes

* **exp-engine:** use new ChartTree api ([5ed24fd](https://www.github.com/werf/werf/commit/5ed24fd46938ecee092de2fb8b7d1b0528741037))

### [1.2.246](https://www.github.com/werf/werf/compare/v1.2.245...v1.2.246) (2023-07-18)


### Bug Fixes

* **exp-engine:** panic on external-dependency namespace annotation ([be555f7](https://www.github.com/werf/werf/commit/be555f7e5f4616588c0f7cc7b4aafadffc550f1c))

### [1.2.245](https://www.github.com/werf/werf/compare/v1.2.244...v1.2.245) (2023-07-18)


### Bug Fixes

* **converge:** remove redundant chart downloading on Helm level ([c97e5d4](https://www.github.com/werf/werf/commit/c97e5d4c2f7ec5206acbe487e2abb26d170452f0))

### [1.2.244](https://www.github.com/werf/werf/compare/v1.2.243...v1.2.244) (2023-07-18)


### Bug Fixes

* move helm deploy action to converge ([33cd546](https://www.github.com/werf/werf/commit/33cd54644a82e593de0842c624d738b59b168738))

### [1.2.243](https://www.github.com/werf/werf/compare/v1.2.242...v1.2.243) (2023-07-10)


### Bug Fixes

* **telemetry:** experimental deploy engine attribute ([753e129](https://www.github.com/werf/werf/commit/753e1299d685fd042bb9dec35b4b59179d915218))

### [1.2.242](https://www.github.com/werf/werf/compare/v1.2.241...v1.2.242) (2023-07-04)


### Bug Fixes

* **dev:** precommit git hooks are not ignored ([76fb7ab](https://www.github.com/werf/werf/commit/76fb7ab28d5811b68c25235f4bfe8ae66e5b8fda))

### [1.2.241](https://www.github.com/werf/werf/compare/v1.2.240...v1.2.241) (2023-06-13)


### Bug Fixes

* **staged-dockerfile:** eliminate excess manifest get request from base image registry ([3103aff](https://www.github.com/werf/werf/commit/3103affb8e3e6cea272f99970586a1221103b70c))

### [1.2.240](https://www.github.com/werf/werf/compare/v1.2.239...v1.2.240) (2023-06-06)


### Bug Fixes

* **ci:** unlabel job should not fail ([#5670](https://www.github.com/werf/werf/issues/5670)) ([3f4267a](https://www.github.com/werf/werf/commit/3f4267ad2b3cedefb557097f6e07d926473d23fd))
* **custom-tags:** no way to tag only certain images ([54ad8a5](https://www.github.com/werf/werf/commit/54ad8a5e285ab46b3c5315a9e87b8800ac47fbee))
* integration tests should not fail ([#5669](https://www.github.com/werf/werf/issues/5669)) ([c4e411a](https://www.github.com/werf/werf/commit/c4e411a0855014e12697808b76e6b5da70931c7f))

### [1.2.239](https://www.github.com/werf/werf/compare/v1.2.238...v1.2.239) (2023-05-29)


### Bug Fixes

* hide build log in export command ([#5658](https://www.github.com/werf/werf/issues/5658)) ([88bc502](https://www.github.com/werf/werf/commit/88bc5024117b9dc5dfb30f59e2b6f72e786ed5db))
* **kubedog:** resource hangs on context canceled ([0ff8176](https://www.github.com/werf/werf/commit/0ff81768864e1fe380dba00086b39d3525d81c06))
* remove abandoned linters, use unused linter ([#5661](https://www.github.com/werf/werf/issues/5661)) ([adbf2c7](https://www.github.com/werf/werf/commit/adbf2c707e5104f18605e8d723f03d3c7bdadc59))

### [1.2.238](https://www.github.com/werf/werf/compare/v1.2.237...v1.2.238) (2023-05-24)


### Bug Fixes

* **deploy:** new engine no activity timeout for hooks ([9dab75d](https://www.github.com/werf/werf/commit/9dab75dfe728778238c9cd29ada86dfe93f0e66d))

### [1.2.237](https://www.github.com/werf/werf/compare/v1.2.236...v1.2.237) (2023-05-23)


### Bug Fixes

* **buildah:** use native-chroot isolation by default with buildah backend ([fdfc558](https://www.github.com/werf/werf/commit/fdfc558bd9863340bffbdf3f5cbd613ca9277871))

### [1.2.236](https://www.github.com/werf/werf/compare/v1.2.235...v1.2.236) (2023-05-23)


### Features

* hide build logs if --require-built-images is in use, show logs on error ([3a7506c](https://www.github.com/werf/werf/commit/3a7506c6787335b8b903e96cffaaf0aa5399fab8))


### Bug Fixes

* **buildah:** push-layers logs in buildah only in verbose mode ([f46abbc](https://www.github.com/werf/werf/commit/f46abbc01bb1f919d18eb9239a2ee353b017e789))
* **staged-dockerfile:** separate FROM stage with caching in the container-registry and possibility to reset by global cache version ([dd3d653](https://www.github.com/werf/werf/commit/dd3d653361c8b6f9d539bc276888d7415683d0a1))

### [1.2.235](https://www.github.com/werf/werf/compare/v1.2.234...v1.2.235) (2023-05-18)


### Bug Fixes

* **deploy:** add debug for new deploy engine ([7481265](https://www.github.com/werf/werf/commit/74812653ef762564ea5ed8e852bb3a29c55366f5))

### [1.2.234](https://www.github.com/werf/werf/compare/v1.2.233...v1.2.234) (2023-05-18)


### Features

* new experimental deploy engine ([8d431c2](https://www.github.com/werf/werf/commit/8d431c2a2c194c806496ccd28df85500cd43428d))


### Bug Fixes

* **bundles/publish:** fix usage of CreateNewBundle without env variables. ([974e01a](https://www.github.com/werf/werf/commit/974e01a59d9f7b34ddd93d87da32b9d7f88fe251))
* **bundles/publish:** fix usage of CreateNewBundle without env variables. ([3e0e079](https://www.github.com/werf/werf/commit/3e0e0798251061f7777e6f55829ca8db19b526fb))

### [1.2.233](https://www.github.com/werf/werf/compare/v1.2.232...v1.2.233) (2023-05-16)


### Bug Fixes

* **staged-dockerfile:** optimize stages dependencies tree builder ([bc3ac92](https://www.github.com/werf/werf/commit/bc3ac92a1a93eef0f74616fbd54849b623a76f52))

### [1.2.232](https://www.github.com/werf/werf/compare/v1.2.231...v1.2.232) (2023-05-16)


### Features

* **dev:** tasks for local development ([#5607](https://www.github.com/werf/werf/issues/5607)) ([5b96afc](https://www.github.com/werf/werf/commit/5b96afc8c1f0e1f3275d5ff1a0a36f8ddb39eef2))
* **multiarch:** support platform setting per image in werf.yaml configuration ([39fd752](https://www.github.com/werf/werf/commit/39fd7524482b2f40b3c8e3df50775a7d82681c6a))


### Bug Fixes

* harbor regular NOT_FOUND error treated as 'broken image' internal registry error ([bc4ef3d](https://www.github.com/werf/werf/commit/bc4ef3db6b9b5395612f5418c1f435ed0b1e7732))
* **multiarch:** use correct multiarch manifests for werf-run and werf-kube-run commands ([fca96f2](https://www.github.com/werf/werf/commit/fca96f23f1acaf9dffe927d6cf84647b76855187))
* rename ambiguous --skip-build to --require-built-images ([#5619](https://www.github.com/werf/werf/issues/5619)) ([2a57b4b](https://www.github.com/werf/werf/commit/2a57b4b780e4bd63660d83044fbe3ab73557359c))
* use 'built image' instead 'cache image' ([fee0d67](https://www.github.com/werf/werf/commit/fee0d67f887e0b7ef95ebe3872e78b3b905adec0))

### [1.2.231](https://www.github.com/werf/werf/compare/v1.2.230...v1.2.231) (2023-05-02)


### Features

* **multiarch:** working werf-export command in multiarch mode ([a699230](https://www.github.com/werf/werf/commit/a6992306d500511c4b29145e7387d37b5adf9fc6))


### Bug Fixes

* restore data-parent attribute ([#5600](https://www.github.com/werf/werf/issues/5600)) ([ecb72d2](https://www.github.com/werf/werf/commit/ecb72d234d08395a4d796a3d3c5117f927fd0c26))

### [1.2.230](https://www.github.com/werf/werf/compare/v1.2.229...v1.2.230) (2023-04-28)


### Bug Fixes

* **docs:** cleanup docs moved to install ([4afcf03](https://www.github.com/werf/werf/commit/4afcf032cdd56fc5dc266e65033c07717cb8b62c))
* **logging:** doubling in build summary block with several sets ([59b7bf5](https://www.github.com/werf/werf/commit/59b7bf5b6502cf02e0212188ed3dbacdb4240d5e))
* speedup docs development ([2dc150a](https://www.github.com/werf/werf/commit/2dc150abfc3219529ae9f3879762b8f18cbb0bb6))

### [1.2.229](https://www.github.com/werf/werf/compare/v1.2.228...v1.2.229) (2023-04-27)


### Bug Fixes

* **docs:** fix configurator link ([924dcb6](https://www.github.com/werf/werf/commit/924dcb6488d7366fe1e9964281b0fd2bc9530b65))

### [1.2.228](https://www.github.com/werf/werf/compare/v1.2.227...v1.2.228) (2023-04-27)


### Bug Fixes

* **docs:** change configurator urls ([12d35e0](https://www.github.com/werf/werf/commit/12d35e0face6de9da77cbd02d746aa4455adae91))
* **multiarch:** custom tags not working ([ce34af1](https://www.github.com/werf/werf/commit/ce34af19b9bef43e21f6335282db14d3126af170))
* revert precompile stanza ([#5572](https://www.github.com/werf/werf/issues/5572)) ([6d26b79](https://www.github.com/werf/werf/commit/6d26b7968060c9e20445e609455c46638a4b114f))
* update docs ([5ced8f2](https://www.github.com/werf/werf/commit/5ced8f2e4a78ff2ff54b900b0cfc85b4fda6a9ec))

### [1.2.227](https://www.github.com/werf/werf/compare/v1.2.226...v1.2.227) (2023-04-26)


### Bug Fixes

* **stapel/imports:** processing of includePaths/excludePaths with globs in file/directory name ([7046ca7](https://www.github.com/werf/werf/commit/7046ca70d7b8b8fab10160cb05df8098c3de186a))
* update ruby gems ([e51e7b6](https://www.github.com/werf/werf/commit/e51e7b676e580fb78157e1a00a806228108bea8f))

### [1.2.226](https://www.github.com/werf/werf/compare/v1.2.225...v1.2.226) (2023-04-21)


### Features

* **multiarch:** support cleanup of images built in multiarch mode ([64b50e8](https://www.github.com/werf/werf/commit/64b50e8182def90063bf196885381a3f73cdfe30))

### [1.2.225](https://www.github.com/werf/werf/compare/v1.2.224...v1.2.225) (2023-04-20)


### Features

* **multiarch:** support for final-repo ([0c632fb](https://www.github.com/werf/werf/commit/0c632fb8e05da276d3576cf8fd475efc27d8af46))
* **multiarch:** support secondary-repo and cache-repo ([d7df554](https://www.github.com/werf/werf/commit/d7df554330b3cc07b3393625681d7ff801ef6b26))

### [1.2.224](https://www.github.com/werf/werf/compare/v1.2.223...v1.2.224) (2023-04-13)


### Features

* **multiarch:** minimal docs about multiplatform mode ([f0579be](https://www.github.com/werf/werf/commit/f0579be5b2f0b44285377f93d114d4b91ce9a7fa))

### [1.2.223](https://www.github.com/werf/werf/compare/v1.2.222...v1.2.223) (2023-04-12)


### Features

* **buildah:** enable :local mode for buildah backend ([d1e400d](https://www.github.com/werf/werf/commit/d1e400dd15e0bf964c7a0e2347062494f7f9a73f))
* **local-stages-storage:** introduce local storage independent of container backend implementation ([e6aa7f1](https://www.github.com/werf/werf/commit/e6aa7f14453be1eb2c5f0032a22255d5edf408bb))
* **multiarch:** support :local mode multiarch building for docker server backend ([e519902](https://www.github.com/werf/werf/commit/e519902272eb3262f146213a6906996be4677c11))

### [1.2.222](https://www.github.com/werf/werf/compare/v1.2.221...v1.2.222) (2023-04-10)


### Bug Fixes

* **werf-builder:** update werf-builder image ([1171fc5](https://www.github.com/werf/werf/commit/1171fc5766ca45bcb9eb6755e887f8300d5ef0a7))

### [1.2.221](https://www.github.com/werf/werf/compare/v1.2.220...v1.2.221) (2023-04-07)


### Bug Fixes

* updated builder image ([0f7f5d6](https://www.github.com/werf/werf/commit/0f7f5d6aa8aeec166d4c119bf5b24381b24a9a5f))

### [1.2.220](https://www.github.com/werf/werf/compare/v1.2.219...v1.2.220) (2023-04-07)


### Features

* **multiarch:** implement creation of manifest list for each werf image in multiplatform mode builds ([5406416](https://www.github.com/werf/werf/commit/5406416dbc74a05fa70994b0cd5cb9f1ba3a2268))
* **multiarch:** use multiplatform images for converge/render and in the build report ([370d14c](https://www.github.com/werf/werf/commit/370d14c4050c3e0758da2cf090288bfa5078eb7e))


### Bug Fixes

* fix self-signed-registry write operations when skip-tls-verify flag used ([94f2b42](https://www.github.com/werf/werf/commit/94f2b42419a4b21437edc74787a7fe18cf1471f9))
* **logs:** fix typo in deprecation warning ([3b2a9b0](https://www.github.com/werf/werf/commit/3b2a9b0c99555d03db05abb4d28f93a70199e04c))
* **multiarch:** managed images adding bug fix ([285e5a6](https://www.github.com/werf/werf/commit/285e5a6d69f5707ae313c1ad2bbe37d2602ae01b))
* **multiarch:** managed images adding bug fix (part 2) ([e685e5e](https://www.github.com/werf/werf/commit/e685e5e68798d1e6c1d84666b79492fa4f00e983))
* **multiarch:** publish git metadata for multiplatform mode images ([1913c95](https://www.github.com/werf/werf/commit/1913c95f95d1676bf4e6b51ce1694d31674d713f))
* **multiarch:** use buildx instead of deprecated DOCKER_BUILDKIT to enable buildkit building ([5334baa](https://www.github.com/werf/werf/commit/5334baa42e352b2a12645dcd1fcd2ef5ea4f7f74))
* werf-in-a-user-namespace not found, cannot start auto host cleanup procedure ([f7ecd7d](https://www.github.com/werf/werf/commit/f7ecd7dccd20c619cce9c4c257b9dbe175302ae3))

### [1.2.219](https://www.github.com/werf/werf/compare/v1.2.218...v1.2.219) (2023-03-29)


### Bug Fixes

* 'certificate signed by unknown authority' and not working skip-tls-verify-registry param ([b646359](https://www.github.com/werf/werf/commit/b646359b80f729159bff6fe1616e3e1e2799be4e))
* **multiarch:** do not override image metadata for secondary platforms ([b49060e](https://www.github.com/werf/werf/commit/b49060ef993d9e9c0c9ee31d49a2562807679e1a))
* **multiarch:** do not override image metadata for secondary platforms (part 2) ([838baef](https://www.github.com/werf/werf/commit/838baef7fa48fd67b8907d68a668e72265066c03))
* restart release-please process ([63f4072](https://www.github.com/werf/werf/commit/63f4072298c43e593e29225b06ba6609247fe762))

### [1.2.218](https://www.github.com/werf/werf/compare/v1.2.217...v1.2.218) (2023-03-28)


### Bug Fixes

* 'exec: werf-in-a-user-namespace: executable file not found in $PATH' when using buildah ([6323d8e](https://www.github.com/werf/werf/commit/6323d8efd55d9f5e15ff2d5ed6e99eeb4a449f6f))
* **multiplatform:** images report contains correct digests ([836fe04](https://www.github.com/werf/werf/commit/836fe044abb1f2a71bd4033112bbd09ee3890b8a))
* **staged-dockerfile:** allow scratch base image ([5bf6a27](https://www.github.com/werf/werf/commit/5bf6a27afcb100d107dc00b5258315515004ded4))

### [1.2.217](https://www.github.com/werf/werf/compare/v1.2.216...v1.2.217) (2023-03-23)


### Bug Fixes

* **buildah:** usage of docker.exactValues affects digest the same way for buildah and docker-server backends ([726ef94](https://www.github.com/werf/werf/commit/726ef9454b8070bf37a09273a034fd9856c5c123))

### [1.2.216](https://www.github.com/werf/werf/compare/v1.2.215...v1.2.216) (2023-03-22)


### Bug Fixes

* **multiarch:** fix panic which occurs when using stapel import from certain stage ([431673f](https://www.github.com/werf/werf/commit/431673fd7930672ca44203b574eb667f5fe32f8b))

### [1.2.215](https://www.github.com/werf/werf/compare/v1.2.214...v1.2.215) (2023-03-22)


### Features

* **multiarch:** add support for target platform in container backends ([22ae3cf](https://www.github.com/werf/werf/commit/22ae3cfd6bdff851822515f216808c73def9f0c9))


### Bug Fixes

* **multiarch:** fix 'werf stage image' command panic related to multiarch refactor ([9eeb9d4](https://www.github.com/werf/werf/commit/9eeb9d4478cff547117fe61f52e4c1d2d25fef94))
* **staged-dockerfile:** meaningful message about staged: true available only for buildah backend and not avaiable for docker server backend ([44485e2](https://www.github.com/werf/werf/commit/44485e2fc18ac393c86740d5109c86fdbf35e966))

### [1.2.214](https://www.github.com/werf/werf/compare/v1.2.213...v1.2.214) (2023-03-17)


### Bug Fixes

* fix(deps): update ([1578b7c](https://github.com/werf/werf/commit/1578b7cf1c8ddc385d59479842d7497c7188281f))

### [1.2.213](https://www.github.com/werf/werf/compare/v1.2.212...v1.2.213) (2023-03-17)


### Bug Fixes

* **ci-env:** existing envs for adding extra annotations not overwritten ([dbaf2a3](https://www.github.com/werf/werf/commit/dbaf2a3eac46898e7fc53d6620f25561f62763a3))

### [1.2.212](https://www.github.com/werf/werf/compare/v1.2.211...v1.2.212) (2023-03-15)


### Bug Fixes

* downgrade minio ([d016fbf](https://www.github.com/werf/werf/commit/d016fbf957fb1b52b2aa0b6a2551ea39ec747bbe))

### [1.2.211](https://www.github.com/werf/werf/compare/v1.2.210...v1.2.211) (2023-03-15)


### Bug Fixes

* update all dependencies ([cc19916](https://www.github.com/werf/werf/commit/cc19916f9a7cdb1ad5ce1d5275b01ccd893eba0e))

### [1.2.210](https://www.github.com/werf/werf/compare/v1.2.209...v1.2.210) (2023-03-14)


### Bug Fixes

* **export:** erroneous export of intermediate images ([24d5f9f](https://www.github.com/werf/werf/commit/24d5f9f09e3e0a9db17da90bb27abb2a0aca1362))
* **multiarch:** fix bug with import servers names collision introduced during refactor ([8f251dc](https://www.github.com/werf/werf/commit/8f251dc85e509e549290ba5928d7ba506ef3d181))

### [1.2.209](https://www.github.com/werf/werf/compare/v1.2.208...v1.2.209) (2023-03-13)


### Bug Fixes

* **multiarch:** fix werf converge, kube-run, run and other commands when platform param specified ([3e2add1](https://www.github.com/werf/werf/commit/3e2add1b35f57194a8a51083fdc2a8d042f0462e))

### [1.2.208](https://www.github.com/werf/werf/compare/v1.2.207...v1.2.208) (2023-03-13)


### Features

* **export:** add --add-label option ([46afee2](https://www.github.com/werf/werf/commit/46afee2256ab8d188b0a2dff4fe42511343f550d))
* **multiarch:** introduced ability to specify multiple platforms ([f0efcb7](https://www.github.com/werf/werf/commit/f0efcb73517f42e7374682eec9028b76db3f8b34))


### Bug Fixes

* docs/Gemfile & docs/Gemfile.lock to reduce vulnerabilities ([aca4bd1](https://www.github.com/werf/werf/commit/aca4bd1ee1fdeb8b54a2c083205507346ab4a3ec))
* playground/Dockerfile.ghost to reduce vulnerabilities ([8d4cca9](https://www.github.com/werf/werf/commit/8d4cca999d582fbb33ac6393f26d98527fdde1d5))

### [1.2.207](https://www.github.com/werf/werf/compare/v1.2.206...v1.2.207) (2023-03-09)


### Bug Fixes

* update go + go modules ([3c7f2d3](https://www.github.com/werf/werf/commit/3c7f2d3a8d272f9078b8a14d2f012d066e92bb4e))
* update werf builder image ([d803468](https://www.github.com/werf/werf/commit/d803468ee0948f3f552de2a792c5398275065ff6))

### [1.2.206](https://www.github.com/werf/werf/compare/v1.2.205...v1.2.206) (2023-03-09)


### Features

* **cleanup:** more logging for saved images ([0bda54d](https://www.github.com/werf/werf/commit/0bda54de45211ad33fbdbf254d821caf4cfcb310))

### [1.2.205](https://www.github.com/werf/werf/compare/v1.2.204...v1.2.205) (2023-03-02)


### Features

* **bundles:** allow usage of bundles with included secret-values as oci chart dependencies ([469678c](https://www.github.com/werf/werf/commit/469678cd71aa97ab5a147164f3949fc3a0c5863b))

### [1.2.204](https://www.github.com/werf/werf/compare/v1.2.203...v1.2.204) (2023-02-28)


### Bug Fixes

* **kube-run:** command stderr was redirected to stdin ([4d038d4](https://www.github.com/werf/werf/commit/4d038d4a2a7bb88329c43d1f0234ea73879330eb))

### [1.2.203](https://www.github.com/werf/werf/compare/v1.2.202...v1.2.203) (2023-02-28)


### Features

* **bundles:** support custom secret values files when publishing bundle ([5290e33](https://www.github.com/werf/werf/commit/5290e337da68935f4732056e5b33c0b63e2a26cd))


### Bug Fixes

* **staged-dockerfile:** fix panic which occurs when using dependencies between images with multistages ([4f6b10e](https://www.github.com/werf/werf/commit/4f6b10e13c60bce18f06df8e12df6c364d7575f2))

### [1.2.202](https://www.github.com/werf/werf/compare/v1.2.201...v1.2.202) (2023-02-21)


### Features

* **bundles:** enable secrets for bundle publish and apply ([5d6dec7](https://www.github.com/werf/werf/commit/5d6dec768cc46d58349b9d46c69f2b5938793a53))


### Bug Fixes

* **helm-dependencies:** automatically fill ~/.werf/local_cache on 'werf helm dependency update' command ([b094521](https://www.github.com/werf/werf/commit/b094521630e61ce987269c13c3152e225ffebba9))
* **helm-dependencies:** enable loading of .helm/charts/CHART-VERSION.tgz charts ([3addc23](https://www.github.com/werf/werf/commit/3addc23805c113759b7af9e0b29cafbdd2b9c6d2))

### [1.2.201](https://www.github.com/werf/werf/compare/v1.2.200...v1.2.201) (2023-02-16)


### Bug Fixes

* **dismiss:** allow --namespace or --release if git repo present ([68f0f14](https://www.github.com/werf/werf/commit/68f0f14caaa6fd270a001e965c664acc07ab0458))

### [1.2.200](https://www.github.com/werf/werf/compare/v1.2.199...v1.2.200) (2023-02-13)


### Features

* add --deploy-report-path, --build-report-path ([7fa1d81](https://www.github.com/werf/werf/commit/7fa1d81eef47bf95165a684ca1932489e9510c4c))
* **dismiss:** add two ways to run without git ([f2e1d16](https://www.github.com/werf/werf/commit/f2e1d16d382ff7c18b6350437f0100d20ec15c3f))


### Bug Fixes

* **build:** 'unsupported MediaType' error when using quay base images ([27b572d](https://www.github.com/werf/werf/commit/27b572d3b731d8ea5eb0846854eb38cd9bde804f))
* **build:** TOOMANYREQUESTS error occurs for the built images ([163961d](https://www.github.com/werf/werf/commit/163961d4022457d8ce8b3cc0e44e1807665d0089))
* **bundles:** --helm-compatible-chart and --rename-chart options for 'bundle copy' and 'bundle publish' ([3333d03](https://www.github.com/werf/werf/commit/3333d031d87825b74e7dce451cb2ad15908ca6ea))
* **compose:** redundant image building with compose down command ([e94e7c4](https://www.github.com/werf/werf/commit/e94e7c49bc2219b84b0ba4812e98f3224411abf7))
* rework build/deploy report options ([40a8e81](https://www.github.com/werf/werf/commit/40a8e8155f36d94d54abe4386bf25b6b5a315faa))
* **staged-dockerfile:** fix multiple stages with the same name from multiple Dockerfiles ([76f654d](https://www.github.com/werf/werf/commit/76f654d2ee871e16e1d787b741e9cdc133c8f82c))

### [1.2.199](https://www.github.com/werf/werf/compare/v1.2.198...v1.2.199) (2023-01-27)


### Features

* added flag to skip helm repo index fetching ([39d004a](https://www.github.com/werf/werf/commit/39d004a78814782ac859520135276148af8dc8c2))


### Bug Fixes

* **logging:** simplify logs, remove excess messages ([1d75db1](https://www.github.com/werf/werf/commit/1d75db11720bb542847a7f42d99a5bd7f68bc029))

### [1.2.198](https://www.github.com/werf/werf/compare/v1.2.197...v1.2.198) (2023-01-20)


### Bug Fixes

* **staged-dockerfile:** correction for ENV and ARG instructions handling ([7a17fc7](https://www.github.com/werf/werf/commit/7a17fc7d57ba84d5597a480781cd51a576a9d6e2))

### [1.2.197](https://www.github.com/werf/werf/compare/v1.2.196...v1.2.197) (2023-01-18)


### Bug Fixes

* **dependencies:** introduce ImageDigest mode, hide ImageID mode ([cc352fd](https://www.github.com/werf/werf/commit/cc352fdf017f2ce52d44fbe6105a381bea7fcbc8))

### [1.2.197](https://www.github.com/werf/werf/compare/v1.2.196...v1.2.197) (2023-01-16)


### Bug Fixes

* **dependencies:** introduce ImageDigest mode, hide ImageID mode

### [1.2.196](https://www.github.com/werf/werf/compare/v1.2.195...v1.2.196) (2023-01-16)


### Features

* **bundle:** allow non strict bundle publishing ([96fd4a1](https://www.github.com/werf/werf/commit/96fd4a17a17104ed56b42430f4110d9aa5f956d5))

### [1.2.195](https://www.github.com/werf/werf/compare/v1.2.194...v1.2.195) (2022-12-28)


### Bug Fixes

* **docs:** actualize sidebar in usage docs ([922d945](https://www.github.com/werf/werf/commit/922d945de0045c486e40c02ded14a0b3996cad98))

### [1.2.194](https://www.github.com/werf/werf/compare/v1.2.193...v1.2.194) (2022-12-27)


### Features

* **docs:** new article for build chapter: storage layout (lang:ru) ([9afe73c](https://www.github.com/werf/werf/commit/9afe73c3ce46459481fdfdc90a292d1715dcbc75))
* **docs:** new article for build chapter: build process (lang:ru) ([7c332ee](https://www.github.com/werf/werf/commit/7c332ee5031fbd5255ee659969747c359bb0540f))
* **docs:** updated usage/build/stapel section (lang:en+ru) ([0c504a4](https://www.github.com/werf/werf/commit/0c504a41996b0d41d4e7230e706da34bacf32ab4))


### Bug Fixes

* **giterminism:** false warning about ignoring Dockerfile when using non-root project directory ([303a6e4](https://www.github.com/werf/werf/commit/303a6e4a728a5a19a40e342e90dd8eac1d7f1783))

### [1.2.193](https://www.github.com/werf/werf/compare/v1.2.192...v1.2.193) (2022-12-17)


### Features

* **docs:** build chapter: overview and configuration articles (lang:ru) ([79d6f81](https://www.github.com/werf/werf/commit/79d6f811e1a3bc37f3debd8ccc014bfb5f086028))
* **docs:** new usage build chapter structure ([78370b0](https://www.github.com/werf/werf/commit/78370b04726096bb8604093898f90da5680aa203))


### Bug Fixes

* broken render output due to lock-related message ([420824b](https://www.github.com/werf/werf/commit/420824b9e6c1f652284dd343f478f6bb9ebfbb14))
* **shallow-clone:** enable auto unshallow unless force-shallow option used ([88d5db9](https://www.github.com/werf/werf/commit/88d5db913dbecf69f84a555f02b12d9a976c6b03))

### [1.2.192](https://www.github.com/werf/werf/compare/v1.2.191...v1.2.192) (2022-12-12)


### Bug Fixes

* **report:** fix panic occured when using final-repo and report ([e62cd78](https://www.github.com/werf/werf/commit/e62cd78ad86c0736c61f5a0b773bc3c77ad2cd27))
* **staged-dockerfile:** do not store non-target Dockerfile stages in the final-repo ([a0d7838](https://www.github.com/werf/werf/commit/a0d78382d228b64990b863c806486adb44042e25))

### [1.2.191](https://www.github.com/werf/werf/compare/v1.2.190...v1.2.191) (2022-11-29)


### Bug Fixes

* fix ssh not available in registry.werf.io/werf images ([5045493](https://www.github.com/werf/werf/commit/50454937890778b2ffe146195c2ccebb505791f6))

### [1.2.190](https://www.github.com/werf/werf/compare/v1.2.189...v1.2.190) (2022-11-17)


### Features

* **staged-dockerfile:** support ONBUILD instructions (part 1, preparations) ([8a813b5](https://www.github.com/werf/werf/commit/8a813b52318f06acb83fc44aff215d95c57e8f09)), closes [#2215](https://www.github.com/werf/werf/issues/2215)


### Bug Fixes

* **build:** inconsistent report path when final-repo used ([5924702](https://www.github.com/werf/werf/commit/59247021054de511b7a08017a1c039d61fef4b50))
* **staged-dockerfile:** fix meta args always expands to empty strings ([8f6b562](https://www.github.com/werf/werf/commit/8f6b562786a7ea92d8208bb6181b621c2c9373f9))

### [1.2.189](https://www.github.com/werf/werf/compare/v1.2.188...v1.2.189) (2022-11-15)


### Features

* **completion:** add fish and powershell support ([b9d0b9d](https://www.github.com/werf/werf/commit/b9d0b9d434ca564ad44204d41ad0b852cc5d7218))

### [1.2.188](https://www.github.com/werf/werf/compare/v1.2.187...v1.2.188) (2022-11-07)


### Features

* **staged-dockerfile:** support werf images dependencies build-args ([8faf229](https://www.github.com/werf/werf/commit/8faf229ff4cf8523ae995a72bd76762cd4024833))


### Bug Fixes

* **staged-dockerfile:** changing FROM base image does not cause rebuilding ([a52991a](https://www.github.com/werf/werf/commit/a52991ae9bf9a83d2b9de11ca11ede3ea95d6297))

### [1.2.187](https://www.github.com/werf/werf/compare/v1.2.186...v1.2.187) (2022-11-02)


### Bug Fixes

* **build:** fix git-patch-* stages mechanics when patches contains binaries ([cfd6b8e](https://www.github.com/werf/werf/commit/cfd6b8e19134f0ee26b324bfc3ed4620540a0ce5))

### [1.2.186](https://www.github.com/werf/werf/compare/v1.2.185...v1.2.186) (2022-11-02)


### Features

* **staged-dockerfile:** implement first stage of build-args expansion ([c0de754](https://www.github.com/werf/werf/commit/c0de754946deddce61dc70e85ea5149466ebc86f))


### Bug Fixes

* **buildah:** broken build on mac/win ([1118613](https://www.github.com/werf/werf/commit/11186132285bc3e8dedc9f3b0ec83b34b2e6b7c6))

### [1.2.185](https://www.github.com/werf/werf/compare/v1.2.184...v1.2.185) (2022-11-01)


### Features

* **staged-dockerfile:** all dockerfile options and instructions /1 ([f0cde50](https://www.github.com/werf/werf/commit/f0cde500b55de4f96b9b47e97bf45bf9a313ad08))
* **staged-dockerfile:** implement COPY --from stage to image name expansion ([824c5bb](https://www.github.com/werf/werf/commit/824c5bb9c9defb6b9c919aef8db17dd7cda3b9c0)), closes [#2215](https://www.github.com/werf/werf/issues/2215)
* **staged-dockerfile:** implement global/Run mounts ([42edf1e](https://www.github.com/werf/werf/commit/42edf1e3b3e20b50a23b79dbac7a8e3715d8ed2b))
* **staged-dockerfile:** refactor calculateBuildContextGlobsChecksum ([e558e1e](https://www.github.com/werf/werf/commit/e558e1e0474a1a4fe272d59884a2241cefceaf63))
* **staged-dockerfile:** refactor dockerfile, use buidkit instruction structs directly ([e138887](https://www.github.com/werf/werf/commit/e1388878af8617089188722e904a3ac4e3134669))
* **staged-dockerfile:** refine ADD instruction digest calculation ([2ef3d11](https://www.github.com/werf/werf/commit/2ef3d11dc3dbbcc89b0ed9d4f52a5eb5a872889e))
* **staged-dockerfile:** refine dockerfile instructions digests calculations ([d15d79f](https://www.github.com/werf/werf/commit/d15d79f2678367decd9b96370664d48fbf8aee5a))
* **staged-dockerfile:** use contents of Copy/Add sources in checksum ([d20e397](https://www.github.com/werf/werf/commit/d20e39712f8d827479e9c4586724ff932d845660))


### Bug Fixes

* **staged-dockerfile:** broken error message ([4dea6b3](https://www.github.com/werf/werf/commit/4dea6b3b99ca7d12c99b29b410b91d9917f7e9b0))
* **staged-dockerfile:** panics on Healthcheck and Maintainer instructions and no duplicates in images-sets ([2a55266](https://www.github.com/werf/werf/commit/2a552663fad00fb020a272936879fdfbefbeb1bd)), closes [#2215](https://www.github.com/werf/werf/issues/2215)
* **staged-dockerfile:** proper instruction arguments unescaping ([e767013](https://www.github.com/werf/werf/commit/e767013db72a4aeeaf258ad21c3919dc219dc492))
* **stapel:** add log line about starting container cleanup ([e8fd1f4](https://www.github.com/werf/werf/commit/e8fd1f479861e955e64374e94ad1d8fec35baebf))

### [1.2.184](https://www.github.com/werf/werf/compare/v1.2.183...v1.2.184) (2022-10-21)


### Features

* **staged-dockerfile:** basic support of all dockerfile stages at conveyor level ([306ed6c](https://www.github.com/werf/werf/commit/306ed6c10696a2746cf7b14143a5dd4edf07a0c0))
* **staged-dockerfile:** implement whether stage uses build-context correctly ([2851923](https://www.github.com/werf/werf/commit/2851923135a4926c8ba44b85917d600ad1e03462))
* **staged-dockerfile:** map dockerfile stages with dependencies to werf internal images ([f5f200e](https://www.github.com/werf/werf/commit/f5f200eaf5bf4e5be1fd4e81ea3040c83bdc2890))


### Bug Fixes

* panic when calling SplitFilepath on windows ([78c10d2](https://www.github.com/werf/werf/commit/78c10d2f77828365beddba9b2316340fa468a83b))

### [1.2.183](https://www.github.com/werf/werf/compare/v1.2.182...v1.2.183) (2022-10-18)


### Bug Fixes

* **bundles:** fix bundle-render and bundle-apply commands could not access .Values.werf.images service values ([276dc6b](https://www.github.com/werf/werf/commit/276dc6b902ca7a45cc9eb77935f2272e7d948eec))

### [1.2.182](https://www.github.com/werf/werf/compare/v1.2.181...v1.2.182) (2022-10-17)


### Features

* **staged-dockerfile:** make and extract context archive only once ([9d72ad1](https://www.github.com/werf/werf/commit/9d72ad1edbf4d365987e8fdcc4d8b5122dbb51fd))


### Bug Fixes

* **helm:** keep all revisions if no succeeded release ([f321b52](https://www.github.com/werf/werf/commit/f321b52332e582542a90a5080f4b1c77ecf8d037))
* **staged-dockerfile:** fix docker ignore path matcher ([c4b6cd5](https://www.github.com/werf/werf/commit/c4b6cd515a902f2a098974d7a02dbc498b7ffb54))
* WERF_DISABLE_DEFAULT_SECRET_VALUES and WERF_DISABLE_DEFAULT_VALUES support for corresponding options ([647700a](https://www.github.com/werf/werf/commit/647700ac82361362600fb09901a259b34881806d))

### [1.2.181](https://www.github.com/werf/werf/compare/v1.2.180...v1.2.181) (2022-10-17)


### Features

* added options to disable usage of default values (and secret values) ([49425ee](https://www.github.com/werf/werf/commit/49425ee48e3dcacecdea61d6582a7c246809ba8e))
* **bundles:** publish .helm/files into bundle ([68c096f](https://www.github.com/werf/werf/commit/68c096fa9c74440529243f8134b46bb9d1615c4b))
* **staged-dockerfile:** add optional image-from-dockerfile reference into Image obj ([deb0827](https://www.github.com/werf/werf/commit/deb08270ce884f8daa20618b06a8147be6179ff6))
* **staged-dockerfile:** complete instructions set with all params in the dockerfile parser pkg ([06f122b](https://www.github.com/werf/werf/commit/06f122b06a47552c822c64c0345e3551cbe07e0c))
* **staged-dockerfile:** Dockerfile and DockerfileStage primitives reworked ([78e2911](https://www.github.com/werf/werf/commit/78e29110ae3cb2e2addabaaa84e9b24408e7d86e))
* **staged-dockerfile:** implement buidkit frontend instructions to dockerfile instructions conversion ([2bc6c30](https://www.github.com/werf/werf/commit/2bc6c30d4c09b6034c8651adf7fc6828d412467c))
* **staged-dockerfile:** initialize dockerfile-images with werf.yaml configration section ([186f563](https://www.github.com/werf/werf/commit/186f563fdfcea879032bb849fc907345ac6c7651)), closes [#2215](https://www.github.com/werf/werf/issues/2215)
* **staged-dockerfile:** move container backend instructions data into dockerfile parser package ([9500967](https://www.github.com/werf/werf/commit/9500967e875f6530ddd23d8b9ce3f6e2a4b37aba))

### [1.2.180](https://www.github.com/werf/werf/compare/v1.2.179...v1.2.180) (2022-10-12)


### Bug Fixes

* **bundles:** fix subcharts dependencies not published, and excess files published into the bundle ([fd15ddd](https://www.github.com/werf/werf/commit/fd15dddb43bb62a74198174ef3d2f3d22aabcc94))
* **helm:** keep all revisions since last succeeded release ([9224014](https://www.github.com/werf/werf/commit/9224014f4e1230bd5863e1a64224a8bc1c8c530c))

### [1.2.179](https://www.github.com/werf/werf/compare/v1.2.178...v1.2.179) (2022-10-11)


### Bug Fixes

* **buildah:** git add result in broken symlinks ([4013833](https://www.github.com/werf/werf/commit/4013833ee1524c19910dda0a27293a99669f6f44))

### [1.2.178](https://www.github.com/werf/werf/compare/v1.2.177...v1.2.178) (2022-10-11)


### Features

* **buildah:** add low level dockerfile stage builder ([76c98c6](https://www.github.com/werf/werf/commit/76c98c66e08faa5153c9c62af476885936836e83))
* **staged-dockerfile:** implement full chain of staged dockerfile building only for single instruction (RUN) ([121ac0c](https://www.github.com/werf/werf/commit/121ac0c2c3ad1f4074ed9111880b2789e5b3a61f))
* **staged-dockerfile:** prepare conveyor, stage and dockerfile parser for new impl ([db8d337](https://www.github.com/werf/werf/commit/db8d3373588a0c19a9415def4245c997574776f9))
* **staged-dockerfile:** refactored container backend dockerfile builder ([a210944](https://www.github.com/werf/werf/commit/a2109442d3e128593df7989192a969809573f3a5))
* **staged-dockerfile:** refactored conveyor, debug container backend staged dockerfile builder ([62b2181](https://www.github.com/werf/werf/commit/62b2181cf86e730de5996651c3e7d402dfb6ddc9))

### [1.2.177](https://www.github.com/werf/werf/compare/v1.2.176...v1.2.177) (2022-10-03)


### Features

* **staged-dockerfile:** refactor build package conveyor images tree creation ([9ecb737](https://www.github.com/werf/werf/commit/9ecb7372862d6a9c4ec88762d80c5d35c77ce6ec))


### Bug Fixes

* **converge:** feature gate for specific images params in werf-converge (due to compatibility issues) ([78c7c28](https://www.github.com/werf/werf/commit/78c7c286bbf9f873655cfe84a77540a59068645b))
* **dismiss:** fix --with-namespace not deleting namespace in dismiss command ([f0ef743](https://www.github.com/werf/werf/commit/f0ef743801b7ae57cdff1d48e0e11c2d359cb82a))

### [1.2.176](https://www.github.com/werf/werf/compare/v1.2.175...v1.2.176) (2022-09-27)


### Bug Fixes

* **buildah:** import with rename and include paths not working properly ([4d35fdb](https://www.github.com/werf/werf/commit/4d35fdb589f9bc3da67e886d168548a57e3b0840))

### [1.2.175](https://www.github.com/werf/werf/compare/v1.2.174...v1.2.175) (2022-09-23)


### Features

* **build:** support using only specific images from werf.yaml or disabling images for all werf commands ([c618043](https://www.github.com/werf/werf/commit/c618043fd5c2f53c88e60287c9b7f161b61f8901))


### Bug Fixes

* **buildah:** add support for git owner/group settings ([623ef86](https://www.github.com/werf/werf/commit/623ef86dfc8c689a81087088b10b4dde9865455f))
* **buildah:** interpret docker.HEALTHCHECK instruction same way as docker-server backend ([ebb506f](https://www.github.com/werf/werf/commit/ebb506f1810c82e4e94c43ad21ac39be6498d125))
* **helm:** fix "missing registry client" error in werf-helm-* commands ([414dd38](https://www.github.com/werf/werf/commit/414dd38d7f7f58e92f497ee40b63645c5d1b4ae1))
* **purge:** add warning about unsupported buildah backend ([14f6f1e](https://www.github.com/werf/werf/commit/14f6f1e66a1872fd3d7dc1a40d400055abea3765))

### [1.2.174](https://www.github.com/werf/werf/compare/v1.2.173...v1.2.174) (2022-09-16)


### Bug Fixes

* **helm:** empty resource annos/labels result in no service annos/labels ([902c5a1](https://www.github.com/werf/werf/commit/902c5a16169a1884ee959a7ac71b974ef5b1f5e4))

### [1.2.173](https://www.github.com/werf/werf/compare/v1.2.172...v1.2.173) (2022-09-15)


### Bug Fixes

* **dismiss:** rework uninstall-with-namespace procedure ([8657449](https://www.github.com/werf/werf/commit/865744926b9669f794114355944223373fad9951))
* **helm:** don't rely on resource Group for resources equality matching ([8e52f59](https://www.github.com/werf/werf/commit/8e52f59a19e7a7d2c81bcad966064ba34e6c2667))

### [1.2.172](https://www.github.com/werf/werf/compare/v1.2.171...v1.2.172) (2022-09-12)


### Bug Fixes

* **bundles:** bundle copy from archive to remote incorrect values ([e9a2c53](https://www.github.com/werf/werf/commit/e9a2c53887547a7b8136e808653263abacb05dad))
* **deploy:** lower releases-history-max default to 5 releases (was 10) ([7e2cc3d](https://www.github.com/werf/werf/commit/7e2cc3deb7e457608e27d3b93b6777987356717f))
* **giterminism:** --add-custom-tag option is not allowed ([8b72dfe](https://www.github.com/werf/werf/commit/8b72dfef1076a8374d89ba9def584095150e6f29))
* **run:** --bash and --shell depend on image entrypoint ([c2369f6](https://www.github.com/werf/werf/commit/c2369f6041c7db78b5cb9a94640e9b25f23f34f8))
* **run:** a container is not cleaned up after execution by default ([c04367c](https://www.github.com/werf/werf/commit/c04367c198cb8567164466758be237eb6ad546ed))

### [1.2.171](https://www.github.com/werf/werf/compare/v1.2.170...v1.2.171) (2022-09-07)


### Bug Fixes

* **buildah:** different processing of CMD/ENTRYPOINT by Stapel and Buildah backend ([97e89b0](https://www.github.com/werf/werf/commit/97e89b0075f35277c3aa2f6b828afe0ed130c984))
* **cleanup:** fallback to batch/v1beta1 when querying CronJobs for cleanup ([2f53aa4](https://www.github.com/werf/werf/commit/2f53aa4bd212ea0d05c6ec8b65ab75866f95d302))
* **telemetry:** repair turn-off telemetry switch ([f8559e9](https://www.github.com/werf/werf/commit/f8559e9f3798fcbc553b2b3596ccee637f356dc3))

### [1.2.170](https://www.github.com/werf/werf/compare/v1.2.169...v1.2.170) (2022-09-06)


### Bug Fixes

* **bundles:** deprecate bundle export/download commands in favor of new copy command abilities ([d49a81f](https://www.github.com/werf/werf/commit/d49a81f645370dda194acab278c1203076524425))
* **bundles:** refactor bundle copy implementation ([63d13d0](https://www.github.com/werf/werf/commit/63d13d019f0c7359e440a8d45f2914dcb1057f03))
* **git:** try to prevent unshallow error 'shallow file has changed since we read it' ([e51546c](https://www.github.com/werf/werf/commit/e51546c78b3460700c30823c8f3de0975f4e027b))

### [1.2.169](https://www.github.com/werf/werf/compare/v1.2.168...v1.2.169) (2022-09-02)


### Features

* **bundle:** introduce bundle archive format, implement copy command to convert archive to registry and vice versa ([345cdf0](https://www.github.com/werf/werf/commit/345cdf07e4b27abc89f72d4935b56d3acd5dc24d))

### [1.2.168](https://www.github.com/werf/werf/compare/v1.2.167...v1.2.168) (2022-09-01)


### Bug Fixes

* **deploy:** fix release-history-max param default value help message ([ff8d11a](https://www.github.com/werf/werf/commit/ff8d11aeba5f097876902a4a3e2b0f50f7331b46))

### [1.2.167](https://www.github.com/werf/werf/compare/v1.2.166...v1.2.167) (2022-09-01)


### Bug Fixes

* **cleanup:** fail on getting manifests for some custom tag metadata ([fa2c72f](https://www.github.com/werf/werf/commit/fa2c72f48bcdfcf32837f667641f85039753b458))
* **cleanup:** rare panic related to misuse of named return argument ([9fad111](https://www.github.com/werf/werf/commit/9fad1110d2b4854e4e0f1573dd3bb41ed1c2143b))
* **helm:** restore 'werf helm *' commands behavior for an env value ([528a706](https://www.github.com/werf/werf/commit/528a7064f33aab27a49192ba8497620298b88d26))

### [1.2.166](https://www.github.com/werf/werf/compare/v1.2.165...v1.2.166) (2022-08-29)


### Bug Fixes

* **bundles:** fix panic in bundle-download command ([d15d676](https://www.github.com/werf/werf/commit/d15d676c289fc78c7750d5e6036a57ea7b47120f))

### [1.2.165](https://www.github.com/werf/werf/compare/v1.2.164...v1.2.165) (2022-08-25)


### Bug Fixes

* **custom-tags:** do not store metadata in the --final-repo ([1a780c5](https://www.github.com/werf/werf/commit/1a780c519ef4553693e2a4e5c13945d723432eea))
* **helm:** use same docker-config as werf uses for helm OCI regsitry related operations ([f9bc4f3](https://www.github.com/werf/werf/commit/f9bc4f3e543374cef76c66acc36fea55a2ca914a))

### [1.2.164](https://www.github.com/werf/werf/compare/v1.2.163...v1.2.164) (2022-08-23)


### Features

* **bundle:** support publishing into insecure registries ([c88eeb3](https://www.github.com/werf/werf/commit/c88eeb318d8162f92f6be2a7bc7cccc5e8371dbb))
* **converge:** do not require to helm-repo-add repositories ([c527871](https://www.github.com/werf/werf/commit/c527871fc6920585a59c746e837362f4efa296d5))


### Bug Fixes

* **cleanup:** ignore WERF_KUBE_CONTEXT env var, support option --scan-only-context ([68677af](https://www.github.com/werf/werf/commit/68677afbaa2b949ea5ecd6fa76b704f33ebb6c1e))
* **render:** do not set empty env in werf render without repo param ([2c4bdff](https://www.github.com/werf/werf/commit/2c4bdffa816a8c8a5496facec886b4eb9a11b36c))

### [1.2.163](https://www.github.com/werf/werf/compare/v1.2.162...v1.2.163) (2022-08-19)


### Bug Fixes

* **buildah:** wrong UID/GID/workdir/entrypoint/cmd in stages ([32843f2](https://www.github.com/werf/werf/commit/32843f2898c4fd79c13e552a909418fbf7874608))

### [1.2.162](https://www.github.com/werf/werf/compare/v1.2.161...v1.2.162) (2022-08-18)


### Bug Fixes

* **stapel:** werf ignores non-zero status code ([cdd3e0a](https://www.github.com/werf/werf/commit/cdd3e0a2d9075b3feffedf980760df99f4e11d03))

### [1.2.161](https://www.github.com/werf/werf/compare/v1.2.160...v1.2.161) (2022-08-18)


### Bug Fixes

* trigger release ([b1d68d5](https://www.github.com/werf/werf/commit/b1d68d549094ee7b6917256a6f75cff445d0b8c7))

### [1.2.160](https://www.github.com/werf/werf/compare/v1.2.159...v1.2.160) (2022-08-16)


### Bug Fixes

* re-trigger release ([2618400](https://www.github.com/werf/werf/commit/26184009d193761b18fb7a6be0194046661d6986))

### [1.2.159](https://www.github.com/werf/werf/compare/v1.2.158...v1.2.159) (2022-08-16)


### Bug Fixes

* re-trigger release ([36805d7](https://www.github.com/werf/werf/commit/36805d721ca6ede31f185c1707eb93de0d0e734c))

### [1.2.158](https://www.github.com/werf/werf/compare/v1.2.157...v1.2.158) (2022-08-16)


### Bug Fixes

* re-trigger release ([0819fba](https://www.github.com/werf/werf/commit/0819fbae73b8d3966e073495bc1af5b3d3573322))

### [1.2.157](https://www.github.com/werf/werf/compare/v1.2.156...v1.2.157) (2022-08-16)


### Bug Fixes

* re-trigger release ([d69ab64](https://www.github.com/werf/werf/commit/d69ab6441a7a54ba1b0368d102b5c36bc5e5d167))

### [1.2.156](https://www.github.com/werf/werf/compare/v1.2.155...v1.2.156) (2022-08-16)


### Bug Fixes

* re-trigger release ([23194af](https://www.github.com/werf/werf/commit/23194af1fdc67c5cdf41f09acff80e71fb7b22ad))

### [1.2.155](https://www.github.com/werf/werf/compare/v1.2.154...v1.2.155) (2022-08-15)


### Miscellaneous Chores

* release v1.2.155 ([c4d14d3](https://www.github.com/werf/werf/commit/c4d14d39a0f8bafaad81717c71a7f236c3ccc380))

### [1.2.154](https://www.github.com/werf/werf/compare/v1.2.153...v1.2.154) (2022-08-14)


### Miscellaneous Chores

* release v1.2.154 ([f3d417c](https://www.github.com/werf/werf/commit/f3d417c53da596cdc66a914a94a989f55ef9c836))

### [1.2.153](https://www.github.com/werf/werf/compare/v1.2.152...v1.2.153) (2022-08-12)


### Miscellaneous Chores

* release v1.2.153 ([1854fed](https://www.github.com/werf/werf/commit/1854fed118ac67d6ec7962931b5fb43905412122))

### [1.2.152](https://www.github.com/werf/werf/compare/v1.2.151...v1.2.152) (2022-08-10)


### Bug Fixes

* **stapel:** custom LD_LIBRARY_PATH in the base image might lead to failed builds ([a06a5fc](https://www.github.com/werf/werf/commit/a06a5fc644f02a5f9a22b77e0851355914e6acc1))

### [1.2.151](https://www.github.com/werf/werf/compare/v1.2.150...v1.2.151) (2022-08-09)


### Bug Fixes

* **helm:** resource Group ignored when checking whether the same resource ([68b7594](https://www.github.com/werf/werf/commit/68b75945d68e6e736fd23459b70273b80ec288b3))

### [1.2.150](https://www.github.com/werf/werf/compare/v1.2.149...v1.2.150) (2022-08-08)


### Bug Fixes

* **buildah:** original ENTRYPOINT/CMD lost on build ([1eebc64](https://www.github.com/werf/werf/commit/1eebc6416ba6f8468c615a7dc435d997f3279569))

### [1.2.149](https://www.github.com/werf/werf/compare/v1.2.148...v1.2.149) (2022-08-08)


### Miscellaneous Chores

* release v1.2.149 ([042d690](https://www.github.com/werf/werf/commit/042d6905c97d677beb012f35485aec7265c53e61))

### [1.2.148](https://www.github.com/werf/werf/compare/v1.2.148...v1.2.148) (2022-08-08)


### Miscellaneous Chores

* release v1.2.148 ([7310d80](https://www.github.com/werf/werf/commit/7310d8026d2d2f484e81910fc6e4b3644c0c9d13))

### [1.2.148](https://www.github.com/werf/werf/compare/v1.2.147...v1.2.148) (2022-08-08)


### Features

* add Selectel CR support ([2090cd9](https://www.github.com/werf/werf/commit/2090cd9457d6157cf530d53876bdd32d97715228))

### [1.2.147](https://www.github.com/werf/werf/compare/v1.2.146...v1.2.147) (2022-08-08)


### Features

* **build:** add report Image.NAME.Rebuilt field ([be6fba7](https://www.github.com/werf/werf/commit/be6fba79ead32409cf85185323a9bad616cd40af))

### [1.2.146](https://www.github.com/werf/werf/compare/v1.2.145...v1.2.146) (2022-08-04)


### Bug Fixes

* **build:** no imagename in error in image `from` directive ([0974f3a](https://www.github.com/werf/werf/commit/0974f3ad6c302d5309f90970cdbafa0c477c15f1))
* **helm:** panic on error when applying resources ([c94cef5](https://www.github.com/werf/werf/commit/c94cef5f38ea603e19f243cb34fb857af881c3d6))

### [1.2.145](https://www.github.com/werf/werf/compare/v1.2.144...v1.2.145) (2022-08-04)


### Features

* **telemetry:** add attributes related to the usage inside CI-systems ([ec02e33](https://www.github.com/werf/werf/commit/ec02e3379a1350eaa51d7326a27e0ea3d4eaa58c))


### Bug Fixes

* **helm:** properly initialize all slice structs ([07b1e42](https://www.github.com/werf/werf/commit/07b1e42f647b093760def883613ef6e66c1547d8))

### [1.2.144](https://www.github.com/werf/werf/compare/v1.2.143...v1.2.144) (2022-07-29)


### Bug Fixes

* **kubedog:** generic: ignore jsonpath errs on Condition search ([2c2b772](https://www.github.com/werf/werf/commit/2c2b77294bb626a16fa37de3f91fe04d1ef1e9e2))

### [1.2.143](https://www.github.com/werf/werf/compare/v1.2.142...v1.2.143) (2022-07-29)


### Bug Fixes

* **helm:** install ./crds fails after dismiss ([a7ee07f](https://www.github.com/werf/werf/commit/a7ee07f85870019efb56b2fb6fa363304e6074d4))

### [1.2.142](https://www.github.com/werf/werf/compare/v1.2.141...v1.2.142) (2022-07-28)


### Features

* `tpl` performance improved ([bc28f48](https://www.github.com/werf/werf/commit/bc28f48de36838d7af5e132c0ddd54c8443d9d8a))

### [1.2.141](https://www.github.com/werf/werf/compare/v1.2.140...v1.2.141) (2022-07-27)


### Bug Fixes

* **local-cache-cleanup:** more correct GC for ~/.local_cache/git_* data ([e93bb73](https://www.github.com/werf/werf/commit/e93bb73d046a4642005f0d3d306b837ed32e68fd))

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
