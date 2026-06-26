# System Requirements Specification (SRS): Production & Inventory Management v2.0

## 1. System Overview
A comprehensive system to manage raw material (RM) inventory, integrate with production Work Orders (WO), enforce FIFO by default, and provide real-time stock visibility. This document outlines the normalized database schema, business logic, and operational edge cases.

---

## 2. Normalized Database Schema

### 2.1 Item Master & BOM (Static Data)
**Table A1: Item Master**
* `Item_Code` (PK): Unique identifier.
* `Description`, `Unit`, `Min_Stock_Level`.
* `Item_Type`: FG (Finished Good), SFG (Semi-Finished), RM (Raw Material).

**Table A2: BOM Ledger (Relational)**
* `BOM_ID` (PK), `Parent_Item_Code` (FK to A1), `Component_Item_Code` (FK to A1).
* `Qty_Per_Set`, `BOM_Version`, `Is_Active` (Boolean).

### 2.2 LOT Management (Traceability)
**Table B1: LOT Master**
*To prevent data duplication upon re-receiving the same lot, MFG and EXP dates belong to the LOT, not the transaction.*
* `LOT_Internal_ID` (PK): System-generated unique ID.
* `Item_Code` (FK to A1).
* `Supplier_LOT_Number`: Text (Uniqueness rule: A `Supplier_LOT_Number` is only unique *per* `Item_Code`).
* `MFG_Date`, `EXP_Date`.
* `Status`: Active, Hold, Quarantined.

### 2.3 Inventory Ledger (Movements)
**Table B2: Stock Ledger**
* `Transaction_ID` (PK).
* `Timestamp`, `User_ID`.
* `Transaction_Type`: `PO_RECEIPT`, `WO_ISSUE`, `ADJ_IN` (Cycle Count up), `ADJ_OUT` (Cycle Count down/Scrap), `RETURN_TO_STOCK`.
* `LOT_Internal_ID` (FK to B1).
* `Qty_Changed`: Positive for IN/ADJ_IN, Negative for OUT/ADJ_OUT.
* `Doc_Ref`: PO Number, WO Number, or Cycle Count ID.

### 2.4 Work Order Tracking
**Table C1: WO Header**
* `WO_Number` (PK), `Target_FG_Code`, `Target_Qty`.
* `WO_Status`: `DRAFT`, `RESERVED`, `IN_PRODUCTION`, `COMPLETED`, `CANCELLED`.

**Table C2: WO Material Lines (Reservations & Actuals)**
* `Line_ID` (PK), `WO_Number` (FK).
* `Required_Item_Code`, `LOT_Internal_ID` (FK).
* `Total_Needed_Qty`, `Reserved_Qty`, `Actual_Used_Qty`, `Damage_Qty`.

---

## 3. Core Business Logic & Formulas

### 3.1 Stock Calculations (Physical vs. Logical)
* **Stock On Hand (Physical):** `SUM(Qty_Changed)` from Table B2 grouped by LOT/Item.
* **Available Stock (Logical):** `Stock On Hand` MINUS `SUM(Reserved_Qty)` from Active WOs in Table C2.

### 3.2 Inventory Control Workflows
* **PO Receiving:** User inputs PO number, selects `Item_Code`, inputs `Supplier_LOT`. System checks Table B1; if LOT exists, it links it. If not, prompts for `MFG/EXP_Dates`, creates LOT in B1, and writes `PO_RECEIPT` to B2.
* **Cycle Counts & Adjustments:** Handled via `ADJ_IN` and `ADJ_OUT` transactions in Table B2. Requires a "Reason Code" (e.g., Audited, Lost, Expired).

### 3.3 System Constraints & Protections
* **Negative Stock Prevention:** The system **strictly prohibits** any `WO_ISSUE` or `ADJ_OUT` transaction in Table B2 that would result in a LOT's `Stock On Hand` dropping below 0.
* **Over-Reservation Lock:** A WO cannot be shifted to `RESERVED` status if `Available Stock` is insufficient. The UI will flag the missing quantity in RED.

---

## 4. Work Order Lifecycle & State Machine
1. **Creation (DRAFT):** WO created, BOM exploded. No reservations made yet.
2. **Reservation (RESERVED):** Auto-FIFO algorithm runs. It sorts Table B1 by `EXP_Date` Ascending. It reserves `Available Stock` logically in Table C2.
3. **Cancellation (CANCELLED):** If a WO is cancelled, all `Reserved_Qty` values in Table C2 are reset to 0. The stock immediately becomes "Available" for other WOs.
4. **Reopening:** If a cancelled WO is reopened, it returns to `DRAFT` and must undergo the Reservation step again (as its previously held stock may have been claimed by another WO).
5. **Confirmation (COMPLETED):** Production finishes. Actual usage is written as `WO_ISSUE` (negative Qty) to Table B2. Logical reservations in C2 drop to 0.
