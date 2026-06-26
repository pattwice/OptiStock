# System Requirements Specification (SRS): Production & Inventory Management v2.1

## 1. System Overview
A comprehensive system to manage raw material (RM) inventory, integrate with production Work Orders (WO), enforce **FEFO (First-Expire-First-Out) followed by FIFO**, and provide real-time stock visibility. 

---

## 2. Normalized Database Schema

### 2.1 Item Master & BOM (Static Data)
**Table A1: Item Master**
* `Item_Code` (PK): Unique identifier.
* `Description`, `Unit`, `Min_Stock_Level`.
* `Item_Type`: FG (Finished Good), SFG (Semi-Finished), RM (Raw Material).

**Table A2: BOM Ledger (Relational)**
* `BOM_ID` (PK), `Parent_Item_Code` (FK -> A1), `Component_Item_Code` (FK -> A1).
* `Qty_Per_Set`, `BOM_Version`, `Is_Active` (Boolean).

### 2.2 LOT Management (Traceability)
**Table B1: LOT Master**
* `LOT_Internal_ID` (PK): System-generated unique ID.
* `Item_Code` (FK -> A1).
* `Supplier_LOT_Number`, `Supplier_Name`.
* `MFG_Date` (Required), `EXP_Date` (Nullable for non-expiring goods).
* `Status`: `Active`, `Hold`, `Quarantined`. *(Note: Re-receiving an existing LOT with different MFG/EXP dates on paperwork must trigger a system warning for manual review).*

### 2.3 Inventory Ledger (Movements)
**Table B2: Stock Ledger**
* `Transaction_ID` (PK), `Timestamp`, `User_ID`.
* `Transaction_Type`: `PO_RECEIPT`, `WO_ISSUE`, `ADJ_IN`, `ADJ_OUT`, `RETURN_TO_STOCK`.
* `LOT_Internal_ID` (FK -> B1).
* `Qty_Changed`: Positive for IN/ADJ_IN/RETURN, Negative for OUT/ADJ_OUT.
* `Doc_Ref` (WO/PO/Cycle Count), `Reason_Code` (e.g., Audited, Line_Waste, Expired).

### 2.4 Work Order Tracking & Audit
**Table C1: WO Header**
* `WO_Number` (PK), `Target_FG_Code` (FK -> A1), `Target_Qty`.
* `WO_Status`: `DRAFT`, `RESERVED`, `IN_PRODUCTION`, `COMPLETED`, `CANCELLED`.

**Table C2: WO Material Lines (Reservations & Actuals)**
* *Cardinality Rule: One row per (WO_Number, Required_Item_Code, LOT_Internal_ID).*
* `Line_ID` (PK), `WO_Number` (FK).
* `Required_Item_Code`, `LOT_Internal_ID` (FK).
* `Total_Needed_Qty` (Rolled up per item), `Reserved_Qty`, `Actual_Used_Qty`, `Damage_Qty`.

**Table D1: WO Audit Log**
* `Log_ID` (PK), `WO_Number` (FK), `User_ID`, `Timestamp`.
* `Action`: (e.g., Manual Override, Status Change, Qty Edit).
* `Old_Value`, `New_Value`.

---

## 3. Core Business Logic & Formulas

### 3.1 Stock Calculations & Constraints
* **Stock On Hand (Physical):** `SUM(Qty_Changed)` from B2.
* **Available Stock (Logical):** `Stock On Hand` MINUS `SUM(Reserved_Qty)` from Active WOs (where `WO_Status` IN (`RESERVED`, `IN_PRODUCTION`)).
* **Negative Stock Prevention:** Transaction/Row-level locking is enforced during reservation to prevent concurrent WOs from reserving the same available quantity. `WO_ISSUE` and `ADJ_OUT` cannot drop Physical Stock below 0.

### 3.2 Allocation Sort Rules (FEFO + FIFO)
When auto-reserving LOTs, the system queries `Available Stock > 0` and `Status = 'Active'` in Table B1.
1. **Primary Sort (FEFO):** `EXP_Date` Ascending (NULLs sorted last).
2. **Secondary Sort (FIFO):** `MFG_Date` Ascending.

### 3.3 Manual LOT Override
Users with permissions can bypass auto-allocation, manually selecting specific active LOTs. This action is recorded in the **WO Audit Log (D1)**.

### 3.4 BOM Explosion Scope
When exploding a BOM, the system queries A2 where `Is_Active = TRUE`. For this phase, it utilizes a flat explosion directly to the `RM` level to calculate `Total_Needed_Qty`. 

---

## 4. Work Order Lifecycle & Ledger Mapping

1. **DRAFT:** WO created. BOM exploded.
2. **RESERVED:** Auto-allocation (FEFO+FIFO) applies. `Available Stock` drops.
3. **IN_PRODUCTION:** Physical work begins. Reservations still hold stock. Line issues/confirmations are active.
4. **CANCELLED:** `Reserved_Qty` drops to 0. Stock becomes available. Reopening forces a return to DRAFT.
5. **COMPLETED (Actual vs. Reserved Mapping):**
   * *Match:* If Used == Reserved -> Writes `WO_ISSUE` to B2.
   * *Under-Usage:* If Used < Reserved -> Writes `WO_ISSUE` for used amount. Writes `RETURN_TO_STOCK` to B2 for the remainder.
   * *Damage/Waste:* Logged on C2 for reporting, mapped to B2 as `ADJ_OUT` with `Reason_Code = 'LINE_WASTE'`. (Strict mapping prevents double-deduction).

---

## 5. Alerts & Reports

### 5.1 Real-Time Alerts
* **Low Stock:** Total Available Stock < `Min_Stock_Level`.
* **Near Expiry:** `EXP_Date` - Current Date <= User-defined threshold (e.g., 30 days).

### 5.2 Required Reports
* **Stock On Hand:** Item, LOT, MFG/EXP, Physical vs Available.
* **Movement Ledger:** Timestamp, Type, Item, Doc Ref, LOT, Qty Changed, User.
* **Shortage/Damage:** Date, WO Ref, Item, Supplier, Line Waste Qty.
