FROM golang:1.21.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o manager .

# Stage 2: Create a minimal image for running the application
FROM alpine:latest AS runner

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/manager .

CMD ["./manager"]