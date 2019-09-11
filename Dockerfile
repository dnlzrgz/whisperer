# Build stage
FROM golang:1.13-alpine3.10 AS build
RUN apk add --no-cache git
WORKDIR /app/
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o whisperer main.go

# Final stage
FROM alpine:3.10.2
LABEL maintainer="danielkvist@protonmail.com"
RUN apk add --no-cache ca-certificates
COPY --from=build /app/whisperer /app/whisperer
COPY --from=build /app/urls.txt /app/urls.txt
WORKDIR /app/
ENTRYPOINT ["./whisperer"]
CMD ["-v", "-d", "3s"]
