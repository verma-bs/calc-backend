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

## API examples

Calculate a mixed-precedence expression:

```sh
curl -i http://localhost:8080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"1+2*3-2^2"}'
```

Expected response:

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"result":3}
```

Test decimal arithmetic:

```sh
curl -i http://localhost:8080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"27.02+20"}'
```

Test invalid syntax:

```sh
curl -i http://localhost:8080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"1+4+3-"}'
```

Expected response: `400 Bad Request` with `{"error":"invalid expression"}`.

Test division by zero:

```sh
curl -i http://localhost:8080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -d '{"expression":"1/0"}'
```

Expected response: `422 Unprocessable Entity` with `{"error":"division by zero"}`.

## Backend design decisions and assumptions

- The API exposes one infix-expression endpoint instead of separate endpoints for each arithmetic operation.
- Go `float64` is used for operator calculations and JSON results. The supported absolute numeric range is limited to `9,007,199,254,740,991` so integer results remain exact relative to JavaScript `Number` behavior.
- Final expressions are normalized to reduce ordinary decimal floating-point display noise, while meaningful large-number increments are preserved.
- Operator precedence is `^` first, followed by `*` and `/`, then `+` and `-`. Power is right-associative.
- The parser supports parentheses, unary signs, decimal literals, and scientific notation. Consecutive `+`/`-` signs are rejected.
- Invalid syntax is a `400` response; arithmetic constraints such as division by zero or out-of-range numbers are `422` responses.
- JSON has a 4 KiB body limit and rejects unknown fields and trailing JSON values.
- The HTTP server is a standard-library `net/http` implementation to keep dependencies and deployment small.

## Test Coverage
 coverage.out under root directory

