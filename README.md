# 🚀 Relay

> **Developer Infrastructure for High-Reliability Webhook Delivery.**  
> Publish events once, reliably deliver them everywhere.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![k6 Tested](https://img.shields.io/badge/k6-Load%20Tested-7D64FF?style=flat&logo=k6&logoColor=white)](https://k6.io/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Architecture: Clean](https://img.shields.io/badge/Architecture-Clean%20Go-blue.svg)]()

---

## 📌 What is Relay?

Relay is a developer-focused webhook delivery and dispatch engine written in Go. Instead of every microservice hand-rolling its own retry loops, worker queues, cryptographic signatures, dead-letter storage, and audit logs, Relay provides this out of the box as production-grade infrastructure.

Applications ingest events with a single HTTP call (`POST /v1/events`), and Relay guarantees resilient delivery to all subscribed destinations—even under severe network failures, flaky endpoints, or high concurrent load.

---

## 🏗️ Architecture & Core Flow

<p align="center">
  <img src="./docs/arch1.svg" alt="Relay Dispatch & Worker Architecture" width="800"/>
</p>

```text
 Developer Application
         │
         │ POST /v1/events (with Idempotency-Key)
         ▼
     Relay API
         │
         ├── 1. Idempotency Check (PostgreSQL)
         ├── 2. Persist Event & Fetch Subscribed Endpoints
         ├── 3. Enqueue Delivery Tasks
         ▼
   Worker Pool (Concurrent Goroutines)
         │
         │ Sign Payload with HMAC-SHA256 (X-Relay-Signature)
         │ HTTP POST with Hard 10s Timeout
         ▼
  Customer Webhook Destination
         │
    ┌────┴──────────────────────────┐
    ▼                               ▼
 [2xx Success]             [Transient 5xx / 429 / Timeout]
    │                               │
  DONE                      Exponential Backoff + Full Jitter
                                    │
                            Retry 1 (5s)
                            Retry 2 (30s)
                            Retry 3 (5m)
                            Retry 4 (30m)
                                    │
                         Max Attempts Reached?
                                    │
                                    ▼
                          Dead Letter Queue (DLQ)
                                    │
                                    ▼
                         POST /v1/events/replay
```

---

## ✨ Features Implemented

* ⚡ **Non-Blocking Ingestion:** Returns `202 Accepted` in `<10ms` while offloading matching and dispatch to background workers.
* 🐘 **Durable PostgreSQL Engine:** Production-ready persistence layer powered by `pgxpool` with connection pooling, transaction isolation, and containerized SQL migrations.
* 🛡️ **Race-Proof Idempotency:** Eliminates duplicate webhook sends via `Idempotency-Key` checking, verified against high-concurrency race conditions.
* 🔒 **Cryptographic Webhook Signatures:** Signs every outgoing request with `HMAC-SHA256` (`X-Relay-Signature` & `X-Relay-Timestamp`) to prevent tampering and replay attacks.
* 🔄 **Exponential Backoff with Full Jitter:** Prevents thundering herds during destination outages by staggering retries across randomized intervals.
* 💀 **Dead Letter Queue (DLQ) & Audit Log:** Records every delivery attempt, status code, and latency in PostgreSQL for post-mortem analysis.
* 🔁 **Immutable Replay Engine:** Replay past failed deliveries on demand (`POST /v1/events/replay`) without mutating original audit histories.
* ⏱️ **Zero-Hang Workers:** Hardened HTTP client with a 10-second timeout prevents slow endpoints from exhausting worker pools.
* 📊 **Performance & Load Tested:** Validated with **k6** load testing suites for smoke testing, sustained throughput, and concurrent idempotency.

---

## 🚀 Getting Started

### Prerequisites
- [Go 1.22+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/compose/)
- [k6](https://k6.io/docs/get-started/installation/) *(optional, for load testing)*

### 1. Start PostgreSQL
Spin up the PostgreSQL 16 container with automatic schema migrations:
```bash
docker compose up -d
```
*Migrations in `migrations/001_init.sql` execute automatically upon initialization.*

### 2. Run the Relay Service
```bash
go run cmd/api/main.go
```
The server will start listening on `http://localhost:3000`.

### 3. Verify Health
```bash
curl -i http://localhost:3000/health
```
```text
HTTP/1.1 200 OK
Content-Type: application/json

{"status":"ok"}
```

---

## 📡 API Reference

### 1. Register an Endpoint
Subscribe a destination URL to specific event topics.

```bash
curl -X POST http://localhost:3000/v1/endpoints \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://webhook.site/your-unique-url",
    "event_types": ["order.created", "invoice.paid"]
  }'
```

**Response (`201 Created`):**
```json
{
  "id": "ep_0192e471-70bf-7128-b2ef-3ad40a1fb8e0",
  "url": "https://webhook.site/your-unique-url",
  "event_types": ["order.created", "invoice.paid"],
  "secret": "whsec_0192e471-70bf-75a0-9706-e63d893ad7a2",
  "status": "active",
  "created_at": "2026-10-20T14:30:00Z"
}
```

---

### 2. Ingest an Event
Ingest a webhook payload. Relay deduplicates via `Idempotency-Key` and matches all active endpoints.

```bash
curl -X POST http://localhost:3000/v1/events \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: req_order_9981" \
  -d '{
    "type": "order.created",
    "data": {
      "order_id": "ord_1024",
      "amount": 25000,
      "currency": "INR"
    }
  }'
```

**Response (`202 Accepted`):**
```json
{
  "id": "evt_0192e472-8bfb-7892-9382-749e7bdac021",
  "message": "Event accepted for async processing",
  "status": "queued"
}
```

---

### 3. Fetch Event by ID
Query the persisted event state from PostgreSQL.

```bash
curl http://localhost:3000/v1/events/evt_0192e472-8bfb-7892-9382-749e7bdac021
```

---

### 4. Replay Events
Trigger redelivery of past events without modifying past audit trails.

```bash
curl -X POST http://localhost:3000/v1/events/replay \
  -H "Content-Type: application/json" \
  -d '{
    "event_ids": ["evt_0192e472-8bfb-7892-9382-749e7bdac021"]
  }'
```

**Response (`200 OK`):**
```json
{
  "status": "success",
  "message": "Events replayed successfully"
}
```

---

## 🔐 Webhook Verification (Customer Side)

Every webhook delivered by Relay includes security headers:
```http
X-Relay-Event-Id: evt_0192e472...
X-Relay-Timestamp: 1792435200
X-Relay-Signature: v1=5d41402abc4b2a76b9719d911017c592...
```

Customers verify the payload integrity and authenticity using their endpoint secret:
```go
expectedSig := hmacSha256(secret, timestamp + "." + rawBody)
isValid := hmac.Equal([]byte(expectedSig), []byte(receivedSig))
```

---

## 🧪 Performance & Load Testing (k6)

Relay includes automated **k6** test suites in `tests/k6/` to validate throughput, latency SLAs, and concurrency guarantees:

```bash
# 1. Smoke test: 1,000 requests across 10 VUs
k6 run tests/k6/smoke.js

# 2. Ingestion latency benchmark under sustained traffic
k6 run tests/k6/ingest.js

# 3. Concurrent idempotency test: 20 simultaneous duplicate requests
k6 run tests/k6/idempotency.js
```

---

## 📂 Project Structure

```text
relay/
├── cmd/
│   └── api/
│       └── main.go                 # App bootstrap, DB pool init & DI wiring
├── internal/
│   ├── api/
│   │   └── handlers.go             # HTTP handlers with idempotency logic
│   ├── id/
│   │   └── id.go                   # Time-ordered UUIDv7 generator (evt_, ep_, whsec_)
│   ├── models/
│   │   ├── delivery.go             # Delivery and attempt data models
│   │   ├── endpoint.go             # Endpoint subscription model
│   │   └── event.go                # Ingested event model
│   ├── signature/
│   │   └── signature.go            # HMAC-SHA256 signer & verifier
│   ├── store/
│   │   ├── memory.go               # In-memory store (for fast unit testing)
│   │   ├── postgres.go             # Production PostgreSQL store (pgxpool)
│   │   └── store.go                # Store interface abstraction
│   └── worker/
│       ├── retry.go                # Exponential backoff + full jitter algorithm
│       └── worker.go               # Concurrent worker pool & delivery dispatcher
├── migrations/
│   └── 001_init.sql                # Relational schema (events, endpoints, deliveries)
├── tests/
│   └── k6/                         # k6 performance & concurrency test suites
│       ├── idempotency.js
│       ├── ingest.js
│       └── smoke.js
├── docker-compose.yml              # PostgreSQL 16 container definition
├── go.mod
├── go.sum
└── README.md
```

---

## 🗺️ Roadmap

- [x] **Phase 1: Core Engine** (Ingestion API, Worker Pool, Retries, Signatures, Idempotency, Replay)
- [x] **Phase 2: PostgreSQL Persistence** (`pgxpool`, Migrations, Durable Deliveries & Audit Logs)
- [ ] **Phase 3: Redis Distributed Queue** (Redis Streams for crash-proof asynchronous job dispatch)
- [ ] **Phase 4: Per-tenant Concurrency & Rate Limiting** (Token Bucket / Leaky Bucket algorithms)
- [ ] **Phase 5: Observability** (Prometheus metrics, Grafana dashboards, OpenTelemetry distributed tracing)

---

## 📜 License
MIT © Akhand Pratap Singh
