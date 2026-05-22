# K6 Skill Developer Guide

## Overview
This guide is for developers maintaining or extending the k6 skill for GIMS ERP.

## Structure
- All templates and docs are in `.agents/skills/k6/`
- Main entry: `SKILL.md`
- See `TEMPLATE_K6_SKILL_STRUCTURE.md` for file layout

## Adding a New Workflow or Module
1. Update templates as needed
2. Add new module doc in `k6/docs/use-cases/modules/<module>/`
3. Update the coverage matrix
4. Add/modify k6 scripts as required
5. Document changes in the changelog

## Testing
- Review generated scripts and docs for accuracy
- Run k6 tests to validate new workflows

## Contribution
- See `TEMPLATE_K6_SKILL_CONTRIBUTING.md`

---
*Keep this guide updated as the skill evolves.*
