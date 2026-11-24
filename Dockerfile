FROM golang:1.27.1-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/calculator-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/calculator-api /usr/local/bin/calculator-api
USER app
ENV HTTP_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/calculator-api"]
