# Builder for email service.
FROM eu.gcr.io/veganbase/service-builder AS builder
WORKDIR /work/services/email-service
RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o /work/bin/email-service

# For root CA certificates.
FROM alpine:3.9 as ca
RUN apk add -U --no-cache ca-certificates

# Build minimal deployment image.
FROM scratch
COPY --from=ca /etc/passwd /etc/passwd
COPY --from=ca /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /work/bin/email-service .
ENTRYPOINT ["./email-service"]
