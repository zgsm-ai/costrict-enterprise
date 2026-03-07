# VibePlus 指南

使用 VibePlus 进行规范驱动开发的 AI 编码助手的说明。

## 三阶段工作流程

### 阶段 1：创建变更
在以下情况下需要创建提案：
- 添加功能或特性
- 进行破坏性变更（API、架构）
- 更改架构或模式
- 优化性能（改变行为）
- 更新安全模式

触发器（示例）：
- "帮我创建一个变更提案"
- "帮我规划一个变更"
- "帮我创建一个提案"
- "我想创建一个规范提案"
- "我想创建一个规范"

跳过提案的情况：
- 错误修复（恢复预期行为）
- 拼写错误、格式、注释
- 依赖更新（非破坏性）
- 配置更改
- 现有行为的测试

**工作流程**
1. 查看 `vibeplus/project.md`, 以了解当前上下文。
2. 选择一个唯一的动词引导的 `change-id`，并在 `vibeplus/changes/<id>/` 下构建 `proposal.md`, `tasks.md`, 可选的 `design.md` 和规范增量。
3. 使用 `## ADDED|MODIFIED|REMOVED Requirements` 起草规范增量，每个需求至少包含一个 `#### Scenario:`。

### 阶段 2：实施变更
将这些步骤作为 TODO 项跟踪，并逐一完成。
1. **阅读 proposal.md** - 了解要构建的内容
2. **阅读 design.md**（如果存在）- 审查技术决策
3. **阅读 tasks.md** - 获取实施清单
4. **按顺序实施任务** - 按顺序完成
5. **确认完成** - 在更新状态前确保 `tasks.md` 中的每个项目都已完成
6. **更新清单** - 所有工作完成后，将每个任务设置为 `- [x]`，使清单反映实际情况
7. **批准关口** - 在提案被审查和批准前不要开始实施

### 阶段 3：归档变更
部署后，创建单独的 PR 来：
- 将 `changes/[name]/` 移动到 `changes/archive/YYYY-MM-DD-[name]/`

## 在任何任务之前

**上下文检查清单：**
- [ ] 阅读 `specs/[capability]/spec.md` 中的相关规范
- [ ] 检查 `changes/` 中待处理的变更是否有冲突
- [ ] 阅读 `vibeplus/project.md` 了解约定

**创建规范之前：**
- 始终检查功能是否已存在
- 优先修改现有规范而不是创建重复项
- 如果请求不明确，在构建脚手架前询问 1-2 个澄清问题

## 目录结构

```
vibeplus/
├── project.md              # 项目约定
├── specs/                  # 当前事实 - 已构建的内容
│   └── [capability]/       # 单一专注功能
│       ├── spec.md         # 需求和场景
│       └── design.md       # 技术模式
├── changes/                # 提案 - 应该变更的内容
│   ├── [change-name]/
│   │   ├── proposal.md     # 原因、内容、影响
│   │   ├── tasks.md        # 实施清单
│   │   ├── design.md       # 技术决策（可选；参见标准）
│   │   └── specs/          # 增量变更
│   │       └── [capability]/
│   │           └── spec.md # ADDED/MODIFIED/REMOVED
│   └── archive/            # 已完成的变更
```

## 创建变更提案

### 决策树

```
新请求？
├─ 恢复规范行为的错误修复？→ 直接修复
├─ 拼写错误/格式/注释？→ 直接修复
├─ 新功能/能力？→ 创建提案
├─ 破坏性变更？→ 创建提案
├─ 架构变更？→ 创建提案
└─ 不明确？→ 创建提案（更安全）
```

### 提案结构

1. **创建目录：** `changes/[change-id]/`（短横线命名法，动词引导，唯一）

2. **编写 proposal.md:**
```markdown
# 变更：[变更的简要描述]

## 原因
[关于问题/机会的 1-2 句话]

## 变更内容
- [变更的要点列表]
- [用 **BREAKING** 标记破坏性变更]

## 影响
- 受影响的规范：[列出功能]
- 受影响的代码：[关键文件/系统]
例如：
- **受影响的规范**：招聘管理 (Recruitment Management)
- **受影响的代码**：
    - `hrms/static/api/menus.json`: 新增菜单项。
    - `hrms/views/`: 新增页面文件。
    - `hrms/service/candidate.go` & `hrms/handler/candidate.go`: 新增或更新查询逻辑。
```

