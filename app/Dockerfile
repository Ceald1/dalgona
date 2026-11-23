FROM golang:1.27.3-trixie

WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download

COPY . .

RUN go mod download

RUN bash gen.sh

RUN go build -a -o app .



CMD ["./app"]
