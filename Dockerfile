FROM golang:1.25-alpine AS build
WORKDIR /app

RUN apk add --no-cache bash

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -mod=readonly -o /app/main cmd/api/main.go;

FROM alpine:latest
RUN apk add --no-cache dumb-init

COPY --from=build /app/main .
EXPOSE 8080
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./main"]