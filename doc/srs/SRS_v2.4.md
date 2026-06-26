# System Requirements Specification (SRS): Production & Inventory Management v2.4

## 1. System Overview

A comprehensive system to manage raw material (RM) inventory, integrate with production Work Orders (WO), enforce **FEFO (First-Expire-First-Out) then FIFO** allocation, and provide real-time stock visibility. The system tracks physical stock via a ledger, logical availability via reservations, and all mutations via an audit trail.

**Phase 1 Scope:** Flat RM reservation and issuance against WOs. SFG on-hand is tracked but not consumed by parent WOs (see §3.4). Multi-level SFG consumption is deferred to a later phase.

---

## 2. Normalized Database Schema

### 2.1 Item Master & BOM

**Table A1: Item Master** — Single source of truth for all items regardless of type.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `Item_Code` | PK | Unique identifier (e.g., 12003750) |
| `Description` | Text | |
| `Unit` | Text | EA, RL, Kg, etc. |
| `Item_Type` | Enum | `FG`, `SFG`, `RM` |
| `Min_Stock_Level` | Numeric | Threshold for low-stock alert |
| `Expiry_Threshold_Days` | Nullable Integer | Item-level near-expiry alert window; falls back to system config if NULL |
| `Shelf_Life_Days` | Nullable Integer | FG/SFG only. Auto-computes `EXP_Date` on `WO_RECEIPT` LOT creation. NULL = non-expiring. |

**Table A2: BOM Ledger** — Relational junction defining parent-component relationships.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `BOM_ID` | PK | |
| `Parent_Item_Code` | FK → A1 | |
| `Component_Item_Code` | FK → A1 | |
| `Qty_Per_Set` | Numeric | Component qty consumed per 1 parent unit |
| `BOM_Version` | Text | e.g., "v1", "v2" |
| `Is_Active` | Boolean | *Constraint: Only one version may be `TRUE` per `Parent_Item_Code` at any time.* |

---

### 2.2 LOT Management

**Table B1: LOT Master** — One record per physical batch. MFG/EXP dates live here, not on the transaction, to prevent duplication on re-receipt.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `LOT_Internal_ID` | PK | System-generated |
| `Item_Code` | FK → A1 | |
| `Supplier_LOT_Number` | Text | Uniqueness: unique *per* `Item_Code` only |
| `Supplier_Name` | Text | NULL for internally produced LOTs |
| `MFG_Date` | Date | Required |
| `EXP_Date` | Nullable Date | NULL = non-expiring |
| `Status` | Enum | `Active`, `Hold`, `Quarantined` |

*Re-receipt rule: If a `Supplier_LOT_Number` already exists for the same `Item_Code` but the incoming paperwork shows a different `MFG_Date` or `EXP_Date`, the system blocks the receipt and raises a warning for manual review before proceeding.*

---

### 2.3 Inventory Ledger

**Table B2: Stock Ledger** — Immutable, append-only record of every quantity movement. Physical stock is always derived from this table.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `Transaction_ID` | PK | |
| `Timestamp` | DateTime | |
| `User_ID` | FK → User | |
| `Transaction_Type` | Enum | See types below |
| `LOT_Internal_ID` | FK → B1 | |
| `Qty_Changed` | Numeric | Positive = stock in; Negative = stock out |
| `Doc_Ref` | Text | WO Number, PO Number, or Cycle Count ID |
| `Reason_Code` | Nullable Text | Required for `ADJ_IN`, `ADJ_OUT`. e.g., `Audited`, `Line_Waste`, `Supplier_Damage`, `Expired`, `Cycle_Count` |

**Transaction types and their sign:**

| Type | Sign | Triggered By |
| :--- | :--- | :--- |
| `PO_RECEIPT` | + | Receiving goods from supplier |
| `WO_ISSUE` | − | RM consumed by a WO at completion |
| `WO_RECEIPT` | + | FG produced by a WO at completion |
| `RETURN_TO_STOCK` | + | Unused reserved RM returned after WO completion |
| `ADJ_IN` | + | Cycle count gain, manual correction |
| `ADJ_OUT` | − | Supplier damage, line waste, scrap, cycle count loss |

