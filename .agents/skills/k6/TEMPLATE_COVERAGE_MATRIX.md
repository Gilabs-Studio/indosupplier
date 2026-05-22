# K6 Transaction Coverage Matrix

| Module   | Workflow(s) Implemented         | Relationships / Dependencies         | Status      |
|----------|---------------------------------|--------------------------------------|-------------|
| Sales    | Revenue Cycle (Quotation→...)   | Inventory, Finance                   | Complete    |
| Purchase | [Planned: Procurement Cycle]    | Inventory, Finance                   | Planned     |
| Finance  | [Planned: Payment, Posting]     | Sales, Purchase, HRD                 | Planned     |
| HRD      | [Planned: Payroll, Attendance]  | Finance                              | Planned     |

- Update this matrix as new workflows are implemented.
- Add new modules/workflows as needed.
- Status: Complete / In Progress / Planned / Blocked

---
*Place this matrix in k6/docs/use-cases/modules/MODULE_TRANSACTION_COVERAGE.md and update as part of the skill.*
