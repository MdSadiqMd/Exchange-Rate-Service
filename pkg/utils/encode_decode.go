package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
	"github.com/go-chi/chi/v5"
)

type errorer interface{ Error() error }

type httpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *httpError) Error() string { return e.Message }

func DecodeConvertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req struct {
		From   string       `json:"from"`
		To     string       `json:"to"`
		Amount domain.Money `json:"amount"`
		Date   string       `json:"date,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, &httpError{Code: http.StatusBadRequest, Message: "invalid JSON body"}
	}
	if req.From == "" || req.To == "" || req.Amount.IsZero() {
		return nil, &httpError{Code: http.StatusBadRequest, Message: "from, to and amount are required"}
	}

	var parsedDate time.Time
	if req.Date != "" {
		var err error
		parsedDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, &httpError{Code: http.StatusBadRequest, Message: "invalid date format"}
		}
	}

	return domain.ConversionRequest{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
		Date:   parsedDate,
	}, nil
}

type GetRateRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func DecodeGetRateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := GetRateRequest{
		From: chi.URLParam(r, "from"),
		To:   chi.URLParam(r, "to"),
	}
	if req.From == "" || req.To == "" {
		return nil, &httpError{Code: http.StatusBadRequest, Message: "from and to required"}
	}
	return req, nil
}

func DecodeEmptyRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return struct{}{}, nil
}

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	if e, ok := response.(errorer); ok && e.Error() != nil {
		EncodeError(ctx, e.Error(), w)
		return nil
	}

	return json.NewEncoder(w).Encode(response)
}

func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	code := http.StatusInternalServerError
	if he, ok := err.(*httpError); ok {
		code = he.Code
	}

	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	})
}
