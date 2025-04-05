FROM debian:stable-slim AS builder

RUN mkdir -p /work /data/tmpl /data/static
ADD yugi /usr/bin/yugi

ADD web/*.html /data/tmpl/
ADD static/* /data/static/
ADD yugi.docker.toml /data/yugi.toml

FROM ronmi/mingo

COPY --from=builder /usr/bin/yugi /usr/bin/yugi
COPY --from=builder /data /data

WORKDIR /work
VOLUME /data/tmpl
VOLUME /data/static
EXPOSE 8080
ENTRYPOINT ["/usr/bin/yugi"]
CMD ["--config", "/data/yugi.toml", "serve"]

