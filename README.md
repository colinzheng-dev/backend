# Monorepo for Veganbase backend services

## Protocol Buffers code generation

All Protocol Buffers definitions live in `services/proto`.

Each service that uses Protocol Buffers has a `proto` directory
containing a `genproto.go` file with something like the following in
it:

```
package blob_service

//go:generate protoc --go_out=plugins=grpc:. -I ../../proto ../../proto/blob_service.proto
```

This generates a `blob_service.pb.go` file in the service `proto`
directory, which can then be used within the service.
