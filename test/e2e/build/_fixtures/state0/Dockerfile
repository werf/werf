FROM ubuntu:20.04 AS builder

ADD src/file* /app/added/
ADD "https://github.com/octocat/Hello-World/tarball/7fd1a60b01f91b314f59955a4e4d4e80d8edf11d" /helloworld.tgz
COPY src/file* /app/copied/


FROM ubuntu:20.04 AS result

ARG CHANGED_ARG="should_be_changed"
ENV COMPOSED_ENV="env-${CHANGED_ARG}"
LABEL COMPOSED_LABEL="label-${CHANGED_ARG}"

SHELL ["/bin/sh", "-c"]
USER 0:0
WORKDIR /

COPY --from=builder /app /app
COPY --from=builder /helloworld.tgz /helloworld.tgz

RUN touch /created-by-run
RUN mkdir -p /persistent/should-exist-in-volume

ENTRYPOINT ["sh", "-ec"]
CMD ["tail -f /dev/null"]
VOLUME /persistent
ONBUILD RUN echo onbuild
STOPSIGNAL SIGTERM
HEALTHCHECK CMD echo healthcheck
EXPOSE 80/tcp
