FROM golang:1.22-alpine AS builder
COPY . /go/src/app
WORKDIR /go/src/app
RUN go build ./cmd/multiplexer

FROM scratch
COPY --from=builder /go/src/app/multiplexer /
COPY ./.env /
COPY ./docs/ /docs/
ENTRYPOINT ["./multiplexer"]