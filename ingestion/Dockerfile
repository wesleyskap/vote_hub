FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY orkai-runiq ./orkai-runiq
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build API and Worker
RUN go build -o /app/api ./cmd/api/main.go
RUN go build -o /app/worker ./cmd/worker/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker

# Default command
CMD ["/app/api"]
