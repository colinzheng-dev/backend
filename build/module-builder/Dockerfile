# Build image with populated module cache.
FROM eu.gcr.io/veganbase/base-builder

# Populate module cache.
COPY go.mod .
COPY go.sum .
RUN go mod download