3. **创建规范增量：** `specs/[capability]/spec.md`
```markdown
## ADDED Requirements
### Requirement: 新功能
系统应提供...

#### Scenario: 成功案例
- **WHEN** 用户执行操作
- **THEN** 预期结果

## MODIFIED Requirements
### Requirement: 现有功能
[完整的修改后的需求]

## REMOVED Requirements
### Requirement: 旧功能
**原因**：[为什么移除]
**迁移**：[如何处理]
```
如果多个功能受到影响，请在 `changes/[change-id]/specs/<capability>/spec.md` 下创建多个增量文件——每个功能一个。

4. **创建 tasks.md:**
```markdown
## 1. 实施
- [ ] 1.1 后端：在 `service/candidate.go` 中实现联合查询逻辑，支持按候选人姓名 (`name`) 和面试官姓名 (`staff_name`) 筛选。需处理 `Candidate` 表与 `Staff` 表的关联。
- [ ] 1.2 前端：更新 `views/interview_record_manage.html` 中的 JS 逻辑，适配新的搜索 API 和评价更新 API。
- [ ] 1.3 配置：更新 `hrms/static/api/menus.json`，在“招聘管理”下添加“面试记录”菜单项。
```

5. **在需要时创建 design.md:**
如果满足以下任何条件，请创建 `design.md`；否则省略它：
- 横切变更（多个服务/模块）或新架构模式
- 新的外部依赖或重要的数据模型变更
- 安全性、性能或迁移复杂性
- 在编码前受益于技术决策的模糊情况

最小的 `design.md` 框架：
```markdown
## 上下文
[背景、约束、利益相关者]

## 目标 / 非目标
- 目标：[...]
- 非目标：[...]

## 决策
- 决策：[内容和原因]
- 考虑的替代方案：[选项 + 理由]

## 风险 / 权衡
- [风险] → 缓解措施

## 迁移计划
[步骤、回滚]

## 开放问题
- [...]
```

## 规范文件格式

### 关键：场景格式

**正确**（使用 #### 标题）：
```markdown
#### Scenario: 用户登录成功
- **WHEN** 提供有效凭据
- **THEN** 返回 JWT 令牌
```

**错误**（不要使用项目符号或粗体）：
```markdown
- **Scenario: 用户登录**  ❌
**Scenario**: 用户登录     ❌
### Scenario: 用户登录      ❌
```

每个需求必须至少有一个场景。

### 需求措辞
- 对规范性要求使用 SHALL/MUST（除非有意非规范性，否则避免使用 should/may）

### 增量操作

- `## ADDED Requirements` - 新功能
- `## MODIFIED Requirements` - 变更的行为
- `## REMOVED Requirements` - 弃用的功能
- `## RENAMED Requirements` - 名称变更

标题使用 `trim(header)` 匹配 - 忽略空白字符。

#### 何时使用 ADDED vs MODIFIED
- ADDED：引入可以独立作为需求的新功能或子功能。当变更正交（例如添加"斜杠命令配置"）而不是改变现有需求的语义时，优先使用 ADDED。
- MODIFIED：更改现有需求的行为、范围或验收标准。始终粘贴完整的更新需求内容（标题 + 所有场景）。归档器将用您在此处提供的内容替换整个需求；部分增量将丢失之前的详细信息。
- RENAMED：仅在名称更改时使用。如果您还更改行为，请使用 RENAMED（名称）加上 MODIFIED（内容）引用新名称。

常见陷阱：使用 MODIFIED 添加新关注点而不包含之前的文本。这会导致在归档时丢失详细信息。如果您没有明确更改现有需求，请在 ADDED 下添加新需求。

正确编写 MODIFIED 需求：
1) 在 `vibeplus/specs/<capability>/spec.md` 中找到现有需求。
2) 复制整个需求块（从 `### Requirement: ...` 到其场景）。
3) 将其粘贴到 `## MODIFIED Requirements` 下并编辑以反映新行为。
4) 确保标题文本完全匹配（不区分空白字符）并至少保留一个 `#### Scenario:`。

