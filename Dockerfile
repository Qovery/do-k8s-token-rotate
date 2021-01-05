FROM golang:1.15.6-buster as build

ENV GOPATH=/go/src/github.com/Qovery/do-k8s-token-rotate
ADD . /mnt
WORKDIR /mnt
RUN go get && go build -o do-k8s-token-rotate main.go

FROM debian:buster-slim

COPY --from=build /mnt/do-k8s-token-rotate /usr/bin/do-k8s-token-rotate
RUN apt-get update && apt-get -y install ca-certificates && chmod 755 /usr/bin/do-k8s-token-rotate && useradd qovery
USER qovery
CMD ["/usr/bin/do-k8s-token-rotate"]
