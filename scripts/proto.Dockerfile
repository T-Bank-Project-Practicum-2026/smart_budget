FROM bufbuild/buf:1.50.0 AS buf
FROM golang:1.27-alpine
COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 \
    && go install connectrpc.com/connect/cmd/protoc-gen-connect-go@v1.18.1
WORKDIR /workspace
