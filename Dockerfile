FROM golang:alpine AS builder

RUN groupadd --gid 2000 standuplytest \
  && useradd --uid 2000 --gid standuplytest --shell /bin/bash --create-home standuplytest

USER 2000

COPY . $GOPATH/src/github.com/kzharkov/standuplytest
WORKDIR $GOPATH/src/github.com/kzharkov/standuplytest

RUN go get -d -v

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/standuplytest

FROM scratch

COPY --from=builder /go/bin/standuplytest /go/bin/standuplytest

EXPOSE 8443

ENTRYPOINT ["/go/bin/standuplytest"]