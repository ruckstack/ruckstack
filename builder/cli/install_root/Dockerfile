FROM alpine:3.12

RUN addgroup -g 1000 -S ruckstack && adduser -S ruckstack -G ruckstack -h /ruckstack --uid 1000

RUN mkdir -p /data/out && \
    mkdir -p /data/project && \
    mkdir -p /data/cache && \
    mkdir -p /data/tmp && \
    chown -R ruckstack:ruckstack /data/out && \
    chown -R ruckstack:ruckstack /data/project && \
    chown -R ruckstack:ruckstack /data/cache && \
    chown -R ruckstack:ruckstack /data/tmp

USER ruckstack

COPY . /ruckstack

VOLUME ["/data/out", "/data/project", "/data/cache", "/data/tmp"]
ENTRYPOINT ["/ruckstack/bin/ruckstack"]
CMD ["--help"]