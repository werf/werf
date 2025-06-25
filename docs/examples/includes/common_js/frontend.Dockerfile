FROM nginx:alpine
ARG BACKEND_URL
ENV BACKEND_URL=${BACKEND_URL}
WORKDIR /usr/share/nginx/html
COPY frontend/index.html .
RUN envsubst < index.html > index.tmp.html && \
    mv index.tmp.html index.html
