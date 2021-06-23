# Base builder image.
FROM golang:1.16.4

# Install OS and Go dependencies for build.
RUN apt-get update
RUN apt-get install -y protobuf-compiler libprotobuf-dev
RUN go env
RUN echo $PATH
ADD https://github.com/vektra/mockery/releases/download/v2.7.5/mockery_2.7.5_Linux_x86_64.tar.gz /tmp/
RUN tar -xvf /tmp/mockery_2.7.5_Linux_x86_64.tar.gz && chmod +x mockery && mv mockery /bin/
RUN go get -u github.com/shuLhan/go-bindata/...
RUN go get -u github.com/golang/protobuf/protoc-gen-go
RUN go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Set up working directory.
RUN mkdir -p /work/bin
WORKDIR /work
