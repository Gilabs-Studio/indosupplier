# K6 Skill Structure

## Directory Layout

```
.agents/skills/k6/
├── SKILL.md                  # Main skill definition/spec
├── README.md                 # Summary and usage
├── TEMPLATE_README_USER.md   # User guide
├── TEMPLATE_K6_SKILL_ACTIONS.md # Supported actions/commands
├── TEMPLATE_WORKFLOW_GILLME.md  # Gill Me session template
├── TEMPLATE_GILLME_QUESTIONS.md # Gill Me question set
├── TEMPLATE_MODULE_DOC.md    # Per-module workflow doc template
├── TEMPLATE_COVERAGE_MATRIX.md  # Coverage matrix template
├── TEMPLATE_K6_SCRIPT.js     # K6 script template
```

## Usage Flow
1. User triggers a skill action (e.g., new workflow, update docs)
2. Agent enters Gill Me mode, asks questions, gives recommendations
3. Agent scaffolds/updates code and documentation using templates
4. Agent maintains coverage matrix and links

## Maintenance
- Update templates and skill files as new needs arise
- Keep in sync with backend and business process changes

---
*This structure ensures maintainability, extensibility, and best practice alignment for k6 transaction testing in GIMS ERP.*
