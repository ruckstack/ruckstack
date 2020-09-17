FROM alpine:3.12

RUN addgroup -g 1000 -S ruckstack && adduser -S ruckstack -G ruckstack -h /ruckstack --uid 1000

RUN mkdir -p /data/out && \
    mkdir -p /data/project && \
    mkdir -p /data/cache && \
    chown -R ruckstack:ruckstack /data/out

USER ruckstack

COPY . /ruckstack

VOLUME ["/data/out", "/data/project", "/data/cache"]
ENTRYPOINT ["/ruckstack/bin/ruckstack"]
CMD ["--help"]