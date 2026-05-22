# K6 Skill: Workflow Relationships

## Purpose
Document and visualize inter-module dependencies for transaction workflows.

## Example Relationships
- **Sales → Inventory**: DeliveryOrder step requires inventory batch selection
- **Sales → Finance**: Invoice and Payment steps post to Finance
- **Purchase → Inventory**: Goods receipt updates inventory
- **Finance → HRD**: Payroll and expense posting

## How to Use
- Update the coverage matrix with relationship notes
- Use diagrams or tables to visualize dependencies
- Document integration points in each module's workflow doc

---
*Keep this file updated as new relationships are discovered or implemented.*
