package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAccountNotFound   = errors.New("account not found")
	ErrCurrencyMismatch  = errors.New("currency mismatch")
)

type TransferResponse struct {
	TransferID  string  `json:"transfer_id"`
	Status      string  `json:"status"`
	FromBalance float64 `json:"from_balance"`
	ToBalance   float64 `json:"to_balance"`
}

type Account struct {
	sync.Mutex // Each account has its own lock
	ID         string
	Balance    float64
	Currency   string
}

type InMemoryStore struct {
	// This lock now only protects the MAPS, not the account balances
	sync.RWMutex
	accounts       map[string]*Account
	idempotencyMap map[string]*TransferResponse
	transfersByID  map[string]*TransferResponse
}

func NewStore() *InMemoryStore {
	return &InMemoryStore{
		accounts:       make(map[string]*Account),
		idempotencyMap: make(map[string]*TransferResponse),
		transfersByID:  make(map[string]*TransferResponse),
	}
}

func (s *InMemoryStore) Transfer(ctx context.Context, key string, fromID, toID string, amount float64, currency string) (*TransferResponse, error) {
	// 1. Idempotency Check (Fast path with Read Lock)
	s.RLock()
	if existing, found := s.idempotencyMap[key]; found {
		s.RUnlock()
		return existing, nil
	}

	fromAcc, ok1 := s.accounts[fromID]
	toAcc, ok2 := s.accounts[toID]
	s.RUnlock()

	if !ok1 || !ok2 {
		return nil, ErrAccountNotFound
	}

	// 2. DEADLOCK PREVENTION: Lock Ordering
	// Always lock the account with the "lexicographically smaller" ID first.
	first, second := fromAcc, toAcc
	if fromID > toID {
		first, second = toAcc, fromAcc
	}

	first.Lock()
	second.Lock()
	defer second.Unlock()
	defer first.Unlock()

	// 3. Double-Check Idempotency (Inside account locks)
	// Use a Write Lock on the store briefly to save the result later
	s.Lock()
	if existing, found := s.idempotencyMap[key]; found {
		s.Unlock()
		return existing, nil
	}
	s.Unlock()

	// 4. Safety & Business Logic
	if fromAcc.Currency != currency || toAcc.Currency != currency {
		return nil, ErrCurrencyMismatch
	}
	if fromAcc.Balance < amount {
		return nil, ErrInsufficientFunds
	}

	// 5. Atomic Balance Update
	fromAcc.Balance -= amount
	toAcc.Balance += amount

	res := &TransferResponse{
		TransferID:  fmt.Sprintf("t_%d", time.Now().UnixNano()),
		Status:      "SUCCESS",
		FromBalance: fromAcc.Balance,
		ToBalance:   toAcc.Balance,
	}

	// 6. Save to Store (Write Lock required for maps)
	s.Lock()
	s.idempotencyMap[key] = res
	s.transfersByID[res.TransferID] = res
	s.Unlock()

	return res, nil
}

func (s *InMemoryStore) GetTransfer(id string) (*TransferResponse, bool) {
	s.RLock()
	defer s.RUnlock()
	res, ok := s.transfersByID[id]
	return res, ok
}
