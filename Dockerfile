FROM golang:1.24.1-alpine AS  builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o simple-auth ./cmd/simple-auth
RUN CGO_ENABLED=0 GOOS=linux go build -o migrator ./cmd/migrator

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/simple-auth .
COPY --from=builder /app/migrator .
COPY config/prod.yml ./config/
RUN mkdir -p /app/migrations
COPY migrations ./migrations/ 
EXPOSE 8084
CMD ["./simple-auth"]
