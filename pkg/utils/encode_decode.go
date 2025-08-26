package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/MdSadiqMd/Exchange-Rate-Service/internal/domain"
)

type errorer interface {
	Error() error
}

type httpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *httpError) Error() string {
	return e.Message
}

func DecodeConvertRequest(_ context.Context, r *http.Request) (interface{}, error) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	amountStr := r.URL.Query().Get("amount")
	date := r.URL.Query().Get("date")
	if from == "" {
		return nil, &httpError{Code: http.StatusBadRequest, Message: "from parameter is required"}
	}
	if to == "" {
		return nil, &httpError{Code: http.StatusBadRequest, Message: "to parameter is required"}
	}
	if amountStr == "" {
		return nil, &httpError{Code: http.StatusBadRequest, Message: "amount parameter is required"}
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, &httpError{Code: http.StatusBadRequest, Message: "invalid amount"}
	}

	var parsedDate time.Time
	if date != "" {
		parsedDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			return nil, &httpError{Code: http.StatusBadRequest, Message: "invalid date format, expected YYYY-MM-DD"}
		}
	}

	return domain.ConversionRequest{
		From:   from,
		To:     to,
		Amount: amount,
		Date:   parsedDate,
	}, nil
}

func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if e, ok := response.(errorer); ok && e.Error() != nil {
		EncodeError(ctx, e.Error(), w)
		return nil
	}

	return json.NewEncoder(w).Encode(response)
}

func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	code := http.StatusInternalServerError
	if httpErr, ok := err.(*httpError); ok {
		code = httpErr.Code
	}

	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	})
}
