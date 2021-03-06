FROM alpine:3.12

RUN addgroup -g 9999 -S ruckstack && adduser -S ruckstack -G ruckstack -h /ruckstack --uid 9999

RUN mkdir -p /workspace && \
    mkdir -p /data && \
    chmod 777 /workspace && \
    chmod 777 /data

USER ruckstack

COPY LICENSE /ruckstack
COPY out/artifacts/linux/ruckstack /ruckstack/bin/ruckstack

ENV RUCKSTACK_WORK_DIR /data

VOLUME ["/workspace", "/data"]
ENTRYPOINT ["/ruckstack/bin/ruckstack"]
CMD ["--help"]