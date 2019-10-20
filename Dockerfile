#
# revision 2 from 20-Oct-2019
#
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
ENV RATE 200
ENV LATENCY 100
ENV ERRORRATE 10
ENV SATURATION 500
CMD ["./webapp"]
