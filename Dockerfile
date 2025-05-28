FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /app/gqls ./cmd/cli/main.go

FROM scratch
COPY --from=builder /app/gqls /usr/local/bin/gqls
ENTRYPOINT ["/usr/local/bin/gqls"]
