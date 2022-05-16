# Compiler image
FROM didstopia/base:go-alpine-3.14 AS go-builder

# Copy the project 
COPY . /tmp/rustbot/
WORKDIR /tmp/rustbot/

# Install dependencies
RUN apk add --no-cache protobuf && \
    make deps

# Build the binary
#RUN make build && ls /tmp/rustbot
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/rustbot



# Runtime image
FROM scratch

# Copy certificates
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the built binary
COPY --from=go-builder /go/bin/rustbot /go/bin/rustbot

# Expose environment variables
ENV DISCORD_BOT_TOKEN                ""
ENV DISCORD_CHAT_CHANNEL_ID          ""
ENV DISCORD_OWNER_ID                 ""
ENV WEBRCON_HOST                     "localhost"
ENV WEBRCON_PORT                     "28016"
ENV WEBRCON_PASSWORD                 ""
ENV DISCORD_KILLFEED_CHANNEL_ID      ""
ENV DISCORD_KILLFEED_PVP_ENABLED     "true"
ENV DISCORD_KILLFEED_OTHER_ENABLED   "false"
ENV DISCORD_LOG_CHANNEL_ID           ""
ENV DISCORD_NOTIFICATIONS_CHANNEL_ID ""
ENV DISCORD_PLAYERLIST_CHANNEL_ID    ""

# Expose volumes
VOLUME [ "/.db" ]

# Run the binary
ENTRYPOINT ["/go/bin/rustbot"]
