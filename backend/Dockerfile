FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd/worker

FROM alpine:3.20 AS api
WORKDIR /app
COPY --from=build /bin/api /app/api
EXPOSE 8080
CMD ["/app/api"]

FROM alpine:3.20 AS worker
WORKDIR /app
COPY --from=build /bin/worker /app/worker
CMD ["/app/worker"]
