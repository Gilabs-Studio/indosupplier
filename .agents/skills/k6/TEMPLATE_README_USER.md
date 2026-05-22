# K6 Skill User Guide

## What is this Skill?
A user-level skill to help you implement, document, and maintain k6 transaction workflow tests for GIMS ERP modules, following best practices and clean architecture.

## What Can You Do?
- Scaffold new workflow tests per module
- Model and visualize workflow relationships
- Use "Gill Me" mode for agent-driven inquiry and recommendations
- Automate and update documentation for each workflow/module

## How to Use
1. **Create a New Workflow Test**
   - Trigger: `k6 skill: new workflow <module> <workflow-name>`
   - Agent will ask you detailed questions (Gill Me) and generate code/docs
2. **Update Documentation**
   - Trigger: `k6 skill: update docs <module> <workflow>`
   - Agent will sync docs with latest implementation
3. **Show Coverage Matrix**
   - Trigger: `k6 skill: show matrix`
   - Agent updates and visualizes transaction coverage and relationships

## Best Practices
- Always answer Gill Me questions as completely as possible
- Review generated docs and scripts for accuracy
- Keep documentation and code in sync with backend changes

## Where to Find Templates
- Gill Me session: `TEMPLATE_WORKFLOW_GILLME.md`
- Module doc: `TEMPLATE_MODULE_DOC.md`
- Coverage matrix: `TEMPLATE_COVERAGE_MATRIX.md`
- K6 script: `TEMPLATE_K6_SCRIPT.js`

---
*For questions or improvements, update the skill files or contact the GIMS Platform Engineering team.*
