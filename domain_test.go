package main

import (
	"context"
	"sync"
	"testing"
)

func TestParallelTransfersWithIdempotency(t *testing.T) {
	store := NewStore()
	// Setup initial state
	store.accounts["ACCID1"] = &Account{ID: "ACCID1", Balance: 1000.00, Currency: "IDR"}
	store.accounts["ACCID2"] = &Account{ID: "ACCID2", Balance: 1000.00, Currency: "IDR"}

	const workers = 100
	const idempotencyKey = "unique-key-123"
	const transferAmount = 10.0

	var wg sync.WaitGroup
	wg.Add(workers)

	// We use a channel to collect all responses to verify they are byte-identical
	responses := make(chan *TransferResponse, workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			resp, err := store.Transfer(ctx, idempotencyKey, "ACCID1", "ACCID2", transferAmount, "IDR")
			if err != nil {
				t.Errorf("Transfer failed: %v", err)
				return
			}
			responses <- resp
		}()
	}

	wg.Wait()
	close(responses)

	// Verify Results
	firstResp := <-responses
	for resp := range responses {
		if resp.TransferID != firstResp.TransferID || resp.FromBalance != firstResp.FromBalance {
			t.Errorf("Idempotency failed: expected identical responses, but got different values")
		}
	}

	// Final check: Balance should only have decreased ONCE (1000 - 10 = 990)
	if store.accounts["ACCID1"].Balance != 990.0 {
		t.Errorf("Double debit detected! Balance: %f, Expected: 990.0", store.accounts["ACCID1"].Balance)
	}
}
