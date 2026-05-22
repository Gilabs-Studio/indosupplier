# K6 Skill: Relationship Diagram (Mermaid)

```mermaid
graph TD
  Sales -->|DeliveryOrder| Inventory
  Sales -->|Invoice/Payment| Finance
  Purchase -->|GoodsReceipt| Inventory
  Finance -->|Payroll/Expense| HRD
```

- Update this diagram as new relationships are added.
- Place in documentation or README as needed.
