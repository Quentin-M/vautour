FROM golang:alpine

RUN apk add --no-cache ca-certificates openssl file dumb-init pkgconfig libcrypto1.1 git
RUN apk add --no-cache -t .build-deps gcc libc-dev jansson bison openssl-dev make automake autoconf libtool \
    && git clone --recursive --branch v3.8.0 https://github.com/VirusTotal/yara.git /tmp/yara \
    && cd /tmp/yara \
    && ./bootstrap.sh \
    && ./configure --with-crypto \
    && make \
    && make install

ADD .   /go/src/github.com/quentin-m/vautour/
WORKDIR /go/src/github.com/quentin-m/vautour/
RUN go build -v -ldflags "-X github.com/coreos/clair/pkg/version.Version=$(git describe --tag --always --dirty)" github.com/quentin-m/vautour/cmd/vautour \
    && mv vautour config/ / \
    && apk del --purge .build-deps \
    && rm -rf /tmp/*

WORKDIR /
VOLUME /config
USER nobody
ENTRYPOINT ["/usr/bin/dumb-init", "--", "/vautour"]