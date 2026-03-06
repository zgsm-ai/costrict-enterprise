# 客户端命令行程序管理器(costrict-keeper)

## 1. 目的

在客户端，可能有多个服务，如果完全由vscode扩展进行管理，会让扩展变得复杂。

使用一个独立的命令行程序管理器，管理多个CLI程序的下载、安装、启动、配置、监控、服务注册，可以大大简化vscode扩展的复杂度。

## 2. 技术原理

costrict-keeper管理客户端可用的一系列的组件(component)，以及需要自动运行的各种服务(service)，维护一个高可用的客户端软件服务子系统(system)。

costrict-keeper连接云端，获取各客户端的子系统定义文件，该文件描述了客户端软件服务子系统的构成以及管理方式。

* 组件：根据定义文件，costrict-keeper负责下载需要的各种组件，并维护这些组件的生命周期。
* 服务：根据定义文件，costrict-keeper负责启动各个服务，并监控服务的运行状态。

## 3. 整体方案

costrict-keeper启动后，从`https://zgsm.sangfor.com/costrict/system/<os>/<plat>/platform.json`获取子系统定义文件的版本列表。根据列表定义的版本，获取最新版本的软件服务子系统定义文件。

根据子系统定义文件，下载客户端需要的各种组件，启动需要自动启动的服务。

costrict-keeper提供可靠性、可调试性辅助机制。

1. 组件版本管理：定时获取各组件的版本列表，必要时更新组件。
2. 服务管理：根据用户请求(RESTful API)动态启停服务。
3. 服务接口管理：根据各服务公开的对外接口，生成`.well-known.json`，暴露所有运行中服务的接口
4. 可靠性：服务启动后，costrict-keeper监控服务的运行，保证服务正常运行。如果服务运行发生故障，costrict-keeper会自动重启服务。
5. 可调试性：costrict-keeper实现日志、事件、指标的上报，包括自动上报，手动上报等。

   - 日志上报：客户端各组件统一把日志保存在logs目录，由costrict-keeper统一上报到云端；
   - 事件上报：客户端各组件自行上报事件，如果上报失败，则把事件保存在events目录，由costrict-keeper定时尝试上报；
   - 指标上报：客户端各组件统一实现/metrics接口，costrict-keeper定时从各个组件采集指标，通过pushgateway上报到prometheus。

## 4. 结构设计

### 4.0. 整体说明

costrict-keeper管理costrict客户端子系统的软件环境，包含以下类型：

- SystemSpecification: 系统定义，该类型定义一个子系统的构成。定义某个版本的客户端子系统，应该包含哪些组件，以及服务。SystemSpecification中组件定义的类型是ComponentSpecification，服务定义的类型是ServiceSpecification；
- Component: 组件，该类型记录一个组件的状态，比如是否已安装到本地，本地版本号等；
- Service: 服务，服务是一种需要自动运行的组件
- SystemKnownledge: 信息关键信息，主要包含服务的关键信息，即Service

### 4.1. SystemSpecification

系统定义SystemSpecification序列化为JSON文件(system-spec.json)，放在云端，供各个客户端下载。

JSON文件格式：

```json
{
    "configuration": "1.0.0",   //配置文件格式的版本
    "platform": "windows",      //配置文件适用的平台
    "arch": "amd64",            //配置文件适用的平台
    "version": "1.2.0",         //配置描述的软件包的版本，即构建出来的windows子系统的版本
    "manager": {                //服务管理器本身，服务管理器本身也是一个服务
        "name": "costrict",     //即costrict-keeper
        "version": "^1.0.0",
        "upgrade": {
            "mode": "auto",     //升级模式
            "lowest": "1.0.1",  //最低版本，当前版本低于该版本，强制升级到该版本
            "highest": "1.2.3"  //最高版本，超过该版本不自动升级
        }
    },
    "components": [{            //需要costrict-keeper管理的组件
        "name": "codebase-syncer",  //程序名称
        "version": "^1.0.0",    //可支持版本范围
        "upgrade": {
            "mode": "auto",     //升级模式
            "lowest": "1.0.1",  //最低版本，当前版本低于该版本，强制升级到该版本
            "highest": "1.2.3"  //最高版本，超过该版本不自动升级
        }
    }, {
        "name": "codebase-indexer",
        "version": "^1.1.0"
    }, {
        "name": "codebase-parser",
        "version": "~1.1.1"
    }, {
        "name": "cotun",
        "version": "~1.0.0"
    }, {
        "name": "tunnel-client",
        "version": "~1.0.0"
    }, {
        "name": "cleaner",
        "version": ">=1.1.0"
    }],
    "services": [{              //需要costrict-keeper管理的服务
        "name": "codebase-syncer",  //程序名称
        "startup": "always",        //启动模式：always=常驻, once=运行一次, none=不自动运行
        "command": "codebase-syncer -s",//设定启动的命令行(比如服务模式启动codebase-syncer),如果不指定，则以不带参数方式启动
        "protocol": "http",         //服务对外接口协议
        "port": 8080,               //建议服务端口，实际运行时根据客户端情况会调整
        "metrics": "/metrics",      //指标采集接口的地址
        "healthy": "/healthy",      //服务可访问性检测接口
        "accessible": "local"       //可访问性：remote(远程访问)/local(本地访问)
    }, {
        "name": "codebase-indexer",
        "startup": "always",
        "protocol": "http",
        "port": 8081
    }, {
        "name": "codebase-parser",
        "startup": "always",
        "protocol": "http",
        "port": 8082
    }, {
        "name": "tunnel-client",
        "startup": "always"
    }, {
        "name": "cleaner",
        "startup": "once"
    }]
}
```

