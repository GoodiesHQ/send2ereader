FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache the module download layer separately so it only reruns when
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a fully static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o send2ereader .

FROM scratch

COPY --from=builder /app/send2ereader /send2ereader

EXPOSE 8080

ENTRYPOINT ["/send2ereader"]
