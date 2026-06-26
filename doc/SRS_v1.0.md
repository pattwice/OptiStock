# System Requirements Specification (SRS): Production & Inventory Management

## 1. System Overview
A comprehensive system to manage raw material (RM) inventory, integrate with production Work Orders (WO), enforce FIFO by default, and provide real-time stock visibility and alerts.

## 2. Database Schema & Data Mapping

### Table A: Items & BOM Master
| Field | Type | Description |
| :--- | :--- | :--- |
| Item_Code | Primary Key | Unique identifier (e.g., 12003750) |
| Description | Text | Item name (e.g., Tape-Logo 1000 Yard) |
| Unit | Text | EA, RL, Kg, etc. |
| Min_Stock_Level | Numeric | Threshold for low stock alerts |
| BOM_Mapping | Relational | Parent FG Code -> Component Item Code -> Qty Per Set |

### Table B: Stock Ledger (Transactions)
| Field | Description |
| :--- | :--- |
| Transaction_Date | Date of IN/OUT movement |
| Transaction_Type | Receiving (IN) or Issuing (OUT) |
| Item_Code | Foreign Key to Item Master |
| Doc_Ref_No | PO Number, WO Number, or SAP Number |
| LOT_Number | Supplier or Internal Batch ID |
| MFG_Date | Manufacturing Date |
| EXP_Date | Expiration Date |
| Qty_In | Quantity received |
| Qty_Out | Quantity issued |
| Damage_Qty | Shortages, Supplier Damage, or Production Waste |

*Note: Stock On Hand per LOT is dynamically calculated as: Sum(Qty In) - Sum(Qty Out) - Sum(Damage_Qty).*

### Table C: Work Order (WO) Tracking
| Field | Description |
| :--- | :--- |
| WO_Number | Primary Key (e.g., 69-05/06) |
| FG_Code | Finished Good to be produced |
| Target_Qty | Total production sets/cartons needed |
| Required_Material | Component pulled from BOM |
| Total_Needed_Qty | Target Qty * BOM Qty Per Set |
| Reserved_LOT | LOT allocated for production |
| Status_Flag | GREEN (Available) / RED (Insufficient) |

---

## 3. Programmatic Logic & Business Rules

### 3.1 Inventory Allocation (FIFO & Manual)
* **Auto-FIFO Algorithm:** When a WO is created, the system filters available LOTs for required materials (Stock > 0). It sorts these LOTs by `EXP_Date` in **Ascending** order. Quantities are reserved sequentially from the oldest expiring LOT until the `Total_Needed_Qty` is met.
* **Manual Override:** Users can view a dropdown of all active LOTs and their balances to manually select a batch, bypassing the Auto-FIFO lock.

### 3.2 Real-Time Alerts
* **Low Stock:** Triggered when the aggregate stock of all LOTs for an `Item_Code` falls below `Min_Stock_Level`.
* **Near Expiry:** Triggered when `EXP_Date` minus Current Date is less than or equal to the defined threshold (e.g., 30 days).

---

## 4. Production Workflow
1. **Initiation:** Production Control generates a new WO and inputs the FG Code and Target Qty.
2. **BOM Explosion:** System pulls the BOM and calculates raw material requirements.
3. **Validation:** System checks real-time stock and flags shortages in RED.
4. **Reservation:** System applies Auto-FIFO to reserve specific LOTs.
5. **Confirmation:** Production line confirms actual usage post-production, officially deducting from the Stock Ledger.
6. **Closure:** WO is closed; defects and actuals are recorded.

---

## 5. Export & Reporting Templates

### 5.1 Stock on Hand Report
* Item Code | Description | LOT | MFG | EXP | Available Qty

### 5.2 Movement Report
* Timestamp | Type (IN/OUT) | Item Code | Doc Ref | LOT | Qty Changed | User

### 5.3 Short/Damage Report
* Date | WO/PO Ref | Item Code | Supplier | Short Qty | Supplier Damage Qty | Line Waste Qty
