FROM ubuntu:22.04

RUN --mount=type=secret,id=ENV_SECRET \
    export MY_ENV_SECRET=$(cat /run/secrets/ENV_SECRET)

RUN --mount=type=secret,id=file \
    export MY_SRC_SECRET=$(cat /run/secrets/file)

RUN --mount=type=secret,id=plainSecret \
    export MY_PLAIN_SECRET=$(cat /run/secrets/plainSecret)

COPY file /file

RUN touch /created-by-run
