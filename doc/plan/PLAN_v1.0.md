# OptiStock — Development Plan v1.0

**Stack:** Go + Fiber · PostgreSQL · React + TypeScript + Ant Design  
**References:** `doc/srs/SRS_v2.5.md` · `doc/architecture/ARCH_v1.0.md`

---

## Phase Overview

| Phase | Name | Deliverable | Est. Days |
| :---: | :--- | :--- | :---: |
| 0 | Foundation | Running dev environment, auth, CI | 3 |
| 1 | Item & LOT Management | Full inventory lifecycle, receiving | 4 |
| 2 | Work Orders | WO lifecycle from DRAFT → COMPLETED | 6 |
| 3 | Approval Flow | Partial close, supervisor sign-off | 3 |
| 4 | Alerts & Reports | Real-time alerts, all 6 reports, Excel export | 4 |
| 5 | Deployment | On-prem production, backups, admin panel | 2 |
| | **Total** | | **~22 days** |

---

## Phase 0 — Foundation

**Goal:** Every developer can run the full stack locally with one command. Auth works end-to-end.

### Backend
- [x] `go mod init` — install Fiber, pgx/v5, golang-migrate, golang-jwt, bcrypt, air (hot reload)
- [x] Project folder structure per `ARCH_v1.0.md §4`
- [x] `internal/config/` — load env vars from `.env`
- [x] `internal/database/` — pgxpool setup + migration runner
- [x] First migration: `000001_create_users.up.sql` (id, name, email, password_hash, role, created_at)
- [x] `internal/auth/` — JWT RS256 issue + verify, bcrypt hash/compare
- [x] Auth endpoints: `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`
- [x] Middleware: JWT guard, role guard, request logger
- [x] `pkg/apperror/` — typed error codes from `ARCH_v1.0.md §6`

### Frontend
- [x] `npm create vite` — React + TypeScript
- [x] Install: Ant Design, Zustand, Axios, React Router
- [x] Login page — calls `/auth/login`, stores JWT in memory + refresh token in httpOnly cookie
- [x] Auth context / Zustand slice — current user, role, logout
- [x] Route guard component — redirects to login if unauthenticated
- [x] Shell layout — sidebar nav, top bar, notification area (placeholder)

### Infrastructure
- [x] `docker-compose.dev.yml` — services: `postgres`, `api` (air), `frontend` (vite)
- [x] `Dockerfile.backend` (dev stage with air)
- [x] `Dockerfile.frontend` (dev stage with vite)
- [x] `.env.example` — all required env vars documented
- [x] golangci-lint config, ESLint + Prettier config

**Done when:** `docker compose up` starts all services; login returns a JWT; protected route rejects unauthenticated requests.

---

## Phase 1 — Item & LOT Management

**Goal:** Warehouse staff can manage items, BOMs, lots, and receive stock. Stock On Hand is visible.

### Migrations
- [x] `000002_create_a1_items.up.sql` — Table A1
- [x] `000003_create_a2_bom.up.sql` — Table A2 + unique constraint (one active version per parent)
- [x] `000004_create_b1_lots.up.sql` — Table B1
- [x] `000005_create_b2_ledger.up.sql` — Table B2 (append-only; add trigger to block UPDATE/DELETE)
- [x] `000006_create_stock_views.up.sql` — `v_lot_physical_stock`, `v_lot_available_stock`, `v_item_available_stock`
- [x] `000007_create_e1_config.up.sql` — Table E1 + seed defaults (NEAR_EXPIRY_DAYS_DEFAULT = 30)

### Backend — Items (A1, A2)
- [x] `domain/item/` — CRUD for Item Master (list with filter by type, get, create, update)
- [x] `domain/item/` — BOM Ledger CRUD (list by parent, create version, activate version)
- [x] BOM version activation enforces the one-active-per-parent constraint

