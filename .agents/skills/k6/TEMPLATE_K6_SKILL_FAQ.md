# K6 Skill FAQ

**Q: What is the purpose of this skill?**
A: To standardize, accelerate, and document k6 transaction workflow testing for GIMS ERP modules, following best practices and clean architecture.

**Q: How do I start a new workflow test?**
A: Use the command `k6 skill: new workflow <module> <workflow-name>`. The agent will guide you through Gill Me mode and generate code/docs.

**Q: How do I update documentation?**
A: Use `k6 skill: update docs <module> <workflow>`. The agent will sync docs with the latest implementation.

**Q: How do I see which workflows are covered?**
A: Use `k6 skill: show matrix`. The agent will update and display the transaction coverage matrix.

**Q: What if my workflow has dependencies on other modules?**
A: The skill supports modeling and documenting inter-module relationships. Use Gill Me mode to clarify dependencies.

**Q: Where are the templates?**
A: All templates are in `.agents/skills/k6/` (see TEMPLATE_* files).

**Q: How do I keep everything up to date?**
A: Always update documentation and the coverage matrix after implementing or changing a workflow.

---
*Expand this FAQ as new questions arise.*
