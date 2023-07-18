# Build stage
FROM golang:1.19-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .

COPY app.env .
COPY ./scripts/start.sh .
COPY ./scripts/wait-for.sh .
COPY db/migration ./db/migration

# Expose port 8080 to the outside world
EXPOSE 8080

RUN chmod +x /app/start.sh
RUN chmod +x /app/wait-for.sh

# Run the binary program
CMD ["/app/wait-for.sh", "database:5432", "--", "/app/start.sh"]


