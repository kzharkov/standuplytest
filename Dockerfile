FROM golang:alpine AS builder

RUN addgroup -S --gid 2000 standuplytest \
  && adduser -S --disabled-password --uid 2000 --ingroup standuplytest --shell /bin/bash -h "$(pwd)" standuplytest

COPY . $GOPATH/src/github.com/kzharkov/standuplytest
WORKDIR $GOPATH/src/github.com/kzharkov/standuplytest

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/standuplytest

FROM scratch

COPY --from=builder /go/bin/standuplytest /go/bin/standuplytest

EXPOSE 8443

USER 2000

ENTRYPOINT ["/go/bin/standuplytest"]