# Builder for services.
FROM eu.gcr.io/veganbase/module-builder

# Copy over whole source tree.
COPY . .

# Do all code generation.
RUN go generate ./...
