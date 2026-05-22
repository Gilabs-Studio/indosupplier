# K6 Skill Example: Sales Revenue Cycle

## 1. Trigger New Workflow
```
k6 skill: new workflow sales revenue-cycle
```

## 2. Gill Me Session (Sample)
- **Agent:** What is the business goal of this workflow?
- **User:** Automate the full Sales revenue cycle.
- **Agent:** Which documents/entities are involved?
- **User:** SalesQuotation, SalesOrder, DeliveryOrder, CustomerInvoice, SalesPayment.
- **Agent:** What data or preconditions are required?
- **User:** At least one active customer, batchable products, and available inventory.
- **Agent:** List the main steps in this workflow.
- **User:** Create SQ → Approve → Create SO → Approve → Create DO → Select Batches → Approve → Create Invoice → Approve → Create Payment → Confirm.
- **Agent:** Which API endpoints are involved?
- **User:** /sales/quotations, /sales/orders, /sales/delivery-orders, /sales/invoices, /sales/payments, etc.
- **Agent:** What are the expected success and failure scenarios?
- **User:** Success: All steps succeed. Failure: Inventory batch missing, duplicate payment, code generation race.
- **Agent:** Are there concurrency or integration points to consider?
- **User:** Yes, batch selection and payment confirmation must be atomic.

## 3. Generated Artifacts
- k6/transaction-test.js (scaffolded/updated)
- k6/docs/use-cases/modules/sales/SALES_REVENUE_CYCLE_IMPLEMENTED.md (generated/updated)
- k6/docs/use-cases/modules/MODULE_TRANSACTION_COVERAGE.md (updated)

---
*Use this example as a reference for future workflow implementations.*
