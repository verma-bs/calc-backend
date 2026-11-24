# Calculator 

A small Go HTTP service for basic arithmetic. The service exposes one JSON endpoint and keeps the calculation logic independent of the HTTP layer.

## Requirements

- Go 1.27.1 or later

## Run locally

From this directory:

```sh
go run ./cmd/api
```

The server listens on `:8080` by default. Set `HTTP_ADDR` to use another address, for example:

```sh
HTTP_ADDR=127.0.0.1:9090 go run ./cmd/api
```

Stop the server with Ctrl+C. It handles interrupt and termination signals and shuts down gracefully.

## API

### `POST /api/v1/calculate`

Send a JSON object containing one infix expression string. The request body is limited to 4 KiB and must contain exactly one JSON value with no unknown fields.

```http
POST /api/v1/calculate HTTP/1.1
Content-Type: application/json

{"expression":"7/2"}
```

For expression previews, send the complete infix expression as a string:

```json
{"expression":"1+2*3"}
```

The service supports `+`, `-`, `*`, `/`, `^`, parentheses, decimal numbers, and scientific notation. It applies standard precedence, with exponentiation right-associative. This example returns `7`. This endpoint accepts only the infix expression request format.

Operands and intermediate results are limited to an absolute value of `9,007,199,254,740,991` (the largest integer safely represented by the frontend's JavaScript `Number` and the backend's `float64`). Larger operands or results are rejected instead of silently losing integer precision. Decimal fractions still use floating-point arithmetic and may be approximate.

Successful responses have status `200 OK`:

```json
{"result":3.5}
```

Division by zero is rejected. Non-finite operands and results are also rejected.

### Errors

Error responses use this JSON shape:

```json
{"error":"division by zero"}
```

| Status | Meaning |
| --- | --- |
| `400 Bad Request` | Invalid JSON, missing or invalid fields, unknown fields, or extra JSON values |
| `405 Method Not Allowed` | The request method is not `POST`; the response includes `Allow: POST` |
| `413 Request Entity Too Large` | The request body exceeds 4 KiB |
| `415 Unsupported Media Type` | The request content type is not `application/json` |
| `422 Unprocessable Entity` | Division by zero, non-finite arithmetic, or a number/result outside the supported numeric range |

All responses use `Content-Type: application/json`.

## Development

Run the test suite and static checks from this directory:

```sh
go test ./...
go vet ./...
```

Build the API executable with:

```sh
go build ./cmd/api
```

## Project layout

```text
cmd/api/                 HTTP server entry point
internal/calculator/     Arithmetic operations and unit tests
internal/httpapi/        HTTP handler and request validation tests
```
