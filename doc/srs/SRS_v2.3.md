# System Requirements Specification (SRS): Production & Inventory Management v2.3

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
| `Shelf_Life_Days` | Nullable Integer | For FG/SFG only. Used to auto-compute `EXP_Date` when a WO_RECEIPT LOT is created. NULL = non-expiring. |

**Table A2: BOM Ledger** — Relational junction defining parent-component relationships.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `BOM_ID` | PK | |
| `Parent_Item_Code` | FK → A1 | |
| `Component_Item_Code` | FK → A1 | |
| `Qty_Per_Set` | Numeric | Component qty consumed per 1 parent unit |
| `BOM_Version` | Text | e.g., "v1", "v2" |
| `Is_Active` | Boolean | *Constraint: Only one version can be `TRUE` per `Parent_Item_Code` at any time.* |

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

*Re-receipt rule: If a `Supplier_LOT_Number` already exists for the same `Item_Code` but the incoming paperwork shows different `MFG_Date` or `EXP_Date`, the system blocks the receipt and raises a warning for manual review before proceeding.*

---

### 2.3 Inventory Ledger

**Table B2: Stock Ledger** — Immutable append-only record of every quantity movement. Physical stock is always derived from this table.

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
| `Actual_Produced_Qty` | Nullable Numeric | Filled at COMPLETED. Used for `WO_RECEIPT` qty. |
| `WO_Status` | Enum | `DRAFT`, `RESERVED`, `IN_PRODUCTION`, `COMPLETED`, `CANCELLED` |

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
| `Damage_Qty` | Nullable Numeric | Line waste at completion |

**Table D1: WO Audit Log** — Append-only record of all changes to a WO.

| Field | Type | Notes |
| :--- | :--- | :--- |
| `Log_ID` | PK | |
| `WO_Number` | FK → C1 | |
| `User_ID` | FK → User | |
| `Timestamp` | DateTime | |
| `Action` | Text | `Status_Change`, `Qty_Edit`, `Manual_Override`, `Reopen` |
| `Old_Value` | Text | JSON or plain string of prior state |
| `New_Value` | Text | JSON or plain string of new state |

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
* **Available Stock (Logical):** `Stock On Hand` MINUS `SUM(Reserved_Qty)` from C2 where the linked WO has `WO_Status IN ('RESERVED', 'IN_PRODUCTION')`.
* **Negative Stock Prevention:** Row-level locking is enforced during reservation to prevent concurrent WOs from double-reserving the same available quantity. No `WO_ISSUE` or `ADJ_OUT` transaction may be written if it would reduce a LOT's Physical Stock below 0.

### 3.2 Allocation Sort Rules (FEFO + FIFO)

When auto-reserving LOTs, eligible candidates from B1 must satisfy: `Available Stock > 0` AND `Status = 'Active'`.

Reservation consumes LOTs in this order, filling each to zero before moving to the next:

1. **Primary (FEFO):** `EXP_Date` Ascending — soonest-to-expire first. `NULL` (non-expiring) sorted last.
2. **Secondary (FIFO):** `MFG_Date` Ascending — oldest manufactured first, used when two LOTs share the same `EXP_Date`.

### 3.3 Manual LOT Override

A user with override permission may bypass auto-allocation and manually select which `Active` LOTs to reserve.

* **Allowed in:** `DRAFT` (before reservation) and `IN_PRODUCTION` (for over-usage resolution).
* **Restrictions:** System still enforces the negative stock and Hold/Quarantine guards; a user cannot override those.
* **Audit:** Every manual override writes to D1 as `Action = 'Manual_Override'`, recording the original auto-allocation as `Old_Value` and the manual selection as `New_Value`.

### 3.4 BOM Explosion Scope

On entering `DRAFT`, the system queries A2 for active records (`Is_Active = TRUE`) under `Target_FG_Code`. It performs a **recursive walk** — FG → SFG → RM — multiplying `Qty_Per_Set` at each level until only `RM` leaf nodes remain. All RM quantities are aggregated into C1.5.

**Phase 1 SFG rule:** SFG on-hand inventory is tracked (it can be received, issued, and reported) but is **not consumed** by a parent WO. The explosion always walks through SFGs to their RM components. SFG consumption by parent WOs is deferred.

### 3.5 Receiving & Damage

**Short shipment** (ordered 100, received 60): Record `PO_RECEIPT` for 60 only. No `ADJ_OUT`. The quantity difference is a procurement matter and is tracked at the PO level, outside this system's scope.

**Supplier damage on arrival** (received 100, but 10 are damaged upon inspection): Record `PO_RECEIPT` for 100, then a separate `ADJ_OUT` for 10 with `Reason_Code = 'Supplier_Damage'`. Physical stock becomes 90, with a full audit trail.

**Cycle count adjustments:** Use `ADJ_IN` or `ADJ_OUT` with `Reason_Code = 'Cycle_Count'`. A `Reason_Code` is mandatory for all adjustment types.

### 3.6 LOT Status Management

LOT status controls eligibility for new reservations. Changes must be logged in D1.

| Transition | Who | Effect on Existing Reservations |
| :--- | :--- | :--- |
| `Active` → `Hold` | QC / Supervisor | Reservations on the LOT remain. System raises an alert on the linked WOs. |
| `Hold` → `Active` | QC / Supervisor | Reservations resume normally. |
| `Active` / `Hold` → `Quarantined` | Supervisor only | Reservations on the LOT are flagged. User must manually reassign or cancel the WO. |
| `Quarantined` → `Hold` | Supervisor only | Intermediate step before releasing back to Active. |

