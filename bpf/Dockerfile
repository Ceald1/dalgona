# ---- stage 1: bpftool source ----
FROM ghcr.io/libbpf/bpftool:latest AS bpftool-provider

# ---- stage 2: build ----
FROM ubuntu:22.04 AS build
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        clang llvm libelf-dev libbpf-dev build-essential make && \
    rm -rf /var/lib/apt/lists/*

COPY --from=bpftool-provider /usr/bin/bpftool /usr/local/bin/bpftool

WORKDIR /work
COPY . .
RUN make

# ---- stage 3: runtime ----
FROM debian:bookworm-slim
RUN apt-get update && \
    apt-get install -y --no-install-recommends make libelf1 zlib1g && \
    rm -rf /var/lib/apt/lists/*

COPY --from=bpftool-provider /usr/bin/bpftool /usr/local/bin/bpftool

WORKDIR /work
COPY --from=build /work /work

CMD ["/bin/bash", "run.sh"]

#FROM ghcr.io/libbpf/bpftool:latest AS bpftool-provider
#
## Stage 2: Build environment
#FROM ubuntu:22.04
#
#ENV DEBIAN_FRONTEND=noninteractive
#
## Copy bpftool executable from Stage 1
#COPY --from=bpftool-provider /usr/bin/bpftool /usr/local/bin/bpftool
#
#RUN apt-get update && \
#    apt-get install -y --no-install-recommends \
#        clang \
#        llvm \
#        libelf-dev \
#        libbpf-dev \
#        build-essential \
#        make \
#        sudo && \
#    rm -rf /var/lib/apt/lists/*
#
#
#WORKDIR /work
#
#COPY . .
#
#RUN make
#
#CMD ["/bin/bash", "run.sh"]
#
