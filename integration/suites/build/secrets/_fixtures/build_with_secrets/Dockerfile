FROM registry.werf.io/base/ubuntu:22.04

RUN --mount=type=secret,id=ENV_SECRET \
    [ "$(cat /run/secrets/ENV_SECRET)" = "WERF_BUILD_SECRET" ] || (echo "Env does not match the expected value" && exit 1)

RUN --mount=type=secret,id=file \
    grep -q "filecontent" /run/secrets/file || (echo "Src secret does not contain the expected content" && exit 1)

RUN --mount=type=secret,id=plainSecret \
    [ "$(cat /run/secrets/plainSecret)" = "plainSecretValue" ] || (echo "PlainSecret does not match the expected value" && exit 1)

RUN --mount=type=secret,id=secret_file_in_home \
    grep -q "secret" /run/secrets/secret_file_in_home || (echo "Src secret does not contain the expected content" && exit 1)

COPY file /file

RUN touch /created-by-run
