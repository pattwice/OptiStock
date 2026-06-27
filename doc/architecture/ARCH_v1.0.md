# Architecture Decision Record: OptiStock v1.0

## 1. System Context

OptiStock is a Production & Inventory Management system (see `doc/srs/SRS_v2.5.md`).  
Core responsibilities: raw material tracking, FEFO+FIFO lot allocation, work order lifecycle, supervisor approval flow, real-time alerts, and audit trail.

---

## 2. Architecture Style

**Layered Monolith** — a single deployable backend with clearly separated internal layers.

Chosen over microservices because:
- The domain is highly relational (ledger, reservations, WO lifecycle all share the same ACID transaction boundary)
- Team size and operational complexity do not justify service mesh overhead
- Row-level locking for concurrent reservation requires a single database connection pool under coordinated control

Internal layers (strict dependency direction — each layer may only call the layer below it):

```
HTTP Request
    │
    ▼
┌─────────────┐
│   Handler   │  Fiber route handlers. Input validation, HTTP concerns only.
├─────────────┤
│   Service   │  Business logic. Orchestrates repositories. Owns transactions.
├─────────────┤
│ Repository  │  All SQL. Returns domain models. No business logic here.
├─────────────┤
│  Database   │  PostgreSQL via pgx connection pool.
└─────────────┘
```

---

## 3. Technology Stack

| Concern | Choice | Reason |
| :--- | :--- | :--- |
| **Backend language** | Go | Compiled, fast, excellent concurrency primitives match reservation locking needs |
| **HTTP framework** | Fiber v2 | Express-style routing, WebSocket support, low overhead |
| **Database** | PostgreSQL 16 | Row-level locking (`SELECT FOR UPDATE`), ACID, JSON columns for audit log values |
| **DB driver** | `pgx/v5` + `pgxpool` | Native PostgreSQL driver; best performance and locking support. No ORM for complex queries. |
| **Migrations** | `golang-migrate` | SQL-first, numbered up/down files, CLI and embedded support |
| **Auth** | JWT (RS256) | Stateless; role claims embedded in token (`user`, `supervisor`, `admin`) |
| **Real-time** | Fiber WebSocket + in-process hub | Push alerts to connected clients without a message broker |
| **Frontend** | React 18 + TypeScript | Component ecosystem, strong typing, wide talent pool |
| **UI library** | Ant Design (antd) | Designed for data-heavy admin UIs; includes Table, Form, Modal, Notification out of the box |
| **Frontend build** | Vite | Fast HMR for development |
| **Containerization** | Docker + Docker Compose | Consistent environments across dev, on-prem production, and CI |
| **Web server (prod)** | Nginx | Serves compiled React static files; reverse proxy to Go API |
| **Cloud storage** | S3-compatible (AWS S3 or MinIO on-prem) | Report exports and database backups |

---

## 4. Project Structure

```
OptiStock/
├── doc/
│   ├── srs/                   ← SRS version history
│   └── architecture/          ← This document and future ADRs
│
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go        ← Entry point: wires config, DB, router, starts Fiber
│   │
│   ├── internal/              ← Private application code
│   │   ├── config/            ← Env vars, app config struct
│   │   ├── database/          ← pgxpool setup, migration runner
│   │   ├── middleware/        ← JWT auth, request logging, role guard
│   │   ├── ws/                ← WebSocket hub: register/unregister clients, broadcast alerts
│   │   │
│   │   ├── domain/            ← One package per business domain
│   │   │   ├── item/          ← A1 Item Master, A2 BOM Ledger
│   │   │   │   ├── handler.go
│   │   │   │   ├── service.go
│   │   │   │   ├── repository.go
│   │   │   │   └── model.go
│   │   │   ├── lot/           ← B1 LOT Master
│   │   │   ├── ledger/        ← B2 Stock Ledger (append-only writes)
│   │   │   ├── workorder/     ← C1, C1.5, C2: WO Header, Requirements, Allocations
│   │   │   ├── approval/      ← D2 Approval Requests
│   │   │   ├── audit/         ← D1 Audit Log
│   │   │   ├── alert/         ← Alert rule evaluation, WebSocket push
│   │   │   └── report/        ← Report generation (Stock On Hand, Movement Ledger, etc.)
│   │   │
│   │   └── auth/              ← User model, JWT issue/verify, role definitions
│   │
│   ├── pkg/                   ← Shared utilities (reusable across domains)
│   │   ├── apperror/          ← Typed application errors (NotFound, Conflict, Forbidden)
│   │   ├── pagination/        ← Cursor/offset pagination helpers
│   │   └── export/            ← CSV / Excel writer utilities
│   │
│   ├── migrations/            ← SQL migration files
│   │   ├── 000001_create_items.up.sql
│   │   ├── 000001_create_items.down.sql
│   │   └── ...
│   │
│   ├── go.mod
│   └── go.sum
│
├── frontend/
│   ├── src/
│   │   ├── api/               ← Axios/fetch clients, one file per domain
│   │   ├── components/        ← Shared UI components (LotBadge, StatusTag, etc.)
│   │   ├── pages/             ← Route-level page components
│   │   │   ├── inventory/
│   │   │   ├── workorders/
│   │   │   ├── reports/
│   │   │   └── admin/
│   │   ├── hooks/             ← Custom React hooks (useStock, useWO, useAlerts)
│   │   ├── store/             ← Global state (Zustand or React Context)
│   │   ├── types/             ← TypeScript interfaces mirroring backend models
│   │   └── main.tsx
│   ├── index.html
│   ├── vite.config.ts
│   └── package.json
│
├── docker/
│   ├── Dockerfile.backend
│   ├── Dockerfile.frontend
│   ├── docker-compose.dev.yml     ← Local development (hot reload, no Nginx)
│   └── docker-compose.prod.yml    ← On-prem production (Nginx, no exposed ports)
│
└── scripts/
    ├── seed.sql                   ← Initial data (system config E1, test items)
    └── backup.sh                  ← pg_dump → S3/MinIO upload
```

