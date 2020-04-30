FROM golang:1.13.10-buster
LABEL maintainer="Siddhartha Basu <siddhartha-basu@northwestern.edu>"
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
COPY --from=0 /event-messenger/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]
