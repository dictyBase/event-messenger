FROM golang:1.13.10-buster
ENV GOPROXY https://proxy.golang.org
RUN mkdir -p /event-messenger
WORKDIR /event-messenger
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD cmd cmd
ADD internal internal
RUN go build -o app cmd/event-messenger/main.go

FROM gcr.io/distroless/static
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="event-messenger" \
      org.label-schema.description="Handle events from nats messaging" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="e.g. https://github.com/dictyBase/event-messenger" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"
COPY --from=0 /event-messenger/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]
