FROM --platform=linux/amd64 golang:1.26-alpine AS builder

# Install CA bundle in builder so we can copy it into scratch
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM scratch

# Copy the CA cert bundle into scratch (this is the key part)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/main /main

EXPOSE 8080
ENTRYPOINT ["/main"]