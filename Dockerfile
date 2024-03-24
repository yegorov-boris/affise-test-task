FROM golang:1.22-alpine AS builder
COPY . /go/src/app
WORKDIR /go/src/app
RUN go build ./cmd/multiplexer

FROM scratch
COPY --from=builder /go/src/app/multiplexer /
COPY .env /
VOLUME ./store /store
EXPOSE 8080
ENTRYPOINT ["./multiplexer"]