RENAMED 示例：
```markdown
## RENAMED Requirements
- FROM: `### Requirement: 登录`
- TO: `### Requirement: 用户认证`
```

## 故障排除

### 常见错误

**"变更必须至少有一个增量"**
- 检查 `changes/[name]/specs/` 存在且包含 .md 文件
- 验证文件具有操作前缀（## ADDED Requirements）

**"需求必须至少有一个场景"**
- 检查场景使用 `#### Scenario:` 格式（4 个井号）
- 不要对场景标题使用项目符号或粗体

**静默场景解析失败**
- 需要精确格式：`#### Scenario: 名称`

## 快速路径脚本

```bash
# 1) 选择变更 id 并构建脚手架
CHANGE=add-two-factor-auth
mkdir -p vibeplus/changes/$CHANGE/{specs/auth}
printf "## 原因\n...\n\n## 变更内容\n- ...\n\n## 影响\n- ...\n" > vibeplus/changes/$CHANGE/proposal.md
printf "## 1. 实施\n- [ ] 1.1 ...\n" > vibeplus/changes/$CHANGE/tasks.md

# 2) 添加增量（示例）
cat > vibeplus/changes/$CHANGE/specs/auth/spec.md << 'EOF'
## ADDED Requirements
### Requirement: 双因素认证
用户在登录期间必须提供第二个因素。

#### Scenario: 需要一次性密码
- **WHEN** 提供有效凭据
- **THEN** 需要一次性密码挑战
EOF
```

## 多功能示例

```
vibeplus/changes/add-2fa-notify/
├── proposal.md
├── tasks.md
└── specs/
    ├── auth/
    │   └── spec.md   # ADDED: 双因素认证
    └── notifications/
        └── spec.md   # ADDED: 一次性密码邮件通知
```

auth/spec.md
```markdown
## ADDED Requirements
### Requirement: 双因素认证
...
```

notifications/spec.md
```markdown
## ADDED Requirements
### Requirement: 一次性密码邮件通知
...
```

## 最佳实践

### 简单性优先
- 默认少于 100 行新代码
- 单文件实现，直到证明不足
- 避免没有明确理由的框架
- 选择无聊但经过验证的模式

### 复杂性触发器
仅在以下情况下添加复杂性：
- 性能数据显示当前解决方案太慢
- 具体的规模要求（>1000 用户，>100MB 数据）
- 多个经过验证需要抽象的用例

### 清晰引用
- 使用 `file.ts:42` 格式表示代码位置
- 引用规范为 `specs/auth/spec.md`
- 链接相关变更和 PR

### 功能命名
- 使用动词-名词：`user-auth`, `payment-capture`
- 每个功能单一目的
- 10 分钟可理解规则
- 如果描述需要 "AND"，则拆分

### 变更 ID 命名
- 使用短横线命名法，简短且描述性：`add-two-factor-auth`
- 优先使用动词引导前缀：`add-`, `update-`, `remove-`, `refactor-`
- 确保唯一性；如果已被占用，附加 `-2`, `-3` 等

## 工具选择指南

| 任务 | 工具 | 原因 |
|------|------|-----|
| 按模式查找文件 | Glob | 快速模式匹配 |
| 搜索代码内容 | Grep | 优化的正则表达式搜索 |
| 读取特定文件 | Read | 直接文件访问 |
| 探索未知范围 | Task | 多步调查 |

## 错误恢复

### 变更冲突
1. 检查是否有重叠的规范
2. 与变更所有者协调
3. 考虑合并提案

### 验证失败
1. 验证规范文件格式
2. 确保场景格式正确

### 缺少上下文
1. 首先阅读 project.md
2. 检查相关规范
3. 查看最近的归档
4. 询问澄清问题

## 快速参考

### 阶段指示器
- `changes/` - 已提议，尚未构建
- `specs/` - 已构建和部署
- `archive/` - 已完成的变更

### 文件用途
- `proposal.md` - 提案原因和具体内容
- `tasks.md` - 实施步骤
- `design.md` - 技术决策
- `spec.md` - 需求和行为

记住：规范是事实。变更是提案。保持它们同步。
