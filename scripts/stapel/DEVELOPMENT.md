Stapel is a self-contained toolbox image mounted by werf into build containers.

Current stapel image is built from Alpine userspace and includes the binaries used by werf (`bash`, `git`, `rsync`, `tar`, `install`, `find`, `base64`, etc.) together with required shared libraries.

Main runtime paths inside the image:

* `/.werf/stapel/embedded/bin`
* `/.werf/stapel/embedded/lib`
* `/.werf/stapel/embedded/libexec`
* `/.werf/stapel/embedded/share`
* `/.werf/stapel/lib`
* `/.werf/stapel/bin`
* `/.werf/stapel/sbin`

werf mounts stapel into each stapel-build container to provide deterministic service tooling.

## Local development flow

1. Update build logic in `stapel/Dockerfile`.
2. Build a development stapel image:

   ```shell
   scripts/stapel/build.sh
   ```

   By default the script builds `registry-write.werf.io/werf/stapel:dev`.

3. (Optional) Build for a specific target platform:

   ```shell
   STAPEL_PLATFORM=linux/arm64 scripts/stapel/build.sh
   STAPEL_PLATFORM=linux/amd64 scripts/stapel/build.sh
   ```

4. Test werf with the development stapel image:

   ```shell
   export WERF_STAPEL_IMAGE_NAME=registry.werf.io/werf/stapel
   export WERF_STAPEL_IMAGE_VERSION=dev
   werf build ...
   ```

## Multi-platform publish (single tag)

Use Docker Buildx to publish `linux/amd64` and `linux/arm64` under one tag:

```shell
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --file stapel/Dockerfile \
  --target final \
  --tag registry-write.werf.io/werf/stapel:dev \
  --push \
  .
```

Verify manifest list:

```shell
docker buildx imagetools inspect registry.werf.io/werf/stapel:dev
```

## Release notes for stapel version bumps

After publishing a new stapel tag, update `VERSION` in `pkg/stapel/stapel.go` and rebuild werf binaries.
