FROM scratch

MAINTAINER name

LABEL maintainer="i@example.com"
LABEL save=me
LABEL pleaseremove=me
LABEL remove.completely=yes
LABEL remove.all=yes
LABEL remove_global_label=yes

ENV PATH=/usr/bin \
    LANG=en_US.utf8 \
    APP_ENV=test \
    APP_VERSION=0.0.1 \
    HOME=/home/app \
    REMOVE=ME

WORKDIR /home/app/scripts

EXPOSE 1234/tcp

VOLUME ["/home/app/data"]
VOLUME ["/home/remove/me"]

USER nobody

STOPSIGNAL SIGTERM

ENTRYPOINT ["/bin/true"]
CMD ["--noop"]