`Hold` and `Quarantined` LOTs are excluded from auto-allocation (§3.2) and cannot be selected in manual override (§3.3).

---

## 4. Work Order Lifecycle & State Machine

### 4.1 Allowed Status Transitions

```
DRAFT → RESERVED → IN_PRODUCTION → COMPLETED
DRAFT | RESERVED | IN_PRODUCTION → CANCELLED
CANCELLED → DRAFT  (Reopen)
```

Skipping `IN_PRODUCTION` (going `RESERVED → COMPLETED` directly) is **not permitted**. The `IN_PRODUCTION` step is required for actuals to be entered.

### 4.2 Lifecycle Mapping

**1. DRAFT — WO Creation & BOM Explosion**
- User creates WO: inputs `WO_Number`, `Target_FG_Code`, `Target_Qty`.
- System runs recursive BOM explosion → populates C1.5.
- No reservations exist yet. No ledger entries written.

**2. RESERVED — Auto-Allocation**
- System runs FEFO+FIFO allocation against available stock.
- If any material has insufficient available stock, the transition is **blocked** and the shortage is flagged per material in red. WO stays in `DRAFT`.
- On success, C2 rows are created with `Reserved_Qty`. Available Stock drops immediately (logical).
- D1 entry: `Action = 'Status_Change'`, `DRAFT → RESERVED`.

**3. IN_PRODUCTION — Active Production**
- Physical production begins. C2 reservations continue to hold stock.
- Users may enter `Actual_Used_Qty` and `Damage_Qty` on C2 lines incrementally.
- If actual consumption will exceed reserved quantity, the user must increase the reservation before the system allows completion (see §4.2 step 4 — Over-Usage Guard).

**4. COMPLETED — Closing & Ledger Write**

The following guard runs first on every C2 line:

> **Completion Guard:** Blocked if `(Actual_Used_Qty + Damage_Qty) > Reserved_Qty` on any line.  
> To resolve: user increases `Reserved_Qty` (via auto-allocation delta or manual override). Each increase is logged to D1 as `Action = 'Qty_Edit'`.

When all lines pass the guard, the system writes to B2 in this exact order per C2 line:

| Ledger Write | Type | Qty | Condition |
| :--- | :--- | :--- | :--- |
| RM consumed | `WO_ISSUE` | `Actual_Used_Qty` (negative) | Always |
| Line waste | `ADJ_OUT` | `Damage_Qty` (negative) | If `Damage_Qty > 0` |
| Return unused | `RETURN_TO_STOCK` | `Reserved_Qty − Actual_Used_Qty − Damage_Qty` (positive) | If remainder > 0 |

Then for the WO header:

* System auto-creates a B1 LOT for `Target_FG_Code`:
  - `Supplier_LOT_Number` = `WO_Number`
  - `Supplier_Name` = NULL (internal)
  - `MFG_Date` = completion date
  - `EXP_Date` = `MFG_Date + Shelf_Life_Days` if defined on A1; else NULL
  - `Status` = `Active`
* Writes `WO_RECEIPT` to B2: qty = `Actual_Produced_Qty`, linked to the new FG LOT.
* All C2 `Reserved_Qty` values are set to 0.

**5. CANCELLED — Release**
- All C2 `Reserved_Qty` values drop to 0. Stock becomes available to other WOs immediately.
- C1.5 requirements are preserved in the database for reference.
- D1 entry: `Action = 'Status_Change'`, `IN_PRODUCTION → CANCELLED` (or whichever prior status).

**6. CANCELLED → DRAFT — Reopen**
- WO returns to `DRAFT`. C1.5 is preserved but C2 rows are cleared.
- User must trigger the Reservation step again, which re-runs FEFO+FIFO against **current** available stock (previously held stock may have been claimed by another WO).
- D1 entry: `Action = 'Reopen'`.

---

## 5. Alerts & Reports

### 5.1 Real-Time Alerts

* **Low Stock:** `Available Stock < Min_Stock_Level` for any `Item_Code`. Alert fires per item, not per LOT.
* **Near Expiry:** `EXP_Date − Today ≤ Expiry_Threshold_Days` (item-level from A1; falls back to `NEAR_EXPIRY_DAYS_DEFAULT` from E1 if NULL). Only fires on LOTs with `Status = 'Active'`.
* **Reserved LOT on Hold:** Fires when a LOT linked to an active WO reservation is changed to `Hold` or `Quarantined`. Alert targets the WO owner.
* **Over-Reservation Blocked:** Fires when a WO cannot transition to `RESERVED` due to insufficient stock. Lists each short material and quantity.

### 5.2 Required Reports

**Stock On Hand**
> Item Code | Description | Item Type | LOT | Supplier | MFG Date | EXP Date | LOT Status | Physical Qty | Reserved Qty | Available Qty

**Movement Ledger**
> Timestamp | Transaction Type | Item Code | LOT | Doc Ref | Qty Changed | Reason Code | User

**WO Summary**
> WO Number | FG Code | Target Qty | Actual Produced Qty | WO Status | Created Date | Completed Date

**Shortage / Damage**
> Date | WO/PO Ref | Item Code | Supplier | Supplier Damage Qty | Line Waste Qty

**Audit Trail**
> Timestamp | WO Number | User | Action | Old Value | New Value
