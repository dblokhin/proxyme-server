# build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o proxyme .

# runner
FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
USER nobody

WORKDIR /
COPY --from=builder /app/proxyme .

ENTRYPOINT ["./proxyme"]
