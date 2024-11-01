FROM ${source_image} AS source_image

FROM ${base_image}
COPY --from=source_image /usr/local/bin/werf /usr/local/bin/werf
