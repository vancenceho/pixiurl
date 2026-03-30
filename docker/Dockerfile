# build 
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 go build -o /server .

# run
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
ENV PORT=8080
EXPOSE 8080
USER nobody

# NOTE: binding port 80 as non-root may fail on some images; either use root in container for port 80
# or listen on 8080 and map ports in compose.
CMD ["/server"]