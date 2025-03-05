FROM golang:1.23-bookworm@sha256:3149bc5043fa58cf127fd8db1fdd4e533b6aed5a40d663d4f4ae43d20386665f
ENV DEBIAN_FRONTEND=noninteractive

ARG TASK_VERSION="latest"
ARG GOLANGCI_LINT_VERSION="v1.57.1"
ARG GCI_VERSION="v0.13.5"
ARG GOFUMPT_VERSION="v0.7.0"
ARG MOCGEN_VERSION="v0.5.0"
ARG ENUMER_VERSION="v1.5.11"
ARG GINKGO_VERSION="v2.20.1"


RUN apt-get update && apt-get install -y \
    apt-utils \
    gcc-aarch64-linux-gnu \
    file \
    nodejs \
    npm \
    curl \
    git \
    libbtrfs-dev && \
    rm -rf /var/cache/apt/* /var/lib/apt/lists/* /var/log/*

RUN npm install -g prettier

RUN go install github.com/go-task/task/v3/cmd/task@${TASK_VERSION} && \
    go install github.com/daixiang0/gci@${GCI_VERSION} && \
    go install mvdan.cc/gofumpt@${GOFUMPT_VERSION} && \
    go install go.uber.org/mock/mockgen@${MOCGEN_VERSION} && \
    go install github.com/dmarkham/enumer@${ENUMER_VERSION} && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} && \
    go install github.com/onsi/ginkgo/v2/ginkgo@${GINKGO_VERSION}

RUN useradd -m dev

WORKDIR /home/dev/werf
COPY go.mod go.sum ./
RUN go mod download
USER dev

ENTRYPOINT ["task"]