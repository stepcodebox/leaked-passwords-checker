FROM golang:1.23.3 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /app/bin

RUN CGO_ENABLED=1 go build -o /app/bin/leaked-passwords-checker .

FROM debian:bookworm

RUN apt-get update && apt-get install -y sqlite3 nano

WORKDIR /app

RUN mkdir -p /app/bin /app/logs /app/configs /app/database

COPY --from=builder /app/bin/leaked-passwords-checker /app/bin/leaked-passwords-checker

COPY infra/docker/top100K.db /app/database/top100K.db

COPY infra/docker/leaked-passwords-checker.json /app/configs/leaked-passwords-checker.json

RUN chmod 644 /app/configs/leaked-passwords-checker.json

RUN touch /app/logs/leaked-passwords-checker.log && chmod 644 /app/logs/leaked-passwords-checker.log

COPY infra/docker/docker-entrypoint.sh docker-entrypoint.sh
RUN chmod +x docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./docker-entrypoint.sh"]
