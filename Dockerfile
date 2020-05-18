FROM golang:1.13-alpine AS builder

WORKDIR /build
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./bin/qr-generator


FROM alpine:3.9

WORKDIR app

COPY --from=builder /build/bin/qr-generator ./

EXPOSE 8080

CMD ["./qr-generator"]
