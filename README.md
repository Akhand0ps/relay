# 🚀 Relay

> **Developer Infrastructure for High-Reliability Webhook Delivery.**  
> Publish events once, reliably deliver them everywhere.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Architecture: Clean](https://img.shields.io/badge/Architecture-Clean%20Go-blue.svg)]()

---

## 📌 What is Relay?

Relay is a hosted webhook delivery engine built in Go. Instead of every microservice building its own webhook retries, queue consumers, cryptographic signing, dead-letter queues, and replay mechanisms, Relay provides them as plug-and-play infrastructure.

Developers publish events with a single HTTP call (`POST /v1/events`), and Relay guarantees delivery to all subscribed customer endpoints—even during network outages, slow destinations, or third-party server crashes.

---

## 🏗️ Architecture & Core Flow

### Phase 1 Core Engine: Event Matching & Worker Queue
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
         ├── 1. Deduplicate & Persist Event
         ├── 2. Match Subscribed Endpoints
         ├── 3. Enqueue Delivery Tasks
         ▼
   Worker Pool (Concurrent Goroutines)
         │
         │ Sign with HMAC-SHA256 (X-Relay-Signature)
         │ HTTP POST with 10s Timeout
         ▼
  Customer Webhook Destination
         │
    ┌────┴──────────────────────────┐
    ▼                               ▼
 [2xx Success]             [Transient 5xx / 429 / Timeout]
    │                               │
  DONE                      Exponential Backoff + Jitter
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

* ⚡ **Non-Blocking Ingestion:** Returns `202 Accepted` in `<10ms` while processing deliveries asynchronously in the background.
* 🔒 **Cryptographic Webhook Signatures:** Signs every payload using `HMAC-SHA256` (`X-Relay-Signature` & `X-Relay-Timestamp`) to prevent tampering and replay attacks.
* 🛡️ **Full Idempotency Protection:** Eliminates duplicate charges and events via `Idempotency-Key` tracking.
* 🔄 **Exponential Backoff with Full Jitter:** Protects destination servers from the Thundering Herd Problem by staggering retry intervals.
* 💀 **Dead Letter Queue (DLQ):** Safely isolates permanently failing endpoints after max attempts.
* 🔁 **Immutable Replay Engine:** Replay past failed events on demand (`POST /v1/events/replay`) without mutating original audit logs.
* ⏱️ **Zero-Hang Workers:** Hardened HTTP client with 10-second timeouts prevents slow endpoints from exhausting worker threads.

---

## 📡 API Reference

### 1. Register an Endpoint
Tell Relay where to deliver specific event types.

```bash
POST /v1/endpoints
Content-Type: application/json

{
  "url": "https://api.example.com/webhooks",
  "event_types": ["invoice.paid", "user.created"]
}
```
**Response (`201 Created`):**
```json
{
  "id": "ep_01a1409d-...",
  "url": "https://api.example.com/webhooks",
  "event_types": ["invoice.paid", "user.created"],
  "secret": "whsec_01a1409d-...",
  "status": "active",
  "created_at": "2026-10-17T..."
}
```

---

### 2. Ingest an Event
Publish an event to Relay.

```bash
POST /v1/events
Content-Type: application/json
Idempotency-Key: inv_req_9981

{
  "type": "invoice.paid",
  "data": {
    "invoice_id": "inv_1837",
    "amount": 4999,
    "currency": "USD"
  }
}
```
**Response (`202 Accepted`):**
```json
{
  "id": "evt_01a1409f-...",
  "message": "Event accepted for async processing",
  "status": "queued"
}
```

---

### 3. Replay Failed Events
Re-dispatch events without modifying past audit trails.

```bash
POST /v1/events/replay
Content-Type: application/json

{
  "event_ids": ["evt_01a1409f-..."]
}
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

Every webhook delivered by Relay includes headers:
```http
X-Relay-Event-Id: evt_01a1...
X-Relay-Timestamp: 1791024000
X-Relay-Signature: v1=4f3b2ac7...
```

Customers verify authenticity using their shared endpoint secret:
```text
signature = HMAC-SHA256(secret, timestamp + "." + raw_body)
```

---

## 📂 Project Structure

```text
relay/
├── cmd/
│   └── api/
│       └── main.go                 # Application bootstrap & dependency injection
├── internal/
│   ├── api/
│   │   └── handlers.go             # HTTP handlers & route logic
│   ├── id/
│   │   └── id.go                   # Time-ordered UUIDv7 generator
│   ├── models/
│   │   ├── delivery.go             # Delivery state machine model
│   │   ├── endpoint.go             # Webhook subscriber model
│   │   └── event.go                # Ingested event model
│   ├── signature/
│   │   └── signature.go            # HMAC-SHA256 cryptographic signer
│   ├── store/
│   │   └── memory.go               # Concurrent thread-safe in-memory store
│   └── worker/
│       ├── retry.go                # Exponential backoff & jitter calculations
│       └── worker.go               # Background worker pool & dispatcher
├── go.mod
└── README.md
```

---

## 🗺️ Roadmap

- [x] Phase 1: Ingestion API, Worker Pool, Retries, Signatures, Idempotency, Replay
- [ ] Phase 2: PostgreSQL Persistence (Audit logs, Durable Storage)
- [ ] Phase 3: Redis Distributed Queue
- [ ] Phase 4: Per-tenant Concurrency & Rate Limiting (Token Bucket)
- [ ] Phase 5: Observability (Prometheus Metrics & Tracing)

---

## 📜 License
MIT © Akhand Pratap Singh
