package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	store := NewStore()
	// Mock Data
	store.accounts["ACCID1"] = &Account{ID: "ACCID1", Balance: 100.0, Currency: "IDR"}
	store.accounts["ACCID2"] = &Account{ID: "ACCID2", Balance: 100.0, Currency: "IDR"}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/transfers/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v1/transfers/" {
			store.handleTransfer(w, r)
			return
		}
		if r.Method == http.MethodGet {
			store.handleGetTransfer(w, r)
			return
		}
		http.NotFound(w, r)
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: loggingMiddleware(mux),
	}

	// Graceful Shutdown Logic
	go func() {
		log.Println("Service started on :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
