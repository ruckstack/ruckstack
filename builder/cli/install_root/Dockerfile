FROM alpine:3.12

RUN addgroup -g 9999 -S ruckstack && adduser -S ruckstack -G ruckstack -h /ruckstack --uid 9999

RUN mkdir -p /data/out && \
    mkdir -p /data/project && \
    mkdir -p /data/cache && \
    mkdir -p /data/tmp && \
    chmod 777 /data/out && \
    chmod 777 /data/project && \
    chmod 777 /data/cache && \
    chmod 777 /data/tmp

USER ruckstack

COPY . /ruckstack

VOLUME ["/data/out", "/data/project", "/data/cache", "/data/tmp"]
ENTRYPOINT ["/ruckstack/bin/ruckstack"]
CMD ["--help"]