### Backend — LOTs (B1)
- [x] `domain/lot/` — LOT Master CRUD (list with filters, get, create on first receipt)
- [x] LOT status transitions (Active ↔ Hold, Active/Hold → Quarantined, Quarantined → Hold) with role guard
- [x] Re-receipt mismatch rule — block and return warning if MFG/EXP dates differ from existing record

### Backend — Ledger & Receiving (B2)
- [x] `domain/ledger/` — read-only query API (list transactions by LOT or item, date range)
- [x] `domain/receiving/` — `POST /receiving/po` — PO_RECEIPT + optional ADJ_OUT for supplier damage
- [x] `domain/receiving/` — `POST /receiving/adjustment` — ADJ_IN / ADJ_OUT with mandatory Reason_Code
- [x] Negative stock prevention enforced in receiving service before insert

### Frontend
- [x] Item Master page — table with Item_Type filter, create/edit modal
- [x] BOM page — tree view per parent item, add/activate version
- [x] LOT Master page — table with status badges, status-change action (role-gated)
- [x] Receiving page — PO Receipt form (item, supplier lot, qty, damage optional)
- [x] Adjustments page — ADJ_IN / ADJ_OUT form with Reason_Code dropdown
- [x] Stock On Hand page — table from `v_lot_available_stock` (physical vs available), exportable

**Done when:** A item can be created, a LOT received via PO, stock appears in Stock On Hand, and a manual adjustment updates it correctly.

---

## Phase 2 — Work Orders

**Goal:** Full WO lifecycle works end-to-end — DRAFT through COMPLETED — with correct ledger writes and audit trail.

### Migrations
- [x] `000008_create_c1_wo_header.up.sql` — Table C1
- [x] `000009_create_c1_5_wo_requirements.up.sql` — Table C1.5
- [x] `000010_create_c2_wo_allocations.up.sql` — Table C2
- [x] `000011_create_d1_audit_log.up.sql` — Table D1
- [x] `000012_update_stock_views_reservations.up.sql` — C2 reservations in available stock views

### Backend — BOM Explosion
- [x] Recursive BOM walk service: FG → SFG → RM, multiply Qty_Per_Set at each level, aggregate to RM
- [x] Guard: only `Is_Active = TRUE` BOM rows used; error if no active BOM found

### Backend — WO Lifecycle
- [x] `domain/workorder/` — create WO (DRAFT), auto-trigger BOM explosion → populate C1.5
- [x] FEFO+FIFO allocation engine:
  - Query `v_lot_available_stock` filtered by item, Status = Active, available > 0
  - Order: EXP_Date ASC NULLS LAST, MFG_Date ASC
  - `SELECT ... FOR UPDATE SKIP LOCKED` inside transaction
  - Fill C2 rows sequentially until Total_Needed_Qty met
- [x] `DRAFT → RESERVED` — run allocation, return auto-proposal to frontend; accept or override
- [x] Manual override — user replaces proposed LOT(s); logged to D1 as `Manual_Override`
- [x] `RESERVED → IN_PRODUCTION` — status change, D1 entry
- [x] Over-usage resolution — delta allocation (same FEFO+FIFO on extra qty), accept or override
- [x] `IN_PRODUCTION → COMPLETED`:
  - Completion guard: `(Actual_Used_Qty + Damage_Qty) <= Reserved_Qty` on every C2 line
  - Ledger writes per C2 line: WO_ISSUE, ADJ_OUT (if damage > 0), RETURN_TO_STOCK (if remainder > 0)
  - Auto-create FG LOT in B1 (Supplier_LOT_Number = WO_Number, MFG_Date = now, EXP_Date from Shelf_Life_Days)
  - Write WO_RECEIPT to B2 for FG lot
  - Set Completion_Pct = 100.00, zero all C2 Reserved_Qty
  - D1 entry: Status_Change
- [x] `* → CANCELLED` — zero C2 Reserved_Qty, D1 entry
- [x] `CANCELLED → DRAFT` (reopen) — clear C2, preserve C1.5, D1 entry as `Reopen`
- [x] Audit Log read API — `GET /workorders/:id/audit`

