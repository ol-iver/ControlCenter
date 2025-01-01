FROM golang:1.21-alpine3.18 as builder

RUN apk update && apk add git make gcc curl libc-dev python3 graphviz ragel npm rsync
RUN npm install -g @vue/cli@v4.5.13
RUN npm install -g @vue/cli-service-global@4.5.13
RUN vue --version

# fix building front-end as we are using very old (two years?) dependencies
# See https://github.com/webpack/webpack/issues/14532
ENV NODE_OPTIONS --openssl-legacy-provider

RUN apk update && apk add ca-certificates

ADD . /src

WORKDIR /src

RUN make static_release

FROM scratch

ARG LIGHTMETER_VERSION
ARG LIGHTMETER_COMMIT
ARG IMAGE_TAG

# List of interesting labels: http://label-schema.org/rc1/
LABEL org.label-schema.name="Lightmeter Control Center"
LABEL org.label-schema.vcs-url="https://gitlab.com/lightmeter/controlcenter"
LABEL org.label-schema.url="https://lightmeter.io"
LABEL org.label-schema.description="Mail server delivery monitoring"
LABEL org.label-schema.usage="https://gitlab.com/lightmeter/controlcenter/-/raw/release/$IMAGE_TAG/README.md"
LABEL org.label-schema.vendor="Lightmeter Project"
LABEL org.label-schema.version="$LIGHTMETER_VERSION"
LABEL org.label-schema.vcs-ref="$LIGHTMETER_COMMIT"
LABEL org.label-schema.schema-version="1.0"
LABEL maintainer="Leandro Santiago <leandro@lightmeter.io>"

COPY --from=builder /src/lightmeter /lightmeter
COPY --from=builder /usr/share/ca-certificates /usr/share/ca-certificates

ENV SSL_CERT_DIR /usr/share/ca-certificates/mozilla

VOLUME /logs
VOLUME /workspace
EXPOSE 8080

ENTRYPOINT ["/lightmeter"]
