# In-memory HTTP Money Transfer Service
Implement a minimal HTTP service in Go which provide a money transfers between accounts in in-memory database

### Acceptance criteria.
- Correctness: no double-debit under 100 parallel requests with duplicate idempotent keys.
- Safety: (example: balances never go negative )
- Idempotent requests: same request returns byte-identical body and status on replays.
- Observability: logs show request_id and duration; metrics increments on each call.
- Code quality: small, testable units; no data races
- Extendable

### Requirements:
- Atomic balance updates
- Idempotent for client retries / re-requests
- Concurrency safety
- Context timeouts
- Basic observability

### Api specs ( can be modified if required )
---

#### 1. Create a transfer
Method: POST

Path: `/v1/transfers`
```
Request:
- from: ACCID1
- to: ACCID2
- amount: 12.34
- currency: IDR

Return (200)
- transfer_id: t_00004
- status: SUCCESS / PENDING / ERROR
- from_balance: 87.66
- to_balance: 112.34

Return (400) - validation

Return (402) - insufficient funds

Return (409) - conflict on in-flight key

Return (422) - currency mismatch

Return (500) - unexpected
```
#### 2. Request tranfer details
Method: GET

`/v1/transfers/{transfer_id}`

```
Return (200)
- transfer_id: t_00001 # uniq transaction id
- status: SUCCESS/FAIL
- from_balance: 87.66
- to_balance: 112.34

Return (404) - no transfer found

Return (500) - unexpected
```

### Functional requirements
- Re-send request should return byte-same original result
- Decimal type
- Context deadline ( 2 seconds )
- Idempotent deadline: 5 minutes
- Start with accounts and balances from JSON

### Non-functional requirements
- Structured logs per request
- Metrics ( total transfers, failed transfers )
- Graceful shutdown