### 4.2. Service

服务对象(Service)序列化为JSON后格式如下：

```json
{
    "name": "tunnel-client",
    "version": "1.1.0",
    "installed": true,
    "startup": "always",
    "status": "running",        //exited(未运行) -> running(运行中) -> exited(未运行)
    "protocol": "http",         //服务对外接口协议
    "port": 8080              //服务端口
}
```

### 4.3. Component

组件对象(Component)序列化为JSON后格式如下:

```json
{
    "name": "tunnel-client",
    "version": "1.1.0",
    "installed": true
}
```

### 4.4. SystemKnownledge

SystemKnownledge保存软件服务子系统的各项参数，该数据结构默认序列化为`%USERPROFILE%/.costrict/share/.well-known.json`

序列化后格式如下：

```json
{
    "logs": {
        "dir": "%USERPROFILE%\\.costrict\\logs",
        "level": "INFO"
    },
    "services": [{
        "name": "tunnel-client",
        "version": "1.1.0",
        "installed": true,      //是否已经安装到本地
        "startup": "always",    //启动方式
        "protocol": "http",     //服务对外接口协议
        "port": 8080,           //服务端口
        "status": "exited"      //exited(未运行) -> running(运行中) -> exited(未运行)
    }, {
        "name": "costrict",     //即costrict-keeper
        "version": "1.0.0",
        "installed": true,
        "startup": "always",
        "protocol": "http",
        "port": 8081,
        "status": "running"
    }]
}
```

### 4.5. 客户端软件服务子系统数据存储区(Storage)

Storage是一个保存软件服务子系统数据的目录，位于`%USERPROFILE%/.costrict`(或`$HOME/.costrict`)。

其结构如下：

```
.costrict-+-/bin:  可执行文件区
          +-/package: 包安装记录区
          +-/logs: 日志区
          +-/share: 交换区(共享数据区,vscode扩展与子系统的数据交换)
          +-/cache: 运行状态区(缓存costrict-keeper管理的组件&服务的状态，保证服务和CLI可以并发工作，无惧程序崩溃)
```

### 4.6. 软件分发服务器的包目录

客户端可通过URL `https://zgsm.sangfor.com/costrict/` 下载软件，该URL下保存多个包，目录结构如下：

```sh
#
# 包管理系统的目录结构：
#
#-/-+-<package>/-+-<os>/-+-<arch>/-+-<ver>/-+-package.json: 对包数据文件进行签名保护
#   |            |       |         |        +-<package-data-file>
#   |            |       |         +-platform.json: 某个平台支持哪些版本
#   |            |       +-amd64-...
#   |            +-windows-...
#   |            +-platforms.json: 某个包支持哪些平台(OS&芯片架构)
#   +-packages.json: 系统有哪些包可以下载
#
```

### 4.7. 用户信息

以下信息由vscode写入，供客户端所有程序读取。

```json
{
    "user": "1382436ddd",
    "username": "zbc",
    "access_token": "xxx",
    "machine_id": "",
    "base_url": ""
}

```

## 5. 接口

### 5.1. RESTful API接口

costrict-keeper可以作为服务进行启动，通过RESTful API对外提供服务。

#### 5.1.1. 组件管理接口

升级指定组件：

```
POST /costrict/api/v1/components/{component}/upgrade?version=1.1.0
```

卸载组件：

```
DELETE /costrict/api/v1/components/{component}
```

获取组件列表：

```
GET /costrict/api/v1/components
```

#### 5.1.2. 服务管理接口

