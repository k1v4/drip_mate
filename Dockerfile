FROM golang:1.26.1-alpine AS builder

WORKDIR /usr/local/src

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . .
RUN go build -o ./bin/app cmd/main/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app /
COPY --from=builder /usr/local/src/migrations /migrations

EXPOSE 8080 50051

COPY .env /

CMD ["/app", "-config=/config/config.yaml"]

