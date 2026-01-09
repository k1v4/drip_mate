FROM golang:1.25-alpine AS builder

WORKDIR /usr/local/src/file_upload_service

COPY file_upload_service/go.mod file_upload_service/go.sum ./
RUN go mod download

COPY file_upload_service/ ./

RUN go build -o ./bin/app cmd/main/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/file_upload_service/bin/app /app

EXPOSE 50053

COPY .env /

CMD ["/app"]
