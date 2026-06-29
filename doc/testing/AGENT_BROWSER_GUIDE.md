# OptiStock — Browser Testing Guide (for QA / testing agents)

Quick-reference for navigating and verifying the app in a browser.  
**Base URL (local dev):** http://localhost:5173  
**API (proxied via Vite):** http://localhost:5173/api/v1/… → backend on :8080

---

## 1. Before you open the browser

| Step | Command | Pass check |
| :--- | :--- | :--- |
| Database | `.\start-db.ps1` | `docker ps` shows `optistock-postgres` |
| API | `.\start-backend.ps1` | Console shows `api listening on :8080` (no migration errors) |
| Frontend | `.\start-frontend.ps1` | Console shows Vite on port 5173 |
| Health | Open http://localhost:8080/api/v1/health | JSON `{ "success": true, "data": { "status": "ok" } }` |

**If API fails on migrations:** see [§8 Troubleshooting](#8-troubleshooting).

---

## 2. Login

| Field | Value |
| :--- | :--- |
| URL | http://localhost:5173/login |
| Email | `admin@optistock.local` |
| Password | `changeme` |
| Role | `admin` (can do everything, including LOT status changes) |

After login you land on **Dashboard** (`/`). Unauthenticated visits to any app route redirect here.

**Logout:** top-right **Logout** button → returns to `/login`.

---

## 3. Site map — direct URLs

Use these URLs to jump straight to a screen (must be logged in except `/login`).

| Area | URL | Sidebar path |
| :--- | :--- | :--- |
| Dashboard | `/` | Dashboard |
| Item Master | `/inventory/items` | Inventory → Items |
| BOM | `/inventory/bom` | Inventory → BOM |
| LOT Master | `/inventory/lots` | Inventory → LOTs |
| PO Receipt | `/inventory/receiving` | Inventory → Receiving |
| Adjustments | `/inventory/adjustments` | Inventory → Adjustments |
| Stock On Hand | `/inventory/stock` | Inventory → Stock On Hand |
| Work Order list | `/workorders` | Production → Work Orders |
| Create work order | `/workorders/new` | Work Orders → **Create WO** button |
| Work order detail | `/workorders/{WO_NUMBER}` | Click WO number in list |

**Not built yet (sidebar disabled):** Reports, Admin.

---

## 4. UI layout (what to look for)

```
┌──────────────┬─────────────────────────────────────────────┐
│ Sidebar      │ Header: title | bell (placeholder) | user   │
│ - Dashboard  ├─────────────────────────────────────────────┤
│ - Inventory▾ │                                             │
│ - Production▾│  Page content (tables, forms, tabs)        │
│ - Reports ✗  │                                             │
│ - Admin ✗    │                                             │
└──────────────┴─────────────────────────────────────────────┘
```

- **Tables:** Ant Design tables with filters at top, primary actions as buttons.
- **Forms:** Modal dialogs (Items) or full-page cards (Receiving, Create WO).
- **Toasts:** Green `message.success` / red `message.error` bottom-right on actions.
- **WO detail:** Card header + tabs (Requirements, LOT Allocations, Actuals, Audit Log).

---

## 5. Recommended test data setup

Run once before inventory / WO flows. Use unique codes if re-testing on a dirty DB.

### 5.1 Items (`/inventory/items`)

| Item code | Type | Notes |
| :--- | :--- | :--- |
| `RM-SUGAR` | RM | Raw material — receive stock |
| `SFG-MIX` | SFG | Optional intermediate |
| `FG-JUICE` | FG | Finished good for work orders |

Click **Add Item**, fill description + unit (e.g. `kg`), save.

### 5.2 BOM (`/inventory/bom`)

1. Select parent **`FG-JUICE`** (or SFG).
2. **Add row:** component `RM-SUGAR`, qty per set `2`, version `v1`.
3. **Activate** version `v1` (only one active version per parent).

Without an **active BOM**, work order creation fails.

### 5.3 Receive stock (`/inventory/receiving`)

| Field | Example |
| :--- | :--- |
| Item | `RM-SUGAR` |
| PO / Doc Ref | `PO-TEST-001` |
| Supplier Lot # | `LOT-A` |
| MFG Date | any date |
| EXP Date | optional |
| Qty Received | `100` |

Submit → toast shows new LOT id. Verify on **Stock On Hand** (`/inventory/stock`).

For FEFO testing, receive a second lot (`LOT-B`) with an **earlier EXP date** and same item.

---

## 6. Module checklists

### 6.1 Item Master — `/inventory/items`

| Action | How | Expect |
| :--- | :--- | :--- |
| List | Open page | Table of items |
| Filter | Type dropdown (FG/SFG/RM) | Table refreshes |
| Create | Add Item → modal → Save | Row appears |
| Edit | Edit on row | Values update |

### 6.2 BOM — `/inventory/bom`

| Action | How | Expect |
| :--- | :--- | :--- |
| Load BOM | Select parent FG/SFG | Table of components |
| Add version row | Add → fill component, qty, version | New row |
| Activate | Activate on version group | `is_active` tag; old version deactivated |

### 6.3 LOT Master — `/inventory/lots`

| Action | How | Expect |
| :--- | :--- | :--- |
| List | Open page | LOTs with status badges |
| Filter | Status dropdown | Filtered list |
| Change status | Status dropdown on row | **Admin/supervisor only** |

Statuses: `Active` (green), `Hold` (orange), `Quarantined` (red).  
Only **Active** LOTs are used for WO reservation.

### 6.4 PO Receipt — `/inventory/receiving`

| Action | How | Expect |
| :--- | :--- | :--- |
| Receive | Fill form → Post Receipt | Success toast with LOT id |
| Damage | Optional damage qty | ADJ_OUT with supplier damage reason |
| Mismatch | Re-receive same lot #, different MFG/EXP | Error toast (lot mismatch) |

### 6.5 Adjustments — `/inventory/adjustments`

| Action | How | Expect |
| :--- | :--- | :--- |
| ADJ_IN | Positive qty + reason code | Stock increases |
| ADJ_OUT | Positive qty (sent as out) + reason | Stock decreases; blocked if negative |

Reason codes: `Audited`, `Line_Waste`, `Supplier_Damage`, `Expired`, `Cycle_Count`.

### 6.6 Stock On Hand — `/inventory/stock`

| Action | How | Expect |
| :--- | :--- | :--- |
| View | Open page | Physical, reserved, available columns |
| Filter | Item code / status | Filtered rows |
| Export | Export CSV button | CSV file download |

After a WO is **RESERVED**, `reserved_qty` should increase and `available_qty` decrease.

### 6.7 Work Orders — `/workorders`

| Action | How | Expect |
| :--- | :--- | :--- |
| List | Open page | WO table; status filter |
| Create | Create WO → FG + target qty | Redirect to detail, status `DRAFT` |
| Requirements tab | Auto on detail | RM lines from BOM explosion; red SHORT if insufficient stock |
| Reserve | **Reserve** (DRAFT) | Confirm dialog → status `RESERVED`, allocations tab filled |
| Start | **Start Production** (RESERVED) | Status `IN_PRODUCTION` |
| Actuals tab | Enter used/damage per line | Enabled in `IN_PRODUCTION` |
| Save actuals | **Save Actuals** | Values persist |
| Complete | **Complete** (full qty only) | Status `COMPLETED`, FG stock increases |
| Cancel | **Cancel** | Status `CANCELLED`, reservations released |
| Reopen | **Reopen** (CANCELLED) | Status `DRAFT`, allocations cleared |
| Audit | Audit Log tab | Status changes and actions listed |

**WO status flow (happy path):**  
`DRAFT` → `RESERVED` → `IN_PRODUCTION` → `COMPLETED`

**Phase 3 not in UI yet:** partial complete (`Actual < Target`) returns approval-required error.

---

## 7. End-to-end scenario (copy-paste flow)

Use this as a single browser regression path (~10 min).

1. **Login** → http://localhost:5173/login  
2. **Create items** → `/inventory/items` (RM + FG minimum)  
3. **BOM** → `/inventory/bom` — FG parent, RM component, activate `v1`  
4. **Receive** → `/inventory/receiving` — 100 units RM  
5. **Stock check** → `/inventory/stock` — available = 100  
6. **Create WO** → `/workorders/new` — e.g. `WO-TEST-01`, FG, qty `10`  
7. **Reserve** → detail page → Reserve → confirm  
8. **Stock check** → reserved qty reflects RM need (BOM qty × target)  
9. **Start** → Start Production  
10. **Actuals** → tab → set actual used = reserved, damage `0`, produced qty = target  
11. **Complete** → confirm  
12. **Stock** → FG line appears; RM reduced  
13. **Audit** → Audit Log tab shows `Status_Change` entries  

---

## 8. Troubleshooting

| Symptom | Likely cause | Fix |
| :--- | :--- | :--- |
| Page spins forever on load | API/DB down | Start db + backend; check `/api/v1/health` |
| `404` on `/api/v1/...` in browser network tab | Stale API process | `.\start-backend.ps1` (kills port 8080 first) |
| `run migrations` error on API start | Dirty migration | See migration fix in repo; reset dirty flag if needed |
| Login fails | Wrong credentials / DB empty | Use seed admin; ensure Postgres running |
| Reserve fails / SHORT on requirements | Insufficient RM stock | Receive more on `/inventory/receiving` |
| Create WO fails | No active BOM | Activate BOM version for target FG |
| Complete fails | Actual + damage > reserved | Lower actuals or resolve over-usage (API only for delta) |
| LOT status change missing | Role = `user` | Login as `admin` or `supervisor` |
| Reports / Admin greyed out | Not implemented | Skip — Phase 4/5 |

**API smoke (optional, not browser):**  
Health: http://localhost:8080/api/v1/health  
Do **not** use PowerShell `curl -d '{...}'` for JSON — it mangles quotes. Use the UI or `Invoke-RestMethod`.

---

## 9. Role matrix (browser-visible)

| Feature | user | supervisor | admin |
| :--- | :---: | :---: | :---: |
| Inventory CRUD, receiving, adjustments | ✓ | ✓ | ✓ |
| LOT status change | | ✓ | ✓ |
| Work orders (full lifecycle) | ✓ | ✓ | ✓ |
| Reports / Admin menu | — | — | — |

*Only the seed **admin** account exists by default.*

---

## 10. References

| Doc | Purpose |
| :--- | :--- |
| `doc/srs/SRS_v2.5.md` | Business rules (FEFO, WO states, ledger) |
| `doc/plan/PLAN_v1.0.md` | Phase scope — what is / isn't built |
| `README.md` | Dev startup commands |

**Last updated for:** Phase 0–2 (auth, inventory, work orders). Update this guide when Reports/Admin ship.
