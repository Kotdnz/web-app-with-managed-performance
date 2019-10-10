FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
COPY web_service_1.go $GOPATH/src
WORKDIR $GOPATH/src
RUN go get -d -v
RUN go build -o /tmp/webapp web_service_1.go

FROM alpine
RUN addgroup -S appgroup && adduser -S appuser -G appgroup && mkdir -p /app
COPY --from=builder /tmp/webapp /app
RUN chmod a+rx /app/webapp
USER appuser
WORKDIR /app
ENV LISTENING_PORT 8080
CMD ["./webapp"]