### Frontend
- [x] WO List page — table with status filter, Completion_Pct column for closed WOs
- [x] WO Create form — FG Code picker, Target Qty
- [x] WO Detail page:
  - Header: status badge, action buttons (Reserve, Start, Complete, Cancel)
  - Material requirements table (C1.5) with RED flag on shortage
  - LOT allocation table (C2) — auto-proposal display with LOT swap option per line
  - Actuals input form (Actual_Used_Qty, Damage_Qty per C2 line)
- [x] Audit Log tab on WO Detail — full D1 history for the WO

**Done when:** A WO can be created, reserved (FEFO+FIFO allocates correctly), moved to IN_PRODUCTION, completed with actuals, ledger shows WO_ISSUE + RETURN_TO_STOCK, and FG stock increases.

---

## Phase 3 — Approval Flow

**Goal:** Partial WO close triggers supervisor sign-off. Full re-appeal and withdrawal cycle works.

### Migrations
- [x] `000013_create_d2_approvals.up.sql` — Table D2

### Backend
- [x] Partial completion detection — if `Actual_Produced_Qty < Target_Qty` on complete attempt, redirect to approval flow
- [x] `PENDING_APPROVAL` transition — insert D2 row (PENDING), D1 entry `Approval_Request`
- [x] `POST /approvals/:id/withdraw` (requester only) — D2 → WITHDRAWN, WO → IN_PRODUCTION, D1 `Approval_Withdrawn`
- [x] `POST /approvals/:id/approve` (supervisor) — D2 → APPROVED, execute ledger writes same as COMPLETED, WO → COMPLETED_PARTIAL, D1 `Approval_Resolved`
- [x] `POST /approvals/:id/reject` (supervisor, mandatory Resolution_Notes) — D2 → REJECTED, WO → IN_PRODUCTION, D1 `Approval_Resolved`
- [x] Re-appeal — new D2 row per re-submission; old rows unchanged
- [x] Cancel during PENDING_APPROVAL — auto-withdraw D2 (Resolved_By = system, Resolution_Notes = WO_CANCELLED)
- [x] `GET /approvals` — supervisor dashboard: list PENDING requests with WO details

### Frontend
- [x] Partial close modal — warning with completion %, confirmation required, "Submit for Approval" button
- [x] Withdraw button on WO Detail (visible to requester while PENDING)
- [x] Supervisor Approval dashboard — list of PENDING requests, approve/reject with notes input
- [x] Re-appeal — "Re-submit" button appears on WO Detail after rejection, opens revised actuals form
- [x] Partial Completion Log tab on WO Detail — full D2 history with statuses and notes

**Done when:** A WO at 60% triggers approval request, supervisor approves, ledger writes at 60%, COMPLETED_PARTIAL status saved. Rejection returns to IN_PRODUCTION. Withdrawal dismisses the request. Re-appeal creates a new D2 row.

---

## Phase 4 — Alerts & Reports

**Goal:** Supervisors and users see real-time alerts. All 6 reports export to Excel and CSV.

### Backend — Alerts
- [ ] `internal/ws/` — WebSocket hub (register/unregister clients, broadcast by role)
- [ ] `GET /api/v1/ws` — WebSocket upgrade endpoint (JWT validated at handshake)
- [ ] Alert rule: **Low Stock** — triggered after every B2 write; check `v_item_available_stock < Min_Stock_Level`
- [ ] Alert rule: **Near Expiry** — daily scheduled job + triggered on LOT create/update
- [ ] Alert rule: **LOT on Hold/Quarantine** — triggered on B1 status change; targets affected WO owners
- [ ] Alert rule: **Over-Reservation Blocked** — triggered when RESERVED transition fails
- [ ] Alert rule: **Approval Pending** — triggered on D2 INSERT (PENDING); targets supervisors
- [ ] Alert rule: **Approval Withdrawn** — triggered on D2 WITHDRAWN; targets supervisors

