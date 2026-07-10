# syntax=docker/dockerfile:1

# --- Build stage ---
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Build the statically linked server binary. Templates and static assets are
# embedded into the binary via //go:embed, so the final image needs no extra files.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# --- Runtime stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/server /app/server

EXPOSE 8080
ENTRYPOINT ["/app/server"]
