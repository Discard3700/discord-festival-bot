FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bot ./cmd/bot && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /import ./cmd/import


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bot /bot
COPY --from=builder /import /import
COPY --from=builder /src/migrations /migrations

USER 65534

ENTRYPOINT ["/bot"]