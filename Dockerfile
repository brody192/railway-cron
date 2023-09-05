FROM golang:1.21.0-alpine3.18 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN go build -ldflags "-w -s -extldflags '-static'" -trimpath -a -o main

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /app/main ./

ENTRYPOINT ["/app/main"]