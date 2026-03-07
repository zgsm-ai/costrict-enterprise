# SmartProxyHandler 流程图

## 智能代理处理器处理流程

```mermaid
graph TD
    A[HTTP请求到达] --> B[SmartProxyHandler.ServeHTTP]
    B --> C{检查是否启用基于请求头的转发}
    C -->|启用| D[遍历HeaderBasedForward.Paths]
    D --> E{请求路径是否匹配配置路径}
    E -->|匹配| F{检查请求头是否存在}
    F -->|存在| G[转发到WithHeaderURL]
    F -->|不存在| H[转发到WithoutHeaderURL]
    G --> I[结束处理]
    H --> I
    E -->|不匹配| J{检查X-Costrict-Version请求头}
    
    C -->|未启用| J
    
    J -->|存在| K[使用DynamicProxyHandler]
    J -->|不存在| L{检查ForwardURL配置}
    L -->|已配置| M[使用StaticProxyHandler]
    L -->|未配置| K
    
    K --> N[DynamicProxyHandler.ServeHTTP]
    M --> O[StaticProxyHandler.ServeHTTP]
    
    N --> P[获取clientId]
    P --> Q[调用PortManager.GetPortFromHeaders]
    Q --> R[获取端口信息]
    R --> S[构建目标URL]
    S --> T[转发请求到目标服务]
    T --> U[返回响应]
    
    O --> V[ProxyLogic.Forward]
    V --> W[构建目标请求]
    W --> X[执行HTTP调用]
    X --> Y[复制响应]
    Y --> U
    
    U --> I

    style A fill:#e1f5fe
    style I fill:#c8e6c9
    style G fill:#fff3e0
    style H fill:#fff3e0
    style K fill:#e8f5e8
    style M fill:#f3e5f5
    style N fill:#e8f5e8
    style O fill:#f3e5f5
```

## SmartProxyHandler.HealthCheck 流程图

```mermaid
graph TD
    A[健康检查请求] --> B[SmartProxyHandler.HealthCheck]
    B --> C[创建HealthStatus结构]
    C --> D[检查动态代理健康状态]
    D --> E[调用checkDynamicProxyHealth]
    E --> F[创建临时请求]
    F --> G[复制请求头]
    G --> H[创建响应记录器]
    H --> I[调用DynamicProxyHandler.HealthCheck]
    I --> J[记录健康状态]
    J --> K[检查静态代理健康状态]
    
    K --> L{StaticProxyHandler是否存在}
    L -->|存在| M[调用checkStaticProxyHealth]
    M --> N[调用ProxyLogic.HealthCheck]
    N --> O[发送健康检查请求]
    O --> P[检查响应状态]
    P --> Q[记录健康状态]
    L -->|不存在| R[跳过静态代理检查]
    
    Q --> S{检查HeaderBasedForward是否启用}
    R --> S
    S -->|启用| T[调用getHeaderBasedForwardStatus]
    T --> U[构建状态信息]
    S -->|未启用| V[跳过基于请求头的转发检查]
    
    U --> W[构建最终响应]
    V --> W
    W --> X[返回JSON响应]
    
    style A fill:#e1f5fe
    style X fill:#c8e6c9
    style I fill:#e8f5e8
    style N fill:#f3e5f5
    style U fill:#fff3e0