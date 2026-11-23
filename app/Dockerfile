FROM golang:1.27.3-trixie

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        clang \
        llvm \
        libelf-dev \
        libbpf-dev \
        build-essential \
        make \
        sudo && \
    rm -rf /var/lib/apt/lists/*


WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download

COPY ./app /app

COPY ./bpf /bpf

RUN go mod download

RUN bash gen.sh

RUN cd /app && go build -a -o app .



CMD ["cd /app", "&&", "./app"]
