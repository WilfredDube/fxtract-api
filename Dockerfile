#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o fxtractapi -v .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/fxtractapi .
ENTRYPOINT [ "./fxtractapi" ]
