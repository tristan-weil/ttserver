FROM alpine:3.12

LABEL org.opencontainers.image.source https://github.com/tristan-weil/ttserver

RUN set -x \
    # add user
    && addgroup -g 101 -S ttserver \
    && adduser -S -D -H -u 101 -h /ttserver/public -s /sbin/nologin -G ttserver -g ttserver ttserver \
    # ttserver
    && apk add --no-cache tini \
    # prepare dirs
    && mkdir -p /ttserver/public /ttserver/etc

COPY build/ttserver                 /usr/local/bin/ttserver
COPY examples/gopher-docker.config  /ttserver/etc/ttserver.config
COPY examples/gopher/               /ttserver/public

RUN set -x \
    && chmod +x /usr/local/bin/ttserver

USER ttserver
WORKDIR /ttserver/public

ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/usr/local/bin/ttserver", "-config", "/ttserver/etc/ttserver.config"]
