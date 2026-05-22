# K6 Transaction Testing Skill for GIMS ERP

## Purpose
A user-level skill to standardize, accelerate, and document k6-based transaction workflow testing in GIMS ERP. This skill enables:
- Rapid implementation of transaction workflows per module (Sales, Purchase, Finance, HRD, etc.)
- Explicit modeling of inter-module relationships
- "Gill Me" mode: agent-driven, detailed workflow inquiry with recommendations
- Automated, clean-architecture-aligned documentation and updates for each workflow/module

## Features

### 1. Workflow Implementation per Module
- Scaffold k6 scripts for any ERP module (Sales, Purchase, etc.)
- Enforce best practices: authentication, CSRF, dynamic data, concurrency, error handling
- Support for full transaction cycles (e.g., SalesQuotation → SalesOrder → DeliveryOrder → Invoice → Payment)
- Allow explicit modeling of dependencies (e.g., Purchase → Inventory → Finance)
- Provide ready-to-use code templates and helper functions

### 2. Workflow Relationship Modeling
- Document and visualize inter-module dependencies (e.g., Sales triggers Inventory, Finance posts)
- Maintain a matrix of transaction coverage and relationships
- Recommend test order and data setup based on dependencies

### 3. "Gill Me" Interactive Inquiry
- Agent asks detailed, context-aware questions about the intended workflow:
  - What is the business goal of this workflow?
  - What are the main entities and documents involved?
  - What are the required preconditions/data?
  - What are the expected success/failure scenarios?
  - Are there concurrency or integration points?
- Agent provides recommendations and best practices after each answer
- Supports iterative refinement: user can update answers, agent adapts recommendations

### 4. Documentation Automation (Clean Architecture)
- For each workflow/module:
  - Generate/maintain docs in k6/docs/use-cases/modules/<module>/<WORKFLOW>.md
  - Structure docs by: business context, workflow steps, API mapping, data dependencies, error handling, test coverage
  - Update module coverage matrix and index
- Ensure docs follow clean architecture layering: use-cases, infrastructure, architecture
- Auto-link new docs in k6/docs/INDEX.md and related files

## Usage
1. To implement a new workflow test:
   - Run: "k6 skill: new workflow <module> <workflow-name>"
   - Agent will enter "Gill Me" mode to collect details
   - Agent scaffolds k6 script and documentation
2. To update documentation:
   - Run: "k6 skill: update docs <module> <workflow>"
   - Agent will sync docs with latest implementation and recommendations
3. To visualize relationships:
   - Run: "k6 skill: show matrix"
   - Agent generates/updates MODULE_TRANSACTION_COVERAGE.md and diagrams

## Example Gill Me Session
- Agent: "What is the business goal of this workflow?"
- User: "Automate the full Sales revenue cycle."
- Agent: "Which documents are involved? (e.g., Quotation, Order, Delivery, Invoice, Payment)"
- ...
- Agent: "Recommended: Always use dynamic product pools and batch selection for inventory consistency."

## Maintenance
- Skill should be updated as new modules or workflows are added
- Documentation and code templates must be kept in sync with backend contracts and business rules

---

**Author:** GIMS Platform Engineering
**Last Updated:** 2026-05-12
