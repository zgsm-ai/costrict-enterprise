# 整体架构流程图

## 系统整体架构流程

```mermaid
graph TB
    subgraph "客户端层"
        A[HTTP客户端]
        B[浏览器/移动应用]
        C[API调用方]
    end
    
    subgraph "负载均衡层"
        D[负载均衡器<br/>Nginx/ALB]
    end
    
    subgraph "代理服务层"
        E[HTTP Server<br/>go-zero框架]
        F[智能路由选择器<br/>SmartProxyHandler]
        G[路由注册表<br/>Routes Registry]
        H[健康检查管理器<br/>Health Manager]
    end
    
    subgraph "转发策略层"
        I[动态代理处理器<br/>DynamicProxyHandler]
        J[静态代理处理器<br/>StaticProxyHandler]  
        K[基于请求头的转发<br/>Header-based Forward]
        L[端口管理器<br/>PortManager]
    end
    
    subgraph "目标服务层"
        M[动态端口服务<br/>Dynamic Port Services]
        N[固定目标服务<br/>Static Target Services]
        O[端口管理服务<br/>Port Manager Service]
        P[健康检查端点<br/>Health Endpoints]
    end
    
    subgraph "缓存层"
        Q[端口信息缓存<br/>Port Cache]
        R[配置缓存<br/>Config Cache]
    end
    
    A --> D
    B --> D
    C --> D
    D --> E
    E --> F
    F --> G
    
    G -->|路由匹配| F
    F -->|策略选择| I
    F -->|策略选择| J
    F -->|策略选择| K
    
    I --> L
    L -->|端口查询| O
    O -->|端口信息| L
    L -->|缓存更新| Q
    Q -->|缓存命中| L
    
    I -->|构建URL| M
    J -->|直接转发| N
    K -->|条件转发| M
    K -->|条件转发| N
    
    M -->|健康检查| P
    N -->|健康检查| P
    P -->|健康状态| H
    
    H -->|健康监控| F
    H -->|健康监控| I
    H -->|健康监控| J
    
    R -->|配置提供| F
    R -->|配置提供| I
    R -->|配置提供| J
    R -->|配置提供| K

    style A fill:#e1f5fe
    style B fill:#e1f5fe
    style C fill:#e1f5fe
    style E fill:#fff3e0
    style F fill:#fff3e0
    style I fill:#e8f5e8
    style J fill:#e8f5e8
    style K fill:#e8f5e8
    style L fill:#e8f5e8
    style M fill:#c8e6c9
    style N fill:#c8e6c9
    style Q fill:#f3e5f5
    style R fill:#f3e5f5
```

## 请求处理完整流程

```mermaid
graph TD
    A[客户端发起请求] --> B[负载均衡器路由]
    B --> C[go-zero HTTP Server接收]
    C --> D[路由匹配器处理]
    
    D --> E{请求路径匹配}
    E -->|健康检查路径| F[健康检查处理器]
    E -->|代理路径| G[智能代理处理器]
    E -->|未知路径| H[404错误]
    
    F --> I[执行健康检查逻辑]
    I --> J[返回健康状态]
    J --> K[响应返回客户端]
    
    G --> L{智能路由策略选择}
    L -->|X-Costrict-Version存在| M[动态代理策略]
    L -->|ForwardURL配置| N[静态代理策略]
    L -->|Header规则匹配| O[基于请求头的转发]
    L -->|默认策略| M
    
    M --> P[获取clientId]
    P --> Q[调用PortManager]
    Q --> R{缓存检查}
    R -->|缓存命中| S[返回缓存端口]
    R -->|缓存未命中| T[调用端口管理服务]
    T --> U[获取端口信息]
    U --> V[更新缓存]
    V --> S
    
    S --> W[构建目标URL]
    W --> X[转发请求到动态服务]
    
    N --> Y[构建固定目标URL]
    Y --> Z[转发请求到静态服务]
    
    O --> AA{请求头检查}
    AA -->|有特定头| AB[转发到WithHeaderURL]
    AA -->|无特定头| AC[转发到WithoutHeaderURL]
    
    X --> AD[目标服务处理请求]
    Y --> AD
    AB --> AD
    AC --> AD
    
    AD --> AE[目标服务返回响应]
    AE --> AF[复制响应头和状态码]
    AF --> AG[复制响应体]
    AG --> AH[响应返回客户端]
    
    H --> AI[返回404错误响应]
    AI --> AH
    
    K --> AH

    style A fill:#e1f5fe
    style AH fill:#c8e6c9
    style F fill:#fff3e0
    style G fill:#fff3e0
    style M fill:#e8f5e8
    style N fill:#e8f5e8
    style O fill:#e8f5e8
    style S fill:#f3e5f5
    style V fill:#f3e5f5
    style X fill:#c8e6c9
    style Y fill:#c8e6c9
    style AD fill:#c8e6c9
    style H fill:#ffcdd2
```

## 健康检查监控流程

