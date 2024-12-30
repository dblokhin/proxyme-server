# build
FROM docker.io/library/golang:1.23-alpine AS builder

RUN apk update && apk add --no-cache \
        ca-certificates \
        tzdata

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -a -o proxyme .

# runner
FROM scratch

WORKDIR /app
COPY --from=builder /app/proxyme .
USER 1000:1000

ENTRYPOINT ["./proxyme"]
