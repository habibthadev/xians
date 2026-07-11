FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/xians-api ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/xians-reaper ./cmd/reaper

FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /bin/xians-api /usr/local/bin/xians-api
COPY --from=builder /bin/xians-reaper /usr/local/bin/xians-reaper

EXPOSE 8080

CMD ["xians-api"]
