FROM ubuntu:20.04 AS builder

ADD src/file* /app/added/
COPY src/file* /app/copied/


FROM ubuntu:20.04 AS result

ARG CHANGED_ARG="should_be_changed"
ENV COMPOSED_ENV="env-${CHANGED_ARG}"
LABEL COMPOSED_LABEL="label-${CHANGED_ARG}"

SHELL ["/bin/sh", "-c"]
USER 0:0
WORKDIR /

COPY --from=builder /app /app

RUN touch /created-by-run-state1

ENTRYPOINT ["sh", "-ec"]
CMD ["tail -f /dev/null"]
ONBUILD RUN echo onbuild
STOPSIGNAL SIGTERM
HEALTHCHECK CMD echo healthcheck
EXPOSE 80/tcp
