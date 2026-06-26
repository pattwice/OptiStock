# System Requirements Specification (SRS): Production & Inventory Management v2.2

## 1. System Overview
A comprehensive system to manage raw material (RM) inventory, integrate with production Work Orders (WO), enforce **FEFO (First-Expire-First-Out) followed by FIFO**, and provide real-time stock visibility. 

---

## 2. Normalized Database Schema

### 2.1 Item Master & BOM (Static Data)
**Table A1: Item Master**
* `Item_Code` (PK): Unique identifier.
* `Description`, `Unit`, `Min_Stock_Level`.
* `Item_Type`: FG (Finished Good), SFG (Semi-Finished), RM (Raw Material).
* `Expiry_Threshold_Days` (Nullable): Item-specific threshold for near-expiry alerts.

**Table A2: BOM Ledger (Relational)**
* *Constraint: Only one `BOM_Version` can be `Is_Active = TRUE` per `Parent_Item_Code` at any given time.*
* `BOM_ID` (PK), `Parent_Item_Code` (FK -> A1), `Component_Item_Code` (FK -> A1).
* `Qty_Per_Set`, `BOM_Version`, `Is_Active` (Boolean).

### 2.2 LOT Management (Traceability)
**Table B1: LOT Master**
* `LOT_Internal_ID` (PK): System-generated unique ID.
* `Item_Code` (FK -> A1).
* `Supplier_LOT_Number`, `Supplier_Name`.
* `MFG_Date` (Required), `EXP_Date` (Nullable for non-expiring goods).
* `Status`: `Active`, `Hold`, `Quarantined`. 

### 2.3 Inventory Ledger (Movements)
**Table B2: Stock Ledger**
* `Transaction_ID` (PK), `Timestamp`, `User_ID`.
* `Transaction_Type`: `PO_RECEIPT`, `WO_ISSUE`, `WO_RECEIPT` (FG Output), `ADJ_IN`, `ADJ_OUT`, `RETURN_TO_STOCK`.
* `LOT_Internal_ID` (FK -> B1).
* `Qty_Changed`: Positive for IN/ADJ_IN/RETURN/WO_RECEIPT, Negative for OUT/ADJ_OUT.
* `Doc_Ref` (WO/PO/Cycle Count).
* `Reason_Code` (Nullable; e.g., Audited, Line_Waste, Supplier_Damage, Expired).

### 2.4 Work Order Tracking & Audit
**Table C1: WO Header**
* `WO_Number` (PK), `Target_FG_Code` (FK -> A1), `Target_Qty`.
* `WO_Status`: `DRAFT`, `RESERVED`, `IN_PRODUCTION`, `COMPLETED`, `CANCELLED`.

**Table C1.5: WO Material Requirements (Aggregated)**
* *Cardinality: One row per (WO_Number, Required_Item_Code).*
* `Req_ID` (PK), `WO_Number` (FK).
* `Required_Item_Code` (FK -> A1).
* `Total_Needed_Qty`.

**Table C2: WO LOT Allocations (Reservations & Actuals)**
* *Cardinality: One row per (Req_ID, LOT_Internal_ID).*
* `Allocation_ID` (PK), `Req_ID` (FK -> C1.5), `LOT_Internal_ID` (FK -> B1).
* `Reserved_Qty`, `Actual_Used_Qty`, `Damage_Qty`.

**Table D1: WO Audit Log**
* `Log_ID` (PK), `WO_Number` (FK), `User_ID`, `Timestamp`.
* `Action`: (e.g., Manual Override, Status Change, Qty Edit).
* `Old_Value`, `New_Value`.

---

## 3. Core Business Logic & Workflows

### 3.1 Stock Calculations & Constraints
* **Stock On Hand (Physical):** `SUM(Qty_Changed)` from B2.
* **Available Stock (Logical):** `Stock On Hand` MINUS `SUM(Reserved_Qty)` from Active WOs (Status: `RESERVED`, `IN_PRODUCTION`).
* **Negative Stock Prevention:** System utilizes row-level locking during reservation to prevent concurrent WOs from reserving the same available quantity. `WO_ISSUE` and `ADJ_OUT` cannot drop Physical Stock below 0.

### 3.2 Allocation Sort Rules (FEFO + FIFO)
When auto-reserving LOTs, the system queries `Available Stock > 0` and `Status = 'Active'` in Table B1.
1. **Primary Sort (FEFO):** `EXP_Date` Ascending (NULLs sorted last).
2. **Secondary Sort (FIFO):** `MFG_Date` Ascending.

### 3.3 BOM Explosion Scope
When exploding a BOM, the system queries A2 where `Is_Active = TRUE`. It performs a **recursive walk** (FG -> SFG -> RM), multiplying quantities at each level down to the lowest level (`RM`), and aggregates the `Total_Needed_Qty` into Table C1.5.

### 3.4 Receiving & Damage
* **PO Receipt:** Adds `PO_RECEIPT` to B2. If shortages or supplier damage are identified upon receipt, a concurrent `ADJ_OUT` transaction with `Reason_Code = 'SUPPLIER_DAMAGE'` is logged to instantly correct physical stock.

---

## 4. Work Order Lifecycle & State Machine

### 4.1 Allowed Status Transitions
* `DRAFT` -> `RESERVED` -> `IN_PRODUCTION` -> `COMPLETED`
* `DRAFT` | `RESERVED` | `IN_PRODUCTION` -> `CANCELLED`
* `CANCELLED` -> `DRAFT` (Reopen)

### 4.2 Lifecycle Mapping
1. **DRAFT:** WO created. Recursive BOM exploded into C1.5.
2. **RESERVED:** Auto-allocation applies, splitting requirements into specific LOTs in C2. `Available Stock` drops.
3. **IN_PRODUCTION:** Physical work begins. Line issues/confirmations are active.
4. **COMPLETED:** * **Over-Usage Guard:** Completion is **blocked** if `Actual_Used_Qty > Reserved_Qty` on any C2 line. The user must manually edit the WO (while `IN_PRODUCTION`) to increase the reservation. This writes an `Action = Qty Edit` to the **Audit Log (D1)**. Only when `Actual <= Reserved` can the WO close.
   * **RM Issue:** `Actual_Used_Qty` writes as `WO_ISSUE` to B2.
   * **Under-Usage:** If `Actual_Used_Qty < Reserved_Qty`, the unused remainder writes as `RETURN_TO_STOCK` to B2.
   * **Waste:** `Damage_Qty` on C2 maps to `ADJ_OUT` with `Reason_Code = 'LINE_WASTE'`.
   * **FG Receipt:** Upon successful completion, the system automatically writes a `WO_RECEIPT` to B2 for the `Target_FG_Code` reflecting the successfully produced Finished Goods.
5. **CANCELLED:** `Reserved_Qty` drops to 0. Stock becomes available.

---

## 5. Alerts & Reports

### 5.1 Real-Time Alerts
* **Low Stock:** Total Available Stock < `Min_Stock_Level`.
* **Near Expiry:** `EXP_Date` - Current Date <= `Expiry_Threshold_Days` (fallback to global config if null).

### 5.2 Required Reports
* **Stock On Hand:** Item, LOT, MFG/EXP, Physical vs Available.
* **Movement Ledger:** Timestamp, Type, Item, Doc Ref, LOT, Qty Changed, User, Reason Code.
* **Shortage/Damage:** Date, WO/PO Ref, Item, Supplier, Line Waste vs Supplier Damage.