---

### 2.4 Work Order Tracking & Audit

**Table C1: WO Header**

| Field | Type | Notes |
| :--- | :--- | :--- |
| `WO_Number` | PK | |
| `Target_FG_Code` | FK → A1 | Must be `Item_Type = FG` or `SFG` |
| `Target_Qty` | Numeric | Planned production output |
| `Actual_Produced_Qty` | Nullable Numeric | Filled at close. Used for `WO_RECEIPT` qty. |
| `Completion_Pct` | Nullable Numeric | Stored at close: `(Actual_Produced_Qty / Target_Qty) × 100`. NULL while open. |
| `WO_Status` | Enum | `DRAFT`, `RESERVED`, `IN_PRODUCTION`, `PENDING_APPROVAL`, `COMPLETED`, `COMPLETED_PARTIAL`, `CANCELLED` |

**Table C1.5: WO Material Requirements** — Aggregated RM needs after BOM explosion. One row per item.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `Req_ID` | PK | |
| `WO_Number` | FK → C1 | |
| `Required_Item_Code` | FK → A1 | Always `Item_Type = RM` in Phase 1 |
| `Total_Needed_Qty` | Numeric | BOM qty rolled up across all levels |

**Table C2: WO LOT Allocations** — Per-LOT split of each material requirement. One row per (Req_ID, LOT).

| Field | Type | Notes |
| :--- | :--- | :--- |
| `Allocation_ID` | PK | |
| `Req_ID` | FK → C1.5 | |
| `LOT_Internal_ID` | FK → B1 | |
| `Reserved_Qty` | Numeric | Logical hold; does not touch B2 |
| `Actual_Used_Qty` | Nullable Numeric | Entered at completion |
| `Damage_Qty` | Nullable Numeric | Line waste entered at completion |

**Table D1: WO Audit Log** — Append-only record of every change to a WO.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `Log_ID` | PK | |
| `WO_Number` | FK → C1 | |
| `User_ID` | FK → User | |
| `Timestamp` | DateTime | |
| `Action` | Text | `Status_Change`, `Qty_Edit`, `Manual_Override`, `Reopen`, `Approval_Request`, `Approval_Resolved` |
| `Old_Value` | Text | JSON or plain string of prior state |
| `New_Value` | Text | JSON or plain string of new state |

**Table D2: WO Approval Requests** — Supervisor sign-off for actions requiring elevated authority.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `Approval_ID` | PK | |
| `WO_Number` | FK → C1 | |
| `Approval_Type` | Enum | `PARTIAL_COMPLETION` *(extensible for future approval types)* |
| `Requested_By` | FK → User | |
| `Requested_At` | DateTime | |
| `Completion_Pct_At_Request` | Numeric | Snapshot of `Actual / Target × 100` at the moment of request |
| `Approval_Status` | Enum | `PENDING`, `APPROVED`, `REJECTED` |
| `Approver_User_ID` | Nullable FK → User | Filled when resolved |
| `Resolved_At` | Nullable DateTime | |
| `Supervisor_Notes` | Nullable Text | Reason for approval or rejection |

---

### 2.5 System Configuration

**Table E1: System Config** — Global default values for system-wide behavior.

| Config_Key | Default | Description |
| :--- | :--- | :--- |
| `NEAR_EXPIRY_DAYS_DEFAULT` | 30 | Fallback alert window when `Expiry_Threshold_Days` is NULL on an item |
| `LOW_STOCK_ALERT_ENABLED` | TRUE | Global toggle |
| `NEAR_EXPIRY_ALERT_ENABLED` | TRUE | Global toggle |

---

## 3. Core Business Logic & Workflows

### 3.1 Stock Calculations & Constraints

