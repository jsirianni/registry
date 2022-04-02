# Intended to be run by Goreleaser
FROM alpine:3
COPY registry .
ENTRYPOINT [ "/registry" ]
