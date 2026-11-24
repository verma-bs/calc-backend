package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"calculator/internal/calculator"
)

const maxBodyBytes = 4 << 10 // 4 KiB is ample for a calculator expression.

type calculateRequest struct {
	Expression string `json:"expression"`
}

type calculateResponse struct {
	Result float64 `json:"result"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", calculate)
	return mux
}

func calculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "use POST")
		return
	}

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		var sizeErr *http.MaxBytesError
		if errors.As(err, &sizeErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
		} else {
			writeError(w, http.StatusBadRequest, "could not read request body")
		}
		return
	}

	var input calculateRequest
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid calculator request")
		return
	}
	if err := dec.Decode(&json.RawMessage{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "expected exactly one JSON object")
		return
	}
	if input.Expression == "" {
		writeError(w, http.StatusBadRequest, "expression is required")
		return
	}
	result, err := calculator.CalculateInfix(input.Expression)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if errors.Is(err, calculator.ErrInvalidExpression) {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, calculateResponse{Result: result})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Response types contain only strings or finite numbers. A transport
	// failure after headers are committed cannot be reported to the client.
	_ = json.NewEncoder(w).Encode(value)
}
