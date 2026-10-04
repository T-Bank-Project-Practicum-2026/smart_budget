FROM golang:1.27.1-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
ARG APP_PATH=./cmd
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/server "$APP_PATH"

FROM alpine:3.22 AS runtime
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/server ./server
USER app
EXPOSE 8080
ENV HTTP_ADDR=:8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/api/health || exit 1
ENTRYPOINT ["/app/server"]
