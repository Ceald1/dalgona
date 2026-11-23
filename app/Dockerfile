FROM golang:1.27.3-trixie

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN go build -a -o app .



CMD ["./app"]
