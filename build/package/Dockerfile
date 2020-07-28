FROM golang:1.13.12-buster
ENV GOPROXY https://proxy.golang.org
RUN mkdir -p /event-messenger
WORKDIR /event-messenger
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD cmd cmd
ADD internal internal
RUN CGO_ENABLED=0 go build -o app cmd/event-messenger/main.go

FROM debian:10.4-slim
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.version=$VERSION \
      org.label-schema.name="event-messenger" \
      org.label-schema.description="Handle events from nats messaging" \
      org.label-schema.vcs-url="https://github.com/dictyBase/event-messenger" \
      org.label-schema.schema-version="1.0"
RUN apt-get update \
    && apt-get -y install curl \
    && curl -o /tmp/wkhtmltopdf.deb -L https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6-1/wkhtmltox_0.12.6-1.buster_amd64.deb \
    && dpkg -i /tmp/wkhtmltopdf.deb || apt-get -y -f install
COPY --from=0 /event-messenger/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]
