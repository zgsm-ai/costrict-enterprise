# 路由注册流程图

## RegisterHandlers 流程图

```mermaid
graph TD
    A[启动应用] --> B[RegisterHandlers被调用]
    B --> C[注册健康检查路由]
    C --> D{ProxyConfig是否存在}
    D -->|不存在| E[跳过代理路由注册]
    D -->|存在| F[创建SmartProxyHandler]
    
    F --> G[记录使用智能代理处理器日志]
    G --> H[初始化支持的HTTP方法列表]
    H --> I{是否启用动态端口}
    I -->|启用| J[动态端口模式路由注册]
    I -->|未启用| K[静态路由模式路由注册]
    
    J --> L[遍历ProxyConfig.Routes]
    K --> L
    L --> M[遍历HTTP方法列表]
    M --> N[创建路由条目<br/>Path: routeConfig.PathPrefix<br/>Handler: SmartProxyHandler.ServeHTTP]
    N --> O[添加路由到routes数组]
    O --> P{是否还有更多方法}
    P -->|是| M
    P -->|否| Q{是否还有更多路由配置}
    
    Q -->|是| L
    Q -->|否| R[调用server.AddRoutes注册所有路由]
    R --> S[路由注册完成]
    
    E --> S

    style A fill:#e1f5fe
    style S fill:#c8e6c9
    style F fill:#fff3e0
    style N fill:#fff3e0
    style R fill:#c8e6c9
```

## registerHealthCheckRoutes 流程图

```mermaid
graph TD
    A[注册健康检查路由] --> B[创建基础健康检查路由]
    B --> C[路由配置:<br/>Method: GET<br/>Path: /api/v1/proxy/health<br/>Handler: proxyHealthCheckHandler]
    C --> D[添加路由到数组]
    D --> E[调用server.AddRoutes<br/>使用/codebase-indexer前缀]
    E --> F{是否启用动态代理}
    F -->|启用| G[创建动态代理健康检查路由]
    F -->|未启用| H[健康检查路由注册完成]
    
    G --> I[路由配置:<br/>Method: GET<br/>Path: /api/v1/dynamic-proxy/health<br/>Handler: dynamicProxyHealthCheckHandler]
    I --> J[添加路由到数组]
    J --> K[调用server.AddRoutes<br/>使用/codebase-indexer前缀]
    K --> H

    style A fill:#e1f5fe
    style H fill:#c8e6c9
    style G fill:#fff3e0
    style I fill:#fff3e0
    style K fill:#c8e6c9
```

## proxyHealthCheckHandler 流程图

```mermaid
graph TD
    A[健康检查请求] --> B[proxyHealthCheckHandler]
    B --> C{serverCtx.ProxyHandler是否存在}
    C -->|存在| D[调用serverCtx.ProxyHandler.HealthCheck]
    C -->|不存在| E[返回501错误<br/>No proxy handler configured]
    
    D --> F[健康检查完成]
    E --> F

    style A fill:#e1f5fe
    style F fill:#c8e6c9
    style D fill:#e8f5e8
    style E fill:#ffcdd2
```

## dynamicProxyHealthCheckHandler 流程图

```mermaid
graph TD
    A[动态代理健康检查请求] --> B[dynamicProxyHealthCheckHandler]
    B --> C{ProxyConfig是否存在且启用动态端口}
    C -->|是| D[创建DynamicProxyHandler]
    C -->|否| E[返回501错误<br/>Dynamic proxy not configured]
    
    D --> F[调用DynamicProxyHandler.HealthCheck]
    F --> G[健康检查完成]
    E --> G

    style A fill:#e1f5fe
    style G fill:#c8e6c9
    style D fill:#e8f5e8
    style F fill:#e8f5e8
    style E fill:#ffcdd2
```

## 路由匹配和处理流程图

```mermaid
graph TD
    A[HTTP请求到达] --> B[go-zero路由匹配]
    B --> C{请求路径匹配哪个路由}
    
    C -->|匹配/codebase-indexer/api/v1/proxy/health| D[调用proxyHealthCheckHandler]
    C -->|匹配/codebase-indexer/api/v1/dynamic-proxy/health| E[调用dynamicProxyHealthCheckHandler]
    C -->|匹配配置的PathPrefix| F[调用SmartProxyHandler.ServeHTTP]
    C -->|无匹配| G[返回404错误]
    
    D --> H[执行健康检查逻辑]
    E --> I[执行动态代理健康检查逻辑]
    F --> J[执行智能代理处理逻辑]
    
    H --> K[返回健康检查响应]
    I --> K
    J --> L[返回代理响应]
    
    K --> M[请求处理完成]
    L --> M
    G --> M

    style A fill:#e1f5fe
    style M fill:#c8e6c9
    style D fill:#fff3e0
    style E fill:#fff3e0
    style F fill:#fff3e0
    style G fill:#ffcdd2
    style H fill:#e8f5e8
    style I fill:#e8f5e8
    style J fill:#e8f5e8
```

## 配置驱动的路由注册流程图

```mermaid
graph TD
    A[配置文件加载] --> B[解析ProxyConfig]
    B --> C[解析Routes数组]
    C --> D[遍历每个RouteConfig]
    D --> E[提取PathPrefix]
    E --> F[提取Target配置]
    F --> G[为该路径前缀创建所有HTTP方法路由]
    G --> H[将路由添加到注册列表]
    H --> I{是否还有更多RouteConfig}
    
    I -->|是| D
    I -->|否| J[所有路由配置处理完成]
    J --> K[在RegisterHandlers中使用这些配置]
    K --> L[动态注册路由到go-zero服务器]
    
    L --> M[路由注册完成，等待请求]

    style A fill:#e1f5fe
    style M fill:#c8e6c9
    style G fill:#fff3e0
    style H fill:#fff3e0
    style L fill:#c8e6c9
```

## 路由优先级和处理顺序流程图

```mermaid
graph TD
    A[HTTP请求到达] --> B[go-zero路由器开始匹配]
    B --> C[检查静态路由优先级]
    
    C --> D[1. 健康检查路由<br/>/codebase-indexer/api/v1/proxy/health]
    C --> E[2. 动态代理健康检查路由<br/>/codebase-indexer/api/v1/dynamic-proxy/health]
    C --> F[3. 配置的代理路由<br/>/codebase-indexer/{PathPrefix}]
    
    D --> G{路径是否完全匹配}
    E --> G
    F --> H{路径是否匹配PathPrefix}
    
    G -->|匹配| I[执行对应的健康检查处理器]
    G -->|不匹配| J[继续检查下一优先级路由]
    H -->|匹配| K[执行SmartProxyHandler]
    H -->|不匹配| L[路由未找到]
    
    I --> M[返回健康检查响应]
    J --> E
    K --> N[执行智能代理逻辑]
    L --> O[返回404错误]
    
    M --> P[请求处理完成]
    N --> P
    O --> P

    style A fill:#e1f5fe
    style P fill:#c8e6c9
    style I fill:#e8f5e8
    style K fill:#e8f5e8
    style L fill:#ffcdd2
    style O fill:#ffcdd2
    style N fill:#fff3e0