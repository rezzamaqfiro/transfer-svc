# Money Transfer Service

## Project Structure

The project consists of three main files:

- `domain.go`: Contains the business logic, types, and thread-safe storage
- `handlers.go`: Handles HTTP requests and middleware
- `main.go`: Sets up the server and handles graceful shutdown

## How to Run the Service

### Initialize the module

First, initialize the Go module:

```bash
go mod init transfer-svc
```

### Run the server

Then, start the server:

```bash
go run .
```

The server will start on http://localhost:8080.

## How to Test via curl

You can use these commands in a second terminal window to verify the requirements:

### 1. Create a Successful Transfer

Note the use of the `X-Idempotency-Key` header for idempotency.

```bash
curl -X POST http://localhost:8080/v1/transfers/ \
-H "Content-Type: application/json" \
-H "X-Idempotency-Key: req_001" \
-d '{
    "from": "ACCID1",
    "to": "ACCID2",
    "amount": 10.50,
    "currency": "IDR"
}'
```

### 2. Test Idempotency (Re-send same request)

Run the exact same command again. You will receive the exact same `transfer_id` and balances as the first call, and the balances in the "database" won't change again.

### 3. Request Transfer Details (GET)

Replace `{id}` with the `transfer_id` received from the POST response:

```bash
curl -X GET http://localhost:8080/v1/transfers/t_1715678901234
```

## Unit Testing

To run the unit tests with verbose output:

```bash
go test . -v
```

To check for data races in the tests:

```bash
go test -race ./...
```

## Key Technical Decisions

- **Concurrency Safety**: Used `sync.RWMutex`. For the Transfer operation, I implemented Lock Ordering (sorting Account IDs) to prevent circular wait deadlocks.
- **Idempotency**: Implemented via a `map[string]*TransferResponse`. It caches the result of the first successful operation.
- **Observability**: Integrated a middleware that logs every request with its duration and uses `sync/atomic` for high-performance metric counters.
- **Graceful Shutdown**: The server listens for SIGINT and SIGTERM, allowing 5 seconds for active requests to finish before closing.