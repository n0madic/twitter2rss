FROM golang:alpine AS builder

RUN apk add --quiet --no-cache build-base git

WORKDIR /src

ENV GO111MODULE=on

ADD go.* ./

RUN go mod download

ADD . .

RUN cd cmd/twitter2rss && \
    go install -ldflags="-linkmode external -extldflags '-static' -s -w"


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/* /

CMD ["/twitter2rss"]
