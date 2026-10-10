FROM golang:1.27-alpine
RUN go install github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc6
WORKDIR /workspace/backend
