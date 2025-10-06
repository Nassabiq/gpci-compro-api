# ====== Builder ======
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# build API
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api
# build Worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/worker ./cmd/worker


# ====== Runtime ======
FROM gcr.io/distroless/static-debian12 AS runtime
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /out/worker /app/worker
# embed needs migration files present in the binary; already embedded.
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/api"]