* **Stock On Hand (Physical):** `SUM(Qty_Changed)` from B2, grouped by `LOT_Internal_ID`.
* **Available Stock (Logical):** `Stock On Hand` MINUS `SUM(Reserved_Qty)` from C2 where the linked WO has `WO_Status IN ('RESERVED', 'IN_PRODUCTION', 'PENDING_APPROVAL')`. Reservations remain held even while awaiting supervisor approval.
* **Negative Stock Prevention:** Row-level locking is enforced during reservation to prevent concurrent WOs from double-reserving the same available quantity. No `WO_ISSUE` or `ADJ_OUT` transaction may be written if it would reduce a LOT's Physical Stock below 0.

### 3.2 Allocation Sort Rules (FEFO + FIFO)

When auto-reserving LOTs, eligible candidates from B1 must satisfy: `Available Stock > 0` AND `Status = 'Active'`.

Reservation fills each LOT to zero before moving to the next:

1. **Primary (FEFO):** `EXP_Date` Ascending — soonest-to-expire first. `NULL` (non-expiring) sorted last.
2. **Secondary (FIFO):** `MFG_Date` Ascending — oldest manufactured first, as tiebreaker.

### 3.3 LOT Selection: Auto with Manual Option

All reservation events — initial allocation at `RESERVED` and delta allocation during over-usage resolution — follow the same two-step pattern:

1. **Auto-propose:** System runs FEFO+FIFO and presents a proposed LOT selection to the user, showing each LOT, its EXP/MFG dates, and the qty to be reserved from it.
2. **User decision:** The user may:
   - **Accept** the auto-proposal → proceeds as-is. No override logged.
   - **Change** one or more LOTs to a different `Active` LOT of their choice → proceeds with the user's selection. Logged to D1 as `Action = 'Manual_Override'`.

System constraints always apply regardless of the user's choice: `Hold`/`Quarantined` LOTs cannot be selected, and the selection cannot exceed available stock.

### 3.4 BOM Explosion Scope

On entering `DRAFT`, the system queries A2 for active records (`Is_Active = TRUE`) under `Target_FG_Code`. It performs a **recursive walk** — FG → SFG → RM — multiplying `Qty_Per_Set` at each level until only `RM` leaf nodes remain. All RM quantities are aggregated into C1.5.

**Phase 1 SFG rule:** SFG on-hand inventory is tracked (received, issued, reported) but is **not consumed** by a parent WO. The explosion always walks through SFGs to their RM components. SFG consumption by parent WOs is deferred.

### 3.5 Receiving & Damage

**Short shipment** (ordered 100, received 60): Record `PO_RECEIPT` for 60 only. No `ADJ_OUT`. The quantity difference is a procurement matter outside this system's scope.

**Supplier damage on arrival** (received 100, but 10 are damaged upon inspection): Record `PO_RECEIPT` for 100, then a separate `ADJ_OUT` for 10 with `Reason_Code = 'Supplier_Damage'`. Physical stock settles at 90 with a full audit trail.

**Cycle count adjustments:** Use `ADJ_IN` or `ADJ_OUT` with `Reason_Code = 'Cycle_Count'`. A `Reason_Code` is mandatory for all adjustment types.

### 3.6 LOT Status Management

LOT status controls eligibility for new reservations. All status changes are logged in D1.

| Transition | Who | Effect on Existing Reservations |
| :--- | :--- | :--- |
| `Active` → `Hold` | QC / Supervisor | Reservations on the LOT remain. System raises an alert to the linked WO owner. |
| `Hold` → `Active` | QC / Supervisor | Reservations resume normally. |
| `Active` / `Hold` → `Quarantined` | Supervisor only | Reservations on the LOT are flagged. User must manually reassign to another LOT or cancel the WO. |
| `Quarantined` → `Hold` | Supervisor only | Intermediate step before releasing back to Active. |

`Hold` and `Quarantined` LOTs are excluded from auto-allocation (§3.2) and cannot be manually selected (§3.3).

---

## 4. Work Order Lifecycle & State Machine

### 4.1 Allowed Status Transitions

