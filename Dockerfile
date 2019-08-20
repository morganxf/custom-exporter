############################
# STEP 1 build executable binary
############################
FROM golang:1.11 AS builder

RUN apt-get update \
 && apt-get install -y vim-tiny

ARG VERSION
WORKDIR /go/src/github.com/morganxf/custom_exporter
ADD . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-X main.version=${VERSION}" -o custom_exporter .

############################
# STEP 2 build a small image
############################
FROM debian:stretch

RUN apt-get update \
 && apt-get install -y procps curl telnet vim-tiny openssh-server openssh-client --no-install-recommends

ENV HOME=/etc/custom_exporter

WORKDIR ${HOME}
# Copy our static executable and configuration.
COPY --from=builder /go/src/github.com/morganxf/custom_exporter/custom_exporter ./


CMD [ "/etc/custom_exporter/custom_exporter" ]
