FROM golang:1.27.3-trixie

WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download

COPY ./app /app

COPY ./bpf /bpf

RUN go mod download

RUN bash gen.sh

RUN cd /app && go build -a -o app .



CMD ["cd /app", "&&", "./app"]