```
DRAFT → RESERVED → IN_PRODUCTION → COMPLETED               (full production: Actual == Target)
                                  → PENDING_APPROVAL        (partial: Actual < Target)
PENDING_APPROVAL → COMPLETED_PARTIAL                        (supervisor approves)
PENDING_APPROVAL → IN_PRODUCTION                            (supervisor rejects)
DRAFT | RESERVED | IN_PRODUCTION | PENDING_APPROVAL → CANCELLED
CANCELLED → DRAFT                                           (Reopen)
```

* `COMPLETED` and `COMPLETED_PARTIAL` are **terminal states**. A WO in either state cannot be reopened or modified.
* Skipping `IN_PRODUCTION` (`RESERVED → COMPLETED` directly) is **not permitted**.

### 4.2 Lifecycle Mapping

---

**1. DRAFT — WO Creation & BOM Explosion**

- User creates WO: inputs `WO_Number`, `Target_FG_Code`, `Target_Qty`.
- System runs recursive BOM explosion → populates C1.5.
- No reservations exist. No ledger entries written.

---

**2. RESERVED — Auto-Allocation**

- System runs FEFO+FIFO and presents proposed LOT allocation (§3.3). User accepts or overrides.
- If any material has insufficient available stock, the transition is **blocked** and the shortage is flagged per material in red. WO stays in `DRAFT`.
- On success, C2 rows are created with `Reserved_Qty`. Available Stock drops logically.
- D1 entry: `Status_Change`, `DRAFT → RESERVED`.

---

**3. IN_PRODUCTION — Active Production**

- Physical production begins. C2 reservations continue to hold stock.
- Users enter `Actual_Used_Qty` and `Damage_Qty` on C2 lines incrementally as production progresses.

**Over-Usage Resolution** (when actuals will exceed reserved):
- User signals that a material line will exceed its `Reserved_Qty`.
- System auto-proposes FEFO+FIFO allocation for the **delta** (extra qty needed). User accepts or changes to a different LOT (§3.3).
- `Reserved_Qty` on existing C2 rows is increased, or new C2 rows are added for newly allocated LOTs.
- Each change is logged to D1: `Action = 'Qty_Edit'` (auto-accept) or `'Manual_Override'` (user-changed LOT).

---

**4a. COMPLETED — Full Production Close**

*Triggered when user submits actuals and `Actual_Produced_Qty == Target_Qty` (or user explicitly confirms full completion).*

**Completion Guard (runs on every C2 line):**
> Blocked if `(Actual_Used_Qty + Damage_Qty) > Reserved_Qty`. User must resolve over-usage first (step 3).

When all lines pass, the system writes to B2 in this order per C2 line:

| Ledger Write | Type | Qty | Condition |
| :--- | :--- | :--- | :--- |
| RM consumed | `WO_ISSUE` | `Actual_Used_Qty` (negative) | Always |
| Line waste | `ADJ_OUT` | `Damage_Qty` (negative) | If `Damage_Qty > 0` |
| Return unused | `RETURN_TO_STOCK` | `Reserved_Qty − Actual_Used_Qty − Damage_Qty` (positive) | If remainder > 0 |

Then for the WO header:
- System auto-creates a B1 LOT for `Target_FG_Code`:
  - `Supplier_LOT_Number` = `WO_Number`, `Supplier_Name` = NULL
  - `MFG_Date` = completion date
  - `EXP_Date` = `MFG_Date + Shelf_Life_Days` (from A1); NULL if not defined
  - `Status` = `Active`
- Writes `WO_RECEIPT` to B2: qty = `Actual_Produced_Qty`, linked to new FG LOT.
- Stores `Completion_Pct = 100.00` on C1.
- All C2 `Reserved_Qty` values set to 0.
- D1 entry: `Status_Change`, `IN_PRODUCTION → COMPLETED`.

---

**4b. PENDING_APPROVAL — Partial Completion Request**

*Triggered when user submits actuals and `Actual_Produced_Qty < Target_Qty`.*

- System displays a prominent warning:
  > "You are closing this WO at **X%** completion (Produced: Y / Target: Z). This action is irreversible and requires supervisor approval."