---

## 5. Database Design Principles

### 5.1 Immutable Ledger (B2)
Table B2 (`stock_ledger`) is **append-only**. No `UPDATE` or `DELETE` statements are permitted at the repository layer. Physical stock is always computed as `SUM(qty_changed)`. This is enforced by:
- Repository method signature returns only `INSERT`
- A PostgreSQL trigger can optionally block `UPDATE`/`DELETE` as a safety net

### 5.2 Reservation Locking Pattern
The allocation service uses `SELECT ... FOR UPDATE SKIP LOCKED` within a transaction to prevent concurrent WOs from double-reserving the same LOT:

```sql
BEGIN;
  SELECT lot_internal_id, available_qty
  FROM   v_lot_available_stock
  WHERE  item_code = $1
    AND  status = 'Active'
    AND  available_qty > 0
  ORDER  BY exp_date ASC NULLS LAST, mfg_date ASC
  FOR UPDATE SKIP LOCKED;

  -- allocate and write C2 rows inside the same transaction
COMMIT;
```

`SKIP LOCKED` means a concurrent WO gets the next unlocked LOT rather than blocking, keeping throughput high.

### 5.3 Computed Views
Frequently used derived values are exposed as PostgreSQL **views** to avoid repeating aggregate logic in Go code:

| View | Formula |
| :--- | :--- |
| `v_lot_physical_stock` | `SUM(qty_changed)` per lot from B2 |
| `v_lot_available_stock` | Physical stock − `SUM(reserved_qty)` from active WOs in C2 |
| `v_item_available_stock` | Available stock aggregated per item (for low-stock alerts) |

### 5.4 Migration Naming Convention
```
{6-digit-seq}_{action}_{table}.{up|down}.sql
000001_create_a1_items.up.sql
000002_create_a2_bom.up.sql
000003_create_b1_lots.up.sql
...
```

---

## 6. API Design

**Style:** REST over HTTPS.  
**Versioning:** URL prefix — `/api/v1/...`  
**Auth:** Bearer JWT in `Authorization` header. Role checked per route by middleware.

### Route Groups

| Prefix | Domain | Min Role |
| :--- | :--- | :--- |
| `/api/v1/items` | A1 Item Master | user |
| `/api/v1/bom` | A2 BOM Ledger | user |
| `/api/v1/lots` | B1 LOT Master | user |
| `/api/v1/ledger` | B2 Stock Ledger (read) | user |
| `/api/v1/receiving` | PO Receipt writes to B2 | user |
| `/api/v1/workorders` | C1/C1.5/C2 full WO lifecycle | user |
| `/api/v1/approvals` | D2 Approval Requests | supervisor |
| `/api/v1/reports` | Report generation & export | user |
| `/api/v1/admin` | System config E1, user management | admin |
| `/api/v1/ws` | WebSocket upgrade endpoint | user |

### Standard Response Envelope

```json
{
  "success": true,
  "data": { ... },
  "meta": { "page": 1, "total": 240 },
  "error": null
}
```

Errors:

```json
{
  "success": false,
  "data": null,
  "error": { "code": "INSUFFICIENT_STOCK", "message": "...", "details": { ... } }
}
```

### Key Error Codes (typed in `pkg/apperror`)

| Code | HTTP | Meaning |
| :--- | :--- | :--- |
| `INSUFFICIENT_STOCK` | 409 | Available stock cannot satisfy reservation |
| `NEGATIVE_STOCK_PREVENTED` | 409 | Transaction would drop physical stock below 0 |
| `LOT_NOT_ELIGIBLE` | 409 | LOT is Hold or Quarantined |
| `WO_STATUS_INVALID` | 422 | Requested transition is not in the state machine |
| `APPROVAL_REQUIRED` | 403 | Action needs supervisor sign-off |
| `COMPLETION_GUARD_FAILED` | 422 | Actual + Damage exceeds Reserved on a C2 line |