##### 5.1.2.1. 启动服务

```
POST /costrict/api/v1/services/{service}/start
```

##### 5.1.2.2. 停止服务

```
POST /costrict/api/v1/services/{service}/stop
```

##### 5.1.2.3. 重启服务

```
POST /costrict/api/v1/services/{service}/restart
```

##### 5.1.2.4. 获取服务列表

```
GET /costrict/api/v1/services
```

##### 5.1.2.5. 获取服务信息(状态&地址等)

```
GET /costrict/api/v1/services/{service}
```

##### 5.1.2.6. 修改服务参数

```
PUT /costrict/api/v1/services/{service}
```

```
PATCH /costrict/api/v1/services/{service}
```

### 5.2. 命令行接口

costrict-keeper可以通过命令行方式调用，以非交互的方式，执行子命令。

costrict-keeper执行的子命令，包括作为服务启动，升级组件，启停服务，上报日志，上报prometheus指标，生成服务`.well-known`文件。

#### 5.2.1. 作为HTTP服务器启动

```sh
costrict server --listen 8080 --config appdata/costrict.json
```

#### 5.2.2. 升级组件

```sh
costrict upgrade codebase-syncer --version 1.2.1
```

#### 5.2.3. 查看组件详情

```sh
costrict list codebase-syncer
```

#### 5.2.3. 启动服务

```sh
costrict service start codebase-syncer
```

#### 5.2.4. 停止服务

```sh
costrict service stop codebase-syncer
```

#### 5.2.5. 重启服务

```sh
costrict service restart codebase-syncer
```

#### 5.2.6. 查看服务和组件的信息

```sh
costrict service status
```

#### 5.2.7. 生成服务.well-known.json文件

```sh
costrict service known --output ./service-statuses.json
```

output未指定，则默认保存到`%USERPROFILE%/.costrict/share/.well-known.json`

#### 5.2.8. 上报日志

采集并打包日志：

```sh
costrict logs save --output costrict-logs.zip 
```

根据配置中的日志目录，将该目录下的所有日志文件打包，输出为output参数指定的压缩包文件。

采集并上传日志：

```sh
costrict logs upload [--addr upload-logs-addr]
```

根据配置中的日志目录，将该目录下的所有日志文件打包，将打包后的日志上传到costrict云端。

接收上报日志的API地址，由命令行参数addr指定，如果没有指定，从配置文件logs.addr中获取。

#### 5.2.9. 上报指标

```sh
costrict metrics [--addr pushgateway-api-addr]
```

根据各服务配置的metrics接口，采集指标数据，并合并上报给pushgateway。

pushgateway提供的上报地址，从命令行参数中获取，如果命令行参数没指定，则从配置文件metrics.pushgateway中获取。

## 6. 流程设计

### 6.1. vscode扩展与costrict-keeper交互

```mermaid
sequenceDiagram
    participant IDE
    participant Cloud
    participant CostrictKeeper
    participant Services
    
    IDE->>IDE: S0: 将访问云端的access-token写到交换区
    IDE->>Cloud: S1: 获取CostrictKeeper包版本列表
    IDE->>Cloud: S2: 获取Latest版本的CostrictKeeper软件包
    IDE->>IDE: S3: 安装启动CostrictKeeper
    
    loop 5分钟内每5秒轮询一次,后续逐渐加长到30s
        IDE->>CostrictKeeper: S4: 获取服务列表，获得服务侦听端口
        alt 服务为可用状态
            IDE->>Services: S5: 调用服务接口，如同步数据等
        else 服务
            IDE->>IDE: 显示服务暂不可用
        end
    end
```

说明：

| 步骤 | 步骤逻辑 | 说明 |
|----|----|-----|
| S0 | 将访问云端的access-token写到交换区 | 保存到`%USERPROFILE%/.costrict/share/access-token` |
| S1 | 获取CostrictKeeper包版本列表 | URL: `https://zgsm.sangfor.com/costrict/costrict/<os>/<plat>/platform.json`|
| S2 | 获取Latest版本的CostrictKeeper软件包 | 根据`platform.json`中的URL获取包描述文件以及包数据文件（做法和上一版本相同） |
| S3 | 安装启动CostrictKeeper | 需要校验包的合法性（做法和上一版本相同） |
| S4 | 获取服务列表，获得服务侦听端口 |获取服务列表有两种方法：RESTful接口，或直接读取`%USERPROFILE%/.costrict/share/.well-known.json`文件<br/>服务列表即Service的数组，Service数据格式参考：`4.2. Service`。<br/>CostrictKeeper本身的接口也在`.well-known.json`中可查 |

