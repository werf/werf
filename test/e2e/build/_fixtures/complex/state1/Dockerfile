FROM ubuntu:22.04 AS builder

ADD src/file* /app/added/
COPY src/file* /app/copied/


FROM ubuntu:22.04 AS result

ARG CHANGED_ARG="should_be_changed"
ENV COMPOSED_ENV="env-${CHANGED_ARG}"
LABEL COMPOSED_LABEL="label-${CHANGED_ARG}"
MAINTAINER "maintainer-${CHANGED_ARG}"

SHELL ["/bin/sh", "-c"]
USER 0:0
WORKDIR /

COPY --from=builder /app /app

RUN touch /created-by-run-state1
RUN --mount=type=tmpfs,target=/tmpfs touch /tmpfs/created-by-run-state1
RUN --mount=type=bind,target=/bind,rw touch /bind/created-by-run-state1
RUN --mount=type=bind,from=builder,target=/bind-from-builder,rw touch /bind-from-builder/created-by-run-state1

ENTRYPOINT ["sh", "-ec"]
CMD ["tail -f /dev/null"]
ONBUILD RUN echo onbuild
STOPSIGNAL SIGTERM
HEALTHCHECK CMD echo healthcheck10
EXPOSE 8000/tcp