### Backend — Reports
- [ ] `domain/report/` — one handler per report, shared export utilities in `pkg/export/`
- [ ] Report: **Stock On Hand** — from `v_lot_available_stock`, filters: item type, lot status, near-expiry flag
- [ ] Report: **Movement Ledger** — from B2, filters: date range, transaction type, item, lot
- [ ] Report: **WO Summary** — C1 join C1.5, filters: status, date range, FG code
- [ ] Report: **Shortage / Damage** — B2 WHERE type IN (ADJ_OUT) + Reason_Code filter
- [ ] Report: **Audit Trail** — D1, filters: WO number, action type, date range, user
- [ ] Report: **Partial Completion Log** — D2 join C1, filters: date range, approval status
- [ ] Excel export — `excelize` library, one sheet per report, column headers match SRS §5.2
- [ ] CSV export — standard encoding, same column order as Excel
- [ ] Cloud upload — after export, optionally push to S3/MinIO under `reports/YYYY-MM-DD/`

### Frontend
- [ ] `hooks/useAlerts.ts` — Zustand slice for alerts, WebSocket connection with auto-reconnect
- [ ] Notification panel — drawer showing recent alerts, unread count badge on bell icon
- [ ] Alert toast — Ant Design notification for incoming WebSocket alerts
- [ ] Reports page — tab per report type, filter controls, "Export Excel" + "Export CSV" buttons
- [ ] Report tables — paginated, sortable columns

**Done when:** Receiving below Min_Stock_Level pushes a LOW_STOCK notification to browser without refresh. All 6 reports load with data and export correctly to Excel.

---

## Phase 5 — Deployment

**Goal:** System runs on the factory server behind Nginx. Daily backups go to S3/MinIO. Admin can manage users and config.

### Backend — Admin & Config
- [ ] `GET/POST/PATCH /api/v1/admin/users` — user management (admin only)
- [ ] `GET/PATCH /api/v1/admin/config` — system config E1 (NEAR_EXPIRY_DAYS_DEFAULT, alert toggles)
- [ ] Seed script — `scripts/seed.sql` with E1 defaults and optional demo items

### Infrastructure
- [ ] `docker-compose.prod.yml` — services: `nginx`, `api`, `postgres` (no exposed ports except Nginx 443)
- [ ] `Dockerfile.backend` (prod stage — `go build`, minimal distroless image)
- [ ] `Dockerfile.frontend` (prod stage — `vite build`, serve via Nginx)
- [ ] `nginx.conf` — reverse proxy `/api/` → Go API; serve React static files; HTTPS (self-signed or Let's Encrypt)
- [ ] Environment variable management — `.env.prod` template, secrets not committed to git
- [ ] `scripts/backup.sh` — `pg_dump | gzip | upload to S3/MinIO`; cron job at 02:00 daily
- [ ] Health check endpoint — `GET /api/v1/health` (DB ping, version)

### Frontend — Admin
- [ ] Admin panel page (admin role only) — user list, create/edit user, role assignment
- [ ] System Config page — edit E1 values (near-expiry threshold, alert toggles)

**Done when:** `docker compose -f docker-compose.prod.yml up -d` runs on the server; factory users can access the UI via browser on LAN; daily backup uploads to cloud storage.

---

## Dependency Order

```
Phase 0  (Foundation)
    │
    ▼
Phase 1  (Items → LOTs → Ledger)
    │
    ▼
Phase 2  (BOM Explosion → Allocation Engine → WO Lifecycle)
    │
    ▼
Phase 3  (Partial Close → Approval → Re-appeal)
    │
    ├──▶ Phase 4  (Alerts & Reports)  ← can start in parallel after Phase 2
    │
    ▼
Phase 5  (Deployment)
```

Phase 4 alerts and reports can begin in parallel with Phase 3 once the ledger writes (Phase 2) are stable.
