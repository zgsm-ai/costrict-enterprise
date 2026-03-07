---
description: "构建新的VibePlus变更并进行严格验证。"
argument-hint: 功能描述或请求
---
<!-- VIBEPLUS:START -->
**护栏原则**
- 优先采用直接、最小化的实现方式，仅在明确需要或被要求时添加复杂性。
- 保持变更范围是紧密围绕用户预期结果展开的。
- VibePlus约束或最佳实践，请一定要参考`vibeplus/AGENTS.md`。
- 在编辑文件前识别任何模糊或不明确的细节，并提出必要的后续问题。

**步骤**
1. 结合`vibeplus/project.md`，并检查相关代码或文档（例如通过`rg`/`ls`），使提案符合当前规则；注意所有需要澄清的空白模块。
2. 选择一个唯一的动词引导的`change-id`，并在`vibeplus/changes/<id>/`下构建`proposal.md`、`tasks.md`和`design.md`（需要时）。
3. 将变更映射到具体能力或需求，将多范围工作分解为具有清晰关系和顺序的不同规格增量。
4. 当解决方案跨越多个系统、引入新模式或在提交规格前需要权衡讨论时，在`design.md`中捕获架构推理。
5. 在`vibeplus/changes/<id>/specs/<capability>/spec.md`（每个能力一个文件夹）中起草规格增量，使用`## ADDED|MODIFIED|REMOVED Requirements`，每个需求至少包含一个`#### Scenario:`，并在相关时交叉引用相关能力。
6. 将`tasks.md`起草为有序的小型可验证工作项目列表，这些项目提供用户可见的进度，包括验证（测试、工具），并突出依赖项或可并行的工作。
<!-- VIBEPLUS:END -->
