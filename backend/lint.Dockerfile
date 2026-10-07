FROM golangci/golangci-lint:v2.14-alpine AS linter
FROM golang:1.27-alpine
COPY --from=linter /usr/bin/golangci-lint /usr/local/bin/golangci-lint
ENV GOTOOLCHAIN=local
WORKDIR /app
ENTRYPOINT ["golangci-lint"]
