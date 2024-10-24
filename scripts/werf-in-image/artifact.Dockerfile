ARG distro
ARG source_image

FROM ${source_image} as source_image 

FROM ${distro}
ARG source
ARG dest
COPY --from=source_image ${source} ${dest}