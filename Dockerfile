# BUILD STAGE
FROM golang:1.23-alpine as builder
RUN apk add --no-cache git
WORKDIR /repo
COPY . .
ENV CGO_ENABLED=0
RUN scripts/install.sh
RUN cherry build -cross-compile=false

# FINAL STAGE
FROM golang:1.23-alpine
RUN apk add --no-cache ca-certificates git
RUN apk add --no-cache ruby ruby-json && \
    gem install rdoc --no-document && \
    gem install github_changelog_generator
COPY --from=builder /repo/bin/cherry /usr/local/bin/
RUN chown -R nobody:nogroup /usr/local/bin/cherry
USER nobody
ENTRYPOINT [ "cherry" ]
