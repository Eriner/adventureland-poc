# One of these days I'll figure out how to get this working in Alpine...
FROM golang:latest as build
RUN mkdir -p ${GO}/src/private/albot
COPY . ${GO}/src/private/albot
WORKDIR ${GO}/src/private/albot
RUN go build -o /exe main.go
RUN go build -o /deps deps.go

FROM ubuntu:bionic AS base

ARG UID=1000
ARG GID=1000

# Install Go
RUN apt-get update && apt-get install -y ca-certificates

# Install (Chromium) Playwright dependencies
RUN apt-get install -y libwoff1 \
                       libopus0 \
                       libwebp6 \
                       libwebpdemux2 \
                       libenchant1c2a \
                       libgudev-1.0-0 \
                       libsecret-1-0 \
                       libhyphen0 \
                       libgdk-pixbuf2.0-0 \
                       libegl1 \
                       libnotify4 \
                       libxslt1.1 \
                       libevent-2.1-6 \
                       libgles2 \
                       libvpx5 \
                       libnss3 \
                       libxss1 \
                       libasound2 \
                       fonts-noto-color-emoji \
                       libdbus-glib-1-2 \
                       libxt6 \
                       ffmpeg

# if group exists, append user to group +audio,video, otherwise add group
RUN ( grep -q -E "^.*:.*:${GID}:.*$" /etc/group && \
      groupadd -r appuser --gid ${GID} && \
      useradd --uid ${UID} -r -g appuser -G audio,video appuser ) || \
    groupadd -r appuser && \
    useradd --uid ${UID} -r -g appuser -G audio,video appuser

RUN mkdir -p /app/config && chown appuser:appuser /app

# HACK needed for Chrome/Playwright to launch. Error that occurs when this hack
# isn't in place are... unhelpful. DO NOT REMOVE!
RUN mkdir -p "/home/appuser/Downloads" && chown -R appuser:appuser "/home/appuser"
USER appuser
COPY --from=build /exe /app/exe
COPY --from=build /deps /app/deps
#HACK
COPY ./config.yaml /app/config/config.yaml
WORKDIR /app
RUN /app/deps
ENTRYPOINT ["/app/exe"]
