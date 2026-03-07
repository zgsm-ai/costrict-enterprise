---
description: "归档已部署的VibePlus变更并更新规格。"
argument-hint: change-id
---
<!-- VIBEPLUS:START -->
**护栏原则**
- 优先采用直接、最小化的实现方式，仅在明确需要或被要求时添加复杂性。
- 保持变更范围是紧密围绕用户预期结果展开的。
- VibePlus约束或最佳实践，请一定要参考`vibeplus/AGENTS.md`。

**步骤**
1. 确定要归档的变更ID：
   - 如果此提示已包含特定变更ID（例如在由斜杠命令参数填充的`<ChangeId>`块内），请使用修剪空白字符后的值。
   - 如果仍无法确定单个变更ID，请停止并告诉用户您尚无法归档任何内容。
2. 检查命令输出，确认目标规格已更新，且变更已放入`changes/archive/`中。
<!-- VIBEPLUS:END -->
