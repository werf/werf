FROM alpine
ARG group
ARG channel
ARG required_version
ARG TARGETARCH

RUN apk add curl gnupg

RUN curl -sSLO "https://tuf.trdl.dev/targets/releases/0.7.0/linux-${TARGETARCH}/bin/trdl" -O "https://tuf.trdl.dev/targets/signatures/0.7.0/linux-${TARGETARCH}/bin/trdl.sig" && \
    curl -sSL https://trdl.dev/trdl.asc | gpg --import && \
    gpg --verify trdl.sig trdl && \
    rm trdl.sig && \
    chmod +x ./trdl && \
    mv trdl /usr/local/bin/trdl && \
    trdl add werf https://tuf.werf.io 1 b7ff6bcbe598e072a86d595a3621924c8612c7e6dc6a82e919abe89707d7e3f468e616b5635630680dd1e98fc362ae5051728406700e6274c5ed1ad92bea52a2 && \
    wget https://github.com/mikefarah/yq/releases/download/v4.24.2/yq_linux_${TARGETARCH} -O /usr/local/bin/yq && \
    chmod +x /usr/local/bin/yq

RUN while true ; do \
        echo "Perform trdl update for werf $group $channel ..." &&\
        trdl update werf ${group} ${channel} &&\
        . $(trdl use werf ${group} ${channel}) &&\
        DOWNLOADED_VERSION=$(werf version | sed -e 's|^v||') &&\
        echo "werf ${group} ${channel}: required version ${required_version}, downloaded version $DOWNLOADED_VERSION" &&\
        [[ "${required_version}" != "$DOWNLOADED_VERSION" ]] || break &&\
        echo "Version mismatch! Will retry update" &&\
        sleep 1; \
done; \
cp $(trdl bin-path werf ${group} ${channel})/werf /usr/local/bin/werf-${group}-${channel}; \
