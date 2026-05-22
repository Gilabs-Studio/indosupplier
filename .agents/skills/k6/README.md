# K6 Skill for GIMS ERP

This skill enables rapid, standardized, and well-documented k6 transaction workflow testing for GIMS ERP modules.

## Key Capabilities
- **Workflow Implementation:** Scaffold and extend k6 scripts per module, following best practices.
- **Relationship Modeling:** Document and visualize inter-module dependencies and test coverage.
- **Gill Me Mode:** Interactive, agent-driven workflow inquiry with recommendations.
- **Documentation Automation:** Maintain clean-architecture-aligned docs for each workflow/module.

## How to Use
1. **New Workflow:**
   - Trigger: `k6 skill: new workflow <module> <workflow-name>`
   - Agent will ask detailed questions (Gill Me) and scaffold code/docs.
2. **Update Docs:**
   - Trigger: `k6 skill: update docs <module> <workflow>`
   - Agent syncs documentation with latest implementation.
3. **Show Matrix:**
   - Trigger: `k6 skill: show matrix`
   - Agent updates and visualizes transaction coverage and relationships.

## Clean Architecture Alignment
- Docs and code are organized by use-cases, infrastructure, and architecture layers.
- Each workflow/module has a dedicated doc in `k6/docs/use-cases/modules/<module>/`.

## Maintenance
- Update this skill as new modules/workflows are added.
- Keep code templates and docs in sync with backend contracts.

---
See SKILL.md for full specification.
