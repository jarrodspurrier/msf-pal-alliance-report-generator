FROM golang:latest as builder

RUN apt-get update && apt-get install -y cron

ADD . /build/

WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

FROM golang:1.14.9-alpine3.12

RUN mkdir /app

COPY --from=builder /build/ /app/
COPY --from=builder /build/msf-pal-alliance-report-generator-cron /etc/crontabs/root

WORKDIR /app

# Start crond in the foreground with log level 8, output to stderr.
CMD ["crond", "-f", "-d", "8"]
