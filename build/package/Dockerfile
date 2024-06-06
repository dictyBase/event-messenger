FROM golang:1.22-bullseye
ENV CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64
RUN apt-get -qq update \
	&& apt-get -yqq install upx
RUN mkdir -p /event-messenger
WORKDIR /event-messenger
COPY go.mod ./
COPY go.sum ./
RUN go mod download
ADD cmd cmd
ADD internal internal
RUN go build \
	-a \
	-ldflags "-s -w -extldflags '-static'" \
	-installsuffix cgo \
	-tags netgo \
	-o /bin/app \ 
	cmd/event-messenger/main.go
RUN upx -q -9 /bin/app

FROM debian:11.7-slim
RUN apt-get update \
	&& apt-get -y install curl \
	&& ln -sf /usr/share/zoneinfo/America/Chicago /etc/localtime \
	&& curl -o /tmp/wkhtmltopdf.deb -L \
	https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-2/wkhtmltox_0.12.6.1-2.bullseye_amd64.deb \
	&& dpkg -i /tmp/wkhtmltopdf.deb || apt-get -y -f install
COPY --from=0 /bin/app /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/app"]
