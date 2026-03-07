
```mermaid
graph TD
    A[生成Token接口] --> B[读取限流配置文件]
    B --> C[查询任务池正运行任务]
    C --> D{到达限流配置?}
    D -->|是| E[生成失败]
    D -->|否| F[根据ClientId生成Token]
```