```mermaid
graph TD
    A[健康检查定时任务] --> B[SmartProxyHandler.HealthCheck]
    B --> C[检查动态代理健康状态]
    C --> D[创建测试请求]
    D --> E[调用DynamicProxyHandler.HealthCheck]
    E --> F[获取clientId]
    F --> G[调用PortManager]
    G --> H[获取端口信息]
    H --> I[构建健康检查URL]
    I --> J[发送健康检查请求]
    J --> K{响应状态}
    K -->|健康| L[记录动态代理健康]
    K -->|不健康| M[记录动态代理异常]
    
    B --> N[检查静态代理健康状态]
    N --> O{StaticProxyHandler是否存在}
    O -->|存在| P[调用ProxyLogic.HealthCheck]
    O -->|不存在| Q[跳过静态代理检查]
    
    P --> R[发送健康检查到静态服务]
    R --> S{响应状态}
    S -->|健康| T[记录静态代理健康]
    S -->|不健康| U[记录静态代理异常]
    
    B --> V[检查基于请求头的转发配置]
    V --> W{HeaderBasedForward是否启用}
    W -->|启用| X[检查配置完整性]
    W -->|未启用| Y[跳过检查]
    
    X --> Z[验证Header配置]
    Z --> AA[验证路径配置]
    AA --> AB[记录配置状态]
    
    L --> AC[聚合健康状态]
    M --> AC
    T --> AC
    U --> AC
    Q --> AC
    Y --> AC
    AB --> AC
    
    AC --> AD[生成综合健康报告]
    AD --> AE[更新监控指标]
    AE --> AF[记录健康检查日志]
    AF --> AG[等待下次检查]

    style A fill:#e1f5fe
    style AG fill:#c8e6c9
    style L fill:#c8e6c9
    style T fill:#c8e6c9
    style AB fill:#c8e6c9
    style M fill:#ffcdd2
    style U fill:#ffcdd2
    style J fill:#e8f5e8
    style R fill:#e8f5e8
```

## 配置管理和热更新流程

```mermaid
graph TD
    A[配置文件加载] --> B[解析ProxyConfig]
    B --> C[验证配置完整性]
    C --> D{配置是否有效}
    D -->|无效| E[记录错误并退出]
    D -->|有效| F[初始化SmartProxyHandler]
    
    F --> G[初始化DynamicProxyHandler]
    G --> H[初始化PortManager]
    H --> I[初始化StaticProxyHandler]
    I --> J[注册路由]
    J --> K[启动服务]
    
    L[配置文件变更监听] --> M{检测到变更}
    M -->|是| N[重新加载配置文件]
    N --> O[解析新配置]
    O --> P{配置是否有效}
    P -->|无效| Q[保持旧配置运行]
    P -->|有效| R[平滑切换配置]
    
    R --> S[更新SmartProxyHandler配置]
    S --> T[更新DynamicProxyHandler配置]
    T --> U[更新PortManager配置]
    U --> V[更新StaticProxyHandler配置]
    V --> W[重新注册路由]
    W --> X[配置更新完成]
    
    Q --> Y[记录配置错误]
    Y --> Z[继续使用旧配置]
    Z --> X
    
    E --> AA[服务启动失败]
    K --> AB[服务运行中]
    X --> AB

    style A fill:#e1f5fe
    style AB fill:#c8e6c9
    style F fill:#fff3e0
    style G fill:#fff3e0
    style H fill:#fff3e0
    style I fill:#fff3e0
    style R fill:#e8f5e8
    style S fill:#e8f5e8
    style T fill:#e8f5e8
    style U fill:#e8f5e8
    style V fill:#e8f5e8
    style E fill:#ffcdd2
    style Q fill:#ffcdd2
```

## 错误处理和恢复流程

```mermaid
graph TD
    A[请求处理过程中发生错误] --> B{错误类型判断}
    
    B -->|请求验证错误| C[ProxyHandler.validateRequest]
    B -->|端口获取错误| D[PortManager.GetPortFromHeaders]
    B -->|目标连接错误| E[ProxyLogic.Forward]
    B -->|响应处理错误| F[ProxyHandler.copyResponse]
    B -->|超时错误| G[HTTP客户端超时]
    
    C --> H[生成ProxyError]
    D --> H
    E --> H
    F --> H
    G --> H
    
    H --> I{错误码判断}
    I -->|PROXY_BAD_REQUEST| J[返回400错误]
    I -->|PROXY_TARGET_UNREACHABLE| K[返回503错误]
    I -->|PROXY_TIMEOUT| L[返回504错误]
    I -->|PROXY_INTERNAL_ERROR| M[返回500错误]
    I -->|其他错误| N[返回500错误]
    
    J --> O[记录错误日志]
    K --> O
    L --> O
    M --> O
    N --> O
    
    O --> P[错误响应返回客户端]
    P --> Q[错误处理完成]
    
    R[错误恢复机制] --> S[检测错误频率]
    S --> T{错误频率是否过高}
    T -->|是| U[触发熔断机制]
    T -->|否| V[继续正常处理]
    
    U --> W[暂时停止转发]
    W --> X[定期尝试恢复]
    X --> Y{恢复是否成功}
    Y -->|成功| Z[恢复正常服务]
    Y -->|失败| W
    
    Z --> V

    style A fill:#ffcdd2
    style Q fill:#c8e6c9
    style H fill:#ffcdd2
    style J fill:#ffcdd2
    style K fill:#ffcdd2
    style L fill:#ffcdd2
    style M fill:#ffcdd2
    style N fill:#ffcdd2
    style O fill:#fff3e0
    style U fill:#fff3e0
    style Z fill:#c8e6c9