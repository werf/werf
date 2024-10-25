ARG distro_image
ARG source_image

FROM ${source_image} AS source_image 

FROM ${distro_image}
ARG source
ARG destination
COPY --from=source_image ${source} ${destination}