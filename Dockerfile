FROM debian:stable-slim AS builder

RUN mkdir -p /work /data/tmpl /data/static
ADD yugi /usr/bin/yugi

ADD web/*.html /data/tmpl/
ADD static/* /data/static/
ADD yugi.docker.toml /data/yugi.toml

WORKDIR /work
VOLUME /data/tmpl
EXPOSE 8080
ENTRYPOINT ["/usr/bin/yugi"]
CMD ["--config", "/data/yugi.toml", "serve"]

