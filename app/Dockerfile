# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.27.3-trixie AS build

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        clang llvm libelf-dev libbpf-dev make && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# deps first, so this layer is cached until go.mod/go.sum change
COPY app/go.mod app/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY app/ /app/
COPY bpf/ /bpf/

# generate eBPF objects + Go bindings
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    bash gen.sh

# static, stripped binary (no -a: it forces a full rebuild every time)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/app .

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/app /app
CMD ["/app"]


#FROM golang:1.27.3-trixie
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
#WORKDIR /app
#
## COPY go.mod go.sum ./
## RUN go mod download
#
#COPY ./app /app
#
#COPY ./bpf /bpf
#
#RUN go mod download
#
#RUN bash gen.sh
#
#RUN go build -a -o app .
#
#
#
#CMD ["./app"]