- User must confirm the warning to proceed.
- WO status moves to `PENDING_APPROVAL`. Reservations remain active.
- An entry is created in **Table D2**: `Approval_Type = 'PARTIAL_COMPLETION'`, `Approval_Status = 'PENDING'`, `Completion_Pct_At_Request` snapshot recorded.
- D1 entry: `Approval_Request`, recording actuals and completion percentage.
- Supervisor receives a notification to review.

---

**4c. COMPLETED_PARTIAL — Supervisor Approves**

- Supervisor reviews the D2 request, optionally adds `Supervisor_Notes`, and approves.
- D2 `Approval_Status` → `APPROVED`. `Approver_User_ID` and `Resolved_At` filled.
- System executes the **same ledger write sequence as §4.2 step 4a**, using the partial actuals.
- `Completion_Pct` is stored on C1 (e.g., 72.50 for 72.5% produced).
- WO status → `COMPLETED_PARTIAL`.
- D1 entry: `Approval_Resolved`, `PENDING_APPROVAL → COMPLETED_PARTIAL`.

**4d. Supervisor Rejects**

- Supervisor rejects with a mandatory `Supervisor_Notes` explaining why.
- D2 `Approval_Status` → `REJECTED`.
- WO status returns to `IN_PRODUCTION`. No ledger writes. Reservations continue unchanged.
- D1 entry: `Approval_Resolved`, `PENDING_APPROVAL → IN_PRODUCTION`.
- User can continue production or re-submit a partial close request with revised actuals.

---

**5. CANCELLED — Release**

- All C2 `Reserved_Qty` values drop to 0. Stock becomes available to other WOs immediately.
- If WO was in `PENDING_APPROVAL`, the outstanding D2 request is also marked `REJECTED` (system-auto, reason = `WO_CANCELLED`).
- C1.5 requirements are preserved in the database for reference.
- D1 entry: `Status_Change`, `{prior status} → CANCELLED`.

---

**6. CANCELLED → DRAFT — Reopen**

- WO returns to `DRAFT`. C1.5 is preserved; C2 rows are cleared.
- User must trigger the Reservation step again. FEFO+FIFO re-runs against **current** available stock (previously held stock may have been claimed by other WOs).
- D1 entry: `Action = 'Reopen'`.

---

## 5. Alerts & Reports

### 5.1 Real-Time Alerts

* **Low Stock:** `Available Stock < Min_Stock_Level` for any `Item_Code`. Fires per item, not per LOT.
* **Near Expiry:** `EXP_Date − Today ≤ Expiry_Threshold_Days` (item-level; falls back to `NEAR_EXPIRY_DAYS_DEFAULT` from E1 if NULL). Fires on `Active` LOTs only.
* **Reserved LOT on Hold/Quarantine:** Fires when a LOT linked to an active WO reservation changes status. Alert targets the WO owner.
* **Over-Reservation Blocked:** Fires when a WO cannot transition to `RESERVED` due to insufficient stock. Lists each short material and missing quantity.
* **Partial Completion Pending Approval:** Fires to supervisors when a new D2 request is submitted. Shows WO number, FG code, and completion percentage.

### 5.2 Required Reports

**Stock On Hand**
> Item Code | Description | Item Type | LOT | Supplier | MFG Date | EXP Date | LOT Status | Physical Qty | Reserved Qty | Available Qty

**Movement Ledger**
> Timestamp | Transaction Type | Item Code | LOT | Doc Ref | Qty Changed | Reason Code | User

**WO Summary**
> WO Number | FG Code | Target Qty | Actual Produced Qty | Completion % | WO Status | Created Date | Completed/Cancelled Date

**Shortage / Damage**
> Date | WO/PO Ref | Item Code | Supplier | Supplier Damage Qty | Line Waste Qty

**Audit Trail**
> Timestamp | WO Number | User | Action | Old Value | New Value

**Partial Completion Log**
> WO Number | FG Code | Target Qty | Actual Produced Qty | Completion % | Requested By | Requested At | Approved/Rejected By | Resolved At | Supervisor Notes
