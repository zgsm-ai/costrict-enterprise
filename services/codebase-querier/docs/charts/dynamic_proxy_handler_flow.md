# DynamicProxyHandler 流程图

## 动态代理处理器处理流程

```mermaid
graph TD
    A[HTTP请求到达] --> B[DynamicProxyHandler.ServeHTTP]
    B --> C[记录请求日志]
    C --> D{请求方法是否为GET}
    D -->|是| E[从URL参数获取clientId]
    D -->|否| F[读取请求体内容]
    
    F --> G[限制请求体大小<br/>最大10MB]
    G --> H{请求体是否超限}
    H -->|超限| I[返回413错误]
    H -->|正常| J[解析JSON获取clientId]
    J --> K[重置请求体]
    
    E --> L[调用PortManager.GetPortFromHeaders]
    K --> L
    L --> M[传递方法、请求头、参数、请求体]
    M --> N[获取端口信息]
    N --> O{获取是否成功}
    O -->|失败| P[返回400错误]
    O -->|成功| Q[记录端口信息]
    
    Q --> R[调用PortManager.BuildTargetURL]
    R --> S[构建目标URL]
    S --> T[创建HTTP客户端]
    T --> U[构建目标请求]
    U --> V[复制请求头<br/>跳过clientId和appName]
    V --> W[复制查询参数]
    W --> X[发送请求到目标服务]
    X --> Y{请求是否成功}
    Y -->|失败| Z[返回502错误]
    Y -->|成功| AA[复制响应头]
    AA --> BB[设置响应状态码]
    BB --> CC[复制响应体<br/>32KB缓冲区]
    CC --> DD[记录成功日志]
    DD --> EE[完成请求处理]
    
    I --> EE
    P --> EE
    Z --> EE

    style A fill:#e1f5fe
    style EE fill:#c8e6c9
    style I fill:#ffcdd2
    style P fill:#ffcdd2
    style Z fill:#ffcdd2
    style N fill:#fff3e0
    style S fill:#fff3e0
    style X fill:#e8f5e8
    style AA fill:#e8f5e8
```

## DynamicProxyHandler.HealthCheck 流程图

```mermaid
graph TD
    A[健康检查请求] --> B[DynamicProxyHandler.HealthCheck]
    B --> C{请求方法是否为GET}
    C -->|是| D[从URL参数获取clientId]
    C -->|否| E[读取请求体内容]
    
    E --> F[限制请求体大小<br/>最大10MB]
    F --> G{请求体是否超限}
    G -->|超限| H[返回错误响应]
    G -->|正常| I[解析JSON获取clientId]
    I --> J[重置请求体]
    
    D --> K[调用PortManager.GetPortFromHeaders]
    J --> K
    K --> L[获取端口信息]
    L --> M{获取是否成功}
    M -->|失败| N[返回错误响应]
    M -->|成功| O[构建目标URL]
    
    O --> P[添加健康检查路径<br/>/health]
    P --> Q[创建HTTP客户端<br/>10秒超时]
    Q --> R[记录开始时间]
    R --> S[发送GET请求到健康检查URL]
    S --> T{请求是否成功}
    T -->|失败| U[记录错误信息]
    U --> V[调用sendHealthCheckResponse<br/>healthy=false]
    
    T -->|成功| W[检查响应状态码]
    W --> X{状态码是否为2xx}
    X -->|是| Y[计算响应时间]
    Y --> Z[调用sendHealthCheckResponse<br/>healthy=true]
    X -->|否| AA[计算响应时间]
    AA --> V
    
    Z --> AB[返回健康检查响应]
    V --> AB
    H --> AB
    N --> AB

    style A fill:#e1f5fe
    style AB fill:#c8e6c9
    style H fill:#ffcdd2
    style N fill:#ffcdd2
    style V fill:#ffcdd2
    style O fill:#fff3e0
    style S fill:#fff3e0
    style Z fill:#c8e6c9
```

## PortManager.GetPortFromHeaders 流程图

```mermaid
graph TD
    A[获取端口信息请求] --> B[GetPortFromHeaders]
    B --> C{请求方法是否为GET}
    C -->|是| D[从params获取clientId]
    C -->|否| E[从body获取clientId]
    
    D --> F{params中是否存在clientId}
    F -->|存在| G[提取clientId值]
    F -->|不存在| H[尝试从headers获取clientId]
    
    E --> I{body长度是否大于0}
    I -->|是| J[解析JSON body]
    J --> K{是否存在clientId字段}
    K -->|存在| L[提取clientId值]
    K -->|不存在| H
    
    I -->|否| H
    
    H --> M{headers中是否存在clientId}
    M -->|存在| N[提取clientId值]
    M -->|不存在| O[返回错误<br/>clientId is required]
    
    G --> P[设置appName<br/>codebase-indexer]
    L --> P
    N --> P
    
    P --> Q[调用GetPort方法]
    Q --> R[构建缓存key<br/>clientID:appName]
    R --> S[检查缓存]
    
    S --> T{缓存是否存在且未过期}
    T -->|是| U[返回缓存的端口信息]
    T -->|否| V[构建请求URL]
    
    V --> W[发送HTTP请求到端口管理服务]
    W --> X{请求是否成功}
    X -->|失败| Y[返回错误]
    X -->|成功| Z[解析响应JSON]
    
    Z --> AA[更新缓存]
    AA --> BB[更新最后访问时间]
    BB --> CC[返回端口信息]
    
    O --> DD[结束]
    Y --> DD
    U --> DD
    CC --> DD

    style A fill:#e1f5fe
    style DD fill:#c8e6c9
    style O fill:#ffcdd2
    style Y fill:#ffcdd2
    style U fill:#fff3e0
    style CC fill:#c8e6c9