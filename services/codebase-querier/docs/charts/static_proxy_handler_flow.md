# StaticProxyHandler 流程图

## 静态代理处理器处理流程

```mermaid
graph TD
    A[HTTP请求到达] --> B[ProxyHandler.ServeHTTP]
    B --> C[记录请求日志]
    C --> D[调用validateRequest验证请求]
    D --> E{验证是否通过}
    E -->|失败| F[调用sendError返回错误]
    E -->|成功| G[调用ProxyLogic.Forward转发请求]
    
    F --> H[结束处理]
    G --> I[记录转发开始日志]
    I --> J[调用buildTargetRequest构建目标请求]
    J --> K{代理模式检查}
    K -->|full_path| L[调用FullPathBuilder.BuildPath]
    K -->|rewrite| M[处理rewrite模式路径]
    
    L --> N[构建完整URL]
    M --> O[移除代理前缀]
    O --> P{是否启用重写}
    P -->|启用| Q[应用重写规则]
    P -->|未启用| R[跳过重写]
    Q --> S[清理路径]
    R --> S
    S --> T[拼接目标URL]
    T --> N
    
    N --> U[解析目标URL]
    U --> V{URL解析是否成功}
    V -->|失败| W[返回内部错误]
    V -->|成功| X[创建HTTP请求]
    
    X --> Y[复制并过滤请求头]
    Y --> Z[设置Host头]
    Z --> AA[设置查询参数]
    AA --> AB[记录调试信息]
    AB --> AC[发送HTTP请求到目标服务]
    
    AC --> AD{请求是否成功}
    AD -->|失败| AE[处理转发错误]
    AE --> AF[返回相应错误响应]
    AD -->|成功| AG[复制响应]
    
    AG --> AH[复制响应头]
    AH --> AI[设置响应状态码]
    AI --> AJ[复制响应体]
    AJ --> AK{复制是否成功}
    AK -->|失败| AL[返回500错误]
    AK -->|成功| AM[记录成功日志]
    
    W --> H
    AF --> H
    AL --> H
    AM --> H

    style A fill:#e1f5fe
    style H fill:#c8e6c9
    style F fill:#ffcdd2
    style AE fill:#ffcdd2
    style W fill:#ffcdd2
    style AL fill:#ffcdd2
    style L fill:#fff3e0
    style Q fill:#fff3e0
    style AC fill:#e8f5e8
    style AG fill:#e8f5e8
    style AM fill:#c8e6c9
```

## ProxyLogic.buildTargetRequest 流程图

```mermaid
graph TD
    A[构建目标请求] --> B[记录调试信息开始]
    B --> C[记录原始请求路径]
    C --> D[记录完整URL]
    D --> E[记录代理模式]
    E --> F{代理模式判断}
    
    F -->|full_path| G[调用pathBuilder.BuildPath]
    G --> H{路径构建是否成功}
    H -->|失败| I[返回内部错误]
    H -->|成功| J[设置fullURL]
    
    F -->|rewrite| K[获取原始路径]
    K --> L[记录rewrite模式初始路径]
    L --> M{路径是否以/proxy开头}
    M -->|是| N[移除/proxy前缀]
    M -->|否| O[保持路径不变]
    
    N --> P[记录移除前缀后的路径]
    O --> P
    P --> Q{是否启用重写}
    Q -->|启用| R[记录重写规则数量]
    R --> S[复制重写规则]
    S --> T[记录重写前路径]
    T --> U[应用路径重写]
    U --> V[记录重写后路径]
    Q -->|未启用| W[跳过重写]
    
    V --> X[清理路径]
    W --> X
    X --> Y[记录清理后的路径]
    Y --> Z[拼接目标URL]
    Z --> J
    
    J --> AA{是否有查询参数}
    AA -->|有| BB[添加查询参数到URL]
    AA -->|无| CC[保持URL不变]
    BB --> DD[记录最终目标URL]
    CC --> DD
    DD --> EE[记录调试信息结束]
    
    EE --> FF[解析目标URL]
    FF --> GG{URL解析是否成功}
    GG -->|失败| I
    GG -->|成功| HH[创建HTTP请求]
    
    HH --> II[复制并过滤请求头]
    II --> JJ{目标URL是否有Host}
    JJ -->|有| KK[设置Host头]
    JJ -->|无| LL[跳过Host头设置]
    
    KK --> MM[设置过滤后的请求头]
    LL --> MM
    MM --> NN[设置查询参数]
    NN --> OO[返回构建好的请求]
    
    I --> PP[结束]
    OO --> PP

    style A fill:#e1f5fe
    style PP fill:#c8e6c9
    style I fill:#ffcdd2
    style GG fill:#ffcdd2
    style G fill:#fff3e0
    style U fill:#fff3e0
    style HH fill:#e8f5e8
    style OO fill:#c8e6c9
```

## ProxyLogic.HealthCheck 流程图

```mermaid
graph TD
    A[健康检查请求] --> B[ProxyLogic.HealthCheck]
    B --> C[记录开始时间]
    C --> D[构建健康检查URL<br/>target.URL/health]
    D --> E[创建GET请求]
    E --> F{请求创建是否成功}
    F -->|失败| G[返回错误]
    F -->|成功| H[发送HTTP请求]
    
    H --> I{请求是否成功}
    I -->|失败| J[返回错误]
    I -->|成功| K[检查响应状态码]
    
    K --> L{状态码是否为2xx}
    L -->|是| M[计算响应时间]
    M --> N[设置healthy=true]
    L -->|否| O[计算响应时间]
    O --> P[设置healthy=false]
    
    N --> Q[返回健康状态]
    P --> Q
    G --> Q
    J --> Q

    style A fill:#e1f5fe
    style Q fill:#c8e6c9
    style G fill:#ffcdd2
    style J fill:#ffcdd2
    style N fill:#c8e6c9
    style P fill:#ffcdd2
    style H fill:#e8f5e8
```

## 路径构建器流程图

### FullPathBuilder.BuildPath 流程

```mermaid
graph TD
    A[构建全路径] --> B{originalPath是否为空}
    B -->|是| C[设置为/]
    B -->|否| D[保持原路径]
    
    C --> E{路径是否以/开头}
    D --> E
    E -->|否| F[添加/前缀]
    E -->|是| G[保持路径格式]
    
    F --> H[解析目标URL]
    G --> H
    H --> I{URL解析是否成功}
    I -->|失败| J[返回解析错误]
    I -->|成功| K[创建URL副本]
    
    K --> L[使用path.Join拼接路径]
    L --> M[清空查询参数]
    M --> N[返回完整URL字符串]
    
    J --> O[结束]
    N --> O

    style A fill:#e1f5fe
    style O fill:#c8e6c9
    style J fill:#ffcdd2
    style F fill:#fff3e0
    style L fill:#fff3e0
    style N fill:#c8e6c9
```

### RewritePathBuilder.BuildPath 流程

```mermaid
graph TD
    A[构建重写路径] --> B{重写规则是否为空}
    B -->|是| C[返回原始路径]
    B -->|否| D[遍历重写规则]
    
    D --> E{当前规则是否匹配路径前缀}
    E -->|匹配| F[执行路径替换]
    E -->|不匹配| G[检查下一条规则]
    
    F --> H[记录路径重写日志]
    H --> I[返回新路径]
    G --> J{是否还有更多规则}
    J -->|是| D
    J -->|否| C
    
    C --> K[结束]
    I --> K

    style A fill:#e1f5fe
    style K fill:#c8e6c9
    style F fill:#fff3e0
    style H fill:#fff3e0
    style I fill:#c8e6c9