# 隧道启动器测试套件

本目录包含为隧道启动器项目生成的Python测试脚本，涵盖REST API接口和命令行接口的正常场景、异常场景和压力场景测试。

## 测试文件结构

```
test/
├── test_utils.py          # 测试工具类，提供API和CLI测试的通用工具
├── test_rest_api.py       # REST API接口测试
├── test_cli.py           # 命令行接口测试
├── run_all_tests.py      # 测试运行器，执行所有测试并生成报告
└── README.md             # 本说明文件
```

## 测试覆盖范围

### REST API接口测试 (test_rest_api.py)

**正常场景:**
- 健康检查接口
- 创建隧道接口
- 删除隧道接口
- 列出隧道接口
- 获取隧道信息接口

**异常场景:**
- 健康检查（服务器不可用）
- 创建隧道（缺少参数）
- 创建隧道（无效端口）
- 创建隧道（重复创建）
- 删除隧道（隧道不存在）
- 获取隧道信息（隧道不存在）

**压力场景:**
- 并发创建隧道
- API响应时间测试

### 命令行接口测试 (test_cli.py)

**正常场景:**
- 启动隧道命令
- 停止隧道命令
- 列出隧道命令
- 列出隧道（带过滤条件）
- 命令帮助信息

**异常场景:**
- 启动隧道（缺少应用名）
- 启动隧道（无效端口）
- 停止隧道（隧道不存在）
- 停止隧道（缺少应用名）
- 列出隧道（过滤条件不存在）
- 命令无效选项

**压力场景:**
- 并发启动停止隧道
- 列出隧道压力测试

## 环境要求

- Python 3.6+
- requests库
- tunnel-starter可执行文件（在项目根目录）

安装依赖：
```bash
pip install requests
```

## 运行测试

### 运行所有测试

```bash
python test/run_all_tests.py
```

### 运行特定类型测试

```bash
# 只运行REST API测试
python test/run_all_tests.py --type api

# 只运行命令行测试
python test/run_all_tests.py --type cli
```

### 直接运行测试文件

```bash
# 运行REST API所有测试
python test/test_rest_api.py

# 运行REST API单个测试用例
python test/test_rest_api.py --test test_health_check_normal
python test/test_rest_api.py --test test_create_tunnel_normal

# 运行命令行所有测试
python test/test_cli.py

# 运行命令行单个测试用例
python test/test_cli.py --test test_start_tunnel_normal
python test/test_cli.py --test test_list_tunnels_normal
```

### 查看可用测试用例

```bash
# 查看REST API可用测试用例
python test/test_rest_api.py --test nonexistent_test

# 查看命令行可用测试用例
python test/test_cli.py --test nonexistent_test
```

### 测试选项

```bash
# 指定输出文件
python test/run_all_tests.py --output custom_results.json

# 不保存测试结果
python test/run_all_tests.py --no-save
```

## 测试结果

测试完成后会生成详细的测试报告，包括：

- 测试执行时间
- 通过/失败的测试数量
- 各测试套件的详细结果
- 测试耗时统计
- 成功率计算

测试结果默认保存到`test_results.json`文件中。

## 配置说明

### 服务器配置

测试脚本默认连接到`http://localhost:8080`，如果需要修改服务器地址，可以在`test_utils.py`中修改：

```python
self.api_utils = TunnelTestUtils("http://your-server:port")
```

### 可执行文件路径

测试脚本默认使用`./tunnel-starter`作为可执行文件路径，如果需要修改，可以在`test_cli.py`中修改：

```python
self.cli_utils = CommandLineTestUtils("/path/to/tunnel-starter")
```

## 注意事项

1. **服务器启动状态**: 运行REST API测试前，请确保tunnel-starter服务器已启动并运行在正确的端口上。

2. **可执行文件**: 运行命令行测试前，请确保已编译tunnel-starter可执行文件，或使用`go run main.go`替代。

3. **权限问题**: 确保有足够的权限创建隧道和绑定端口。

4. **端口冲突**: 测试会使用一些端口，请确保这些端口没有被其他进程占用。

5. **清理测试数据**: 测试脚本会尝试清理创建的测试数据，但如果测试被中断，可能需要手动清理。

## 故障排除

### 常见问题

1. **连接被拒绝**
   - 确保tunnel-starter服务器正在运行
   - 检查服务器地址和端口配置

2. **命令执行失败**
   - 确保tunnel-starter可执行文件存在
   - 检查文件权限

3. **端口绑定失败**
   - 检查端口是否被其他进程占用
   - 尝试使用不同的端口

4. **测试超时**
   - 增加测试超时时间
   - 检查服务器性能

### 调试模式

要启用更详细的日志输出，可以修改Python日志级别：

```python
import logging
logging.basicConfig(level=logging.DEBUG)
```

## 扩展测试

如果需要添加新的测试用例，可以：

1. 在相应的测试文件中添加新的测试方法
2. 使用`test_utils.py`中提供的工具类
3. 遵循现有的测试命名约定（以`test_`开头）

## 贡献指南

欢迎贡献新的测试用例和改进建议。请确保：

1. 新测试用例有清晰的描述
2. 包含正常场景、异常场景和压力场景
3. 测试完成后清理测试数据
4. 遵循现有的代码风格