---

## 7. Real-Time Alerts (WebSocket)

A lightweight in-process hub runs inside the Go server. No external broker needed for Phase 1.

```
Client browser
    │  ws://host/api/v1/ws  (upgrades to WebSocket, JWT validated at handshake)
    ▼
WebSocket Hub (internal/ws/hub.go)
    │  Receives alert events from Service layer via a Go channel
    ▼
Alert Service (internal/domain/alert/)
    │  Evaluates rules: low stock, near expiry, lot status change, approval pending
    ▼
Triggered by: ledger writes, lot status changes, D2 inserts
```

Alert payload shape:

```json
{
  "type": "LOW_STOCK",
  "severity": "warning",
  "payload": {
    "item_code": "12003750",
    "description": "Tape-Logo 1000 Yard",
    "available_qty": 5,
    "min_stock_level": 50
  },
  "timestamp": "2026-06-27T09:00:00Z"
}
```

Alert types: `LOW_STOCK`, `NEAR_EXPIRY`, `LOT_ON_HOLD`, `LOT_QUARANTINED`, `OVER_RESERVATION_BLOCKED`, `APPROVAL_PENDING`, `APPROVAL_WITHDRAWN`.

---

## 8. Authentication & Roles

JWT signed with RS256 (private key on server, public key distributed to frontend).  
Token payload:

```json
{
  "sub": "user-uuid",
  "name": "Somchai K.",
  "role": "supervisor",
  "exp": 1234567890
}
```

| Role | Capabilities |
| :--- | :--- |
| `user` | Read all stock/WO data, create/edit WOs (up to IN_PRODUCTION), receive/adjust stock |
| `supervisor` | All of `user` + approve/reject partial completions, change LOT status to Quarantined, access audit trail |
| `admin` | All of `supervisor` + manage users, update system config (E1), manage item master |

---

## 9. Deployment Architecture

### On-Premise (Core — always running)

```
Factory LAN
┌──────────────────────────────────────────┐
│  Linux Server (bare metal or VM)         │
│                                          │
│  ┌──────────┐    ┌──────────────────┐   │
│  │  Nginx   │───▶│  Go API (Fiber)  │   │
│  │ :443     │    │  :8080           │   │
│  └──────────┘    └────────┬─────────┘   │
│       │                   │             │
│  React static             │             │
│  files served             ▼             │
│  by Nginx         ┌──────────────┐      │
│                   │  PostgreSQL  │      │
│                   │  :5432       │      │
│                   └──────────────┘      │
└──────────────────────────────────────────┘
        │ Client browsers on factory LAN
```

### Cloud (Hybrid — backup & export)

```
┌──────────────────────────────────────┐
│  Cloud (AWS / GCP)                   │
│                                      │
│  S3 Bucket                           │
│  ├── reports/          ← exported    │
│  │     YYYY-MM-DD/       CSV/Excel   │
│  └── backups/          ← daily       │
│        pg_dump.gz        pg_dump     │
│                                      │
│  (Optional Phase 2)                  │
│  Read replica PostgreSQL             │
│  for report queries                  │
└──────────────────────────────────────┘
         ▲
         │  HTTPS / S3 API (outbound from on-prem server)
         │  scripts/backup.sh runs via cron
```

### Docker Compose (prod, on-prem)

Services: `nginx`, `api` (Go binary), `postgres`.  
The `api` container reads environment variables for DB credentials, JWT keys, S3 config.  
No ports exposed to LAN except Nginx on 443.

---

## 10. Development Workflow

| Step | Tool |
| :--- | :--- |
| Local dev | `docker-compose.dev.yml` — Postgres in container, Go with `air` hot reload, Vite dev server |
| Migrations | `golang-migrate up` / `down` against local Postgres |
| API testing | Bruno or Postman collection (to be added to `scripts/`) |
| Type safety | Frontend `types/` hand-synced with backend models (Phase 2: codegen from OpenAPI) |
| Linting | `golangci-lint` (backend), ESLint + Prettier (frontend) |

---

## 11. Resolved Decisions

| # | Decision | Resolution |
| :--- | :--- | :--- |
| 1 | Frontend state manager | **Zustand** — preferred for alert WebSocket subscriptions and cross-page WO state |
| 2 | Report export format | **Excel (.xlsx) + CSV** — `excelize` Go library for `.xlsx` generation |
| 3 | User management | **Built-in** — admin creates users inside the system (username + bcrypt password). LDAP/AD deferred to Phase 2. |
| 4 | Backup frequency | **Daily cron** — `scripts/backup.sh` runs `pg_dump` and uploads to S3/MinIO. WAL streaming deferred to Phase 2. |
