FROM alpine:latest

ENV PORT=33333
ENV APP_USER=gomud
ENV APP_UID=10001

# This helper image is only used to open a telnet client against the game server.
RUN apk add --no-cache busybox-extras \
    && adduser -D -u "${APP_UID}" "${APP_USER}"

# Match the main container's non-root runtime model.
USER ${APP_USER}

ENTRYPOINT telnet go-mud-server ${PORT}
