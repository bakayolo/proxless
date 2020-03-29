FROM golang:1.13-alpine as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o proxless cmd/main.go

FROM alpine:3.9.3
WORKDIR /app
COPY --from=builder /app/proxless proxless
ENTRYPOINT ["/app/proxless"]
