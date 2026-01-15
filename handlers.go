package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	totalTransfers  uint64
	failedTransfers uint64
)

type TransferRequest struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (s *InMemoryStore) handleTransfer(w http.ResponseWriter, r *http.Request) {
	_ = r.Header.Get("X-Request-ID")
	key := r.Header.Get("X-Idempotency-Key")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Execute with domain logic
	resp, err := s.Transfer(r.Context(), key, req.From, req.To, req.Amount, req.Currency)

	if err != nil {
		atomic.AddUint64(&failedTransfers, 1)
		handleError(w, err)
		return
	}

	atomic.AddUint64(&totalTransfers, 1)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *InMemoryStore) handleGetTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Simple path parsing for /v1/transfers/{id}
	path := r.URL.Path
	const prefix = "/v1/transfers/"
	if len(path) <= len(prefix) {
		http.Error(w, "missing transfer id", http.StatusBadRequest)
		return
	}
	transferID := path[len(prefix):]

	resp, found := s.GetTransfer(transferID)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no transfer found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleError(w http.ResponseWriter, err error) {
	switch err {
	case ErrInsufficientFunds:
		w.WriteHeader(http.StatusPaymentRequired)
	case ErrCurrencyMismatch:
		w.WriteHeader(http.StatusUnprocessableEntity)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("method=%s path=%s duration=%s", r.Method, r.URL.Path, time.Since(start))
	})
}
