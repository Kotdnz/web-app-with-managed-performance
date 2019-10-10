FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
COPY . $GOPATH/src
WORKDIR $GOPATH/src
RUN go get -d -v
RUN go build -o /tmp/webapp *.go

FROM alpine
RUN addgroup -S appgroup && adduser -S appuser -G appgroup && mkdir -p /app
COPY --from=builder /tmp/webappp /app
RUN chmod a+rx /app/webapp
USER appuser
WORKDIR /app
ENV LISTENING_PORT 8080
CMD ["./webapp"]
