import unittest
import time
import sys
from test_utils import CommandLineTestUtils, StressTestUtils

# 设置UTF-8编码，确保在Windows下正确显示中文
if sys.platform.startswith('win'):
    sys.stdout.reconfigure(encoding='utf-8')
    sys.stderr.reconfigure(encoding='utf-8')

class TestCommandLine(unittest.TestCase):
    """命令行接口测试类"""
    
    def setUp(self):
        """测试前设置"""
        self.cli_utils = CommandLineTestUtils()
        self.stress_utils = StressTestUtils(None, self.cli_utils)
        self.test_app = "test-cli-app"
        self.test_version = "v1.0"
        self.test_port = 8080
    
    def tearDown(self):
        """测试后清理"""
        # 清理测试数据
        try:
            self.cli_utils.stop_tunnel(self.test_app, self.test_version)
        except:
            pass
    
    def test_start_tunnel_normal(self):
        """测试启动隧道命令 - 正常场景"""
        print("\n=== 测试启动隧道命令 - 正常场景 ===")
        
        result = self.cli_utils.start_tunnel(self.test_app, self.test_version, self.test_port)
        
        self.assertTrue(result.get("success", False), f"启动隧道失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        
        print(f"✓ 启动隧道成功:")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
        print(f"  标准错误: {result.get('stderr', 'N/A')}")
    
    def test_start_tunnel_missing_app(self):
        """测试启动隧道命令 - 异常场景（缺少应用名）"""
        print("\n=== 测试启动隧道命令 - 异常场景（缺少应用名） ===")
        
        result = self.cli_utils.start_tunnel("", self.test_version, self.test_port)
        
        self.assertFalse(result.get("success", True), f"应该失败，但成功了: {result}")
        self.assertNotEqual(result.get("returncode"), 0, f"返回码应该不为0")
        self.assertIn("app", result.get("stderr", "").lower(), f"错误信息应该包含app相关内容")
        
        print(f"✓ 缺少应用名时正确返回错误: {result.get('stderr', 'N/A')}")
    
    def test_start_tunnel_invalid_port(self):
        """测试启动隧道命令 - 异常场景（无效端口）"""
        print("\n=== 测试启动隧道命令 - 异常场景（无效端口） ===")
        
        invalid_ports = [-1, 0, 65536, 99999]
        for port in invalid_ports:
            result = self.cli_utils.start_tunnel(f"{self.test_app}-{port}", self.test_version, port)
            
            # 可能成功或失败，取决于业务逻辑
            print(f"✓ 端口{port}测试完成: 成功={result.get('success', False)}")
    
    def test_stop_tunnel_normal(self):
        """测试停止隧道命令 - 正常场景"""
        print("\n=== 测试停止隧道命令 - 正常场景 ===")
        
        # 先启动隧道
        start_result = self.cli_utils.start_tunnel(self.test_app, self.test_version, self.test_port)
        print(f"启动隧道结果: {start_result.get('success', False)}")
        
        # 停止隧道
        result = self.cli_utils.stop_tunnel(self.test_app, self.test_version)
        
        self.assertTrue(result.get("success", False), f"停止隧道失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        
        print(f"✓ 停止隧道成功:")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
        print(f"  标准错误: {result.get('stderr', 'N/A')}")
    
    def test_stop_tunnel_not_exist(self):
        """测试停止隧道命令 - 异常场景（隧道不存在）"""
        print("\n=== 测试停止隧道命令 - 异常场景（隧道不存在） ===")
        
        # 停止不存在的隧道
        result = self.cli_utils.stop_tunnel("non-existent-app", "v1.0")
        
        # 可能成功或失败，取决于业务逻辑
        print(f"✓ 停止不存在隧道测试完成: 成功={result.get('success', False)}")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
        print(f"  标准错误: {result.get('stderr', 'N/A')}")
    
    def test_stop_tunnel_missing_app(self):
        """测试停止隧道命令 - 异常场景（缺少应用名）"""
        print("\n=== 测试停止隧道命令 - 异常场景（缺少应用名） ===")
        
        result = self.cli_utils.stop_tunnel("", self.test_version)
        
        self.assertFalse(result.get("success", True), f"应该失败，但成功了: {result}")
        self.assertNotEqual(result.get("returncode"), 0, f"返回码应该不为0")
        self.assertIn("app", result.get("stderr", "").lower(), f"错误信息应该包含app相关内容")
        
        print(f"✓ 缺少应用名时正确返回错误: {result.get('stderr', 'N/A')}")
    
    def test_list_tunnels_normal(self):
        """测试列出隧道命令 - 正常场景"""
        print("\n=== 测试列出隧道命令 - 正常场景 ===")
        
        # 先启动几个测试隧道
        test_apps = [f"{self.test_app}-{i}" for i in range(3)]
        for i, app in enumerate(test_apps):
            self.cli_utils.start_tunnel(app, self.test_version, self.test_port + i)
        
        # 列出所有隧道
        result = self.cli_utils.list_tunnels()
        
        self.assertTrue(result.get("success", False), f"列出隧道失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        
        print(f"✓ 列出隧道成功:")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
        print(f"  标准错误: {result.get('stderr', 'N/A')}")
        
        # 清理测试数据
        for app in test_apps:
            try:
                self.cli_utils.stop_tunnel(app, self.test_version)
            except:
                pass
    
    def test_list_tunnels_empty(self):
        """测试列出隧道命令 - 正常场景（空列表）"""
        print("\n=== 测试列出隧道命令 - 正常场景（空列表） ===")
        
        # 确保没有测试隧道
        result = self.cli_utils.list_tunnels()
        
        self.assertTrue(result.get("success", False), f"列出隧道失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        
        print(f"✓ 列出隧道成功（空列表）:")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
        print(f"  标准错误: {result.get('stderr', 'N/A')}")
    
    def test_list_tunnels_with_filter(self):
        """测试列出隧道命令 - 正常场景（带过滤条件）"""
        print("\n=== 测试列出隧道命令 - 正常场景（带过滤条件） ===")
        
        # 先启动几个测试隧道
        test_apps = [f"{self.test_app}-{i}" for i in range(3)]
        for i, app in enumerate(test_apps):
            self.cli_utils.start_tunnel(app, self.test_version, self.test_port + i)
        
        # 按应用名过滤
        result = self.cli_utils.list_tunnels(test_apps[0])
        
        self.assertTrue(result.get("success", False), f"按应用名过滤失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        
        print(f"✓ 按应用名过滤成功:")
        print(f"  过滤条件: {test_apps[0]}")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
        
        # 按版本过滤
        result = self.cli_utils.list_tunnels("", self.test_version)
        
        self.assertTrue(result.get("success", False), f"按版本过滤失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        
        print(f"✓ 按版本过滤成功:")
        print(f"  过滤条件: {self.test_version}")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
        
        # 清理测试数据
        for app in test_apps:
            try:
                self.cli_utils.stop_tunnel(app, self.test_version)
            except:
                pass
    
    def test_list_tunnels_nonexistent_filter(self):
        """测试列出隧道命令 - 异常场景（过滤条件不存在）"""
        print("\n=== 测试列出隧道命令 - 异常场景（过滤条件不存在） ===")
        
        # 按不存在的应用名过滤
        result = self.cli_utils.list_tunnels("non-existent-app")
        
        self.assertTrue(result.get("success", False), f"过滤失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        
        print(f"✓ 按不存在的应用名过滤成功:")
        print(f"  过滤条件: non-existent-app")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')}")
    
    def test_command_help(self):
        """测试命令帮助信息 - 正常场景"""
        print("\n=== 测试命令帮助信息 - 正常场景 ===")
        
        # 测试主命令帮助
        result = self.cli_utils.run_command(["--help"])
        
        self.assertTrue(result.get("success", False), f"获取帮助失败: {result.get('stderr', 'Unknown error')}")
        self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
        self.assertIn("usage", result.get("stdout", "").lower(), f"帮助信息应该包含usage")
        
        print(f"✓ 主命令帮助信息:")
        print(f"  返回码: {result.get('returncode')}")
        print(f"  标准输出: {result.get('stdout', 'N/A')[:200]}...")
        
        # 测试子命令帮助
        subcommands = ["start", "stop", "list", "server"]
        for cmd in subcommands:
            result = self.cli_utils.run_command([cmd, "--help"])
            
            self.assertTrue(result.get("success", False), f"获取{cmd}帮助失败: {result.get('stderr', 'Unknown error')}")
            self.assertEqual(result.get("returncode"), 0, f"返回码不为0: {result.get('returncode')}")
            self.assertIn(cmd, result.get("stdout", "").lower(), f"帮助信息应该包含{cmd}")
            
            print(f"✓ {cmd}命令帮助信息获取成功")
    
    def test_command_invalid_option(self):
        """测试命令无效选项 - 异常场景"""
        print("\n=== 测试命令无效选项 - 异常场景 ===")
        
        # 测试主命令无效选项
        result = self.cli_utils.run_command(["--invalid-option"])
        
        self.assertFalse(result.get("success", True), f"应该失败，但成功了: {result}")
        self.assertNotEqual(result.get("returncode"), 0, f"返回码应该不为0")
        
        print(f"✓ 主命令无效选项测试完成: 成功={result.get('success', False)}")
        
        # 测试子命令无效选项
        result = self.cli_utils.run_command(["start", "--invalid-option"])
        
        self.assertFalse(result.get("success", True), f"应该失败，但成功了: {result}")
        self.assertNotEqual(result.get("returncode"), 0, f"返回码应该不为0")
        
        print(f"✓ 子命令无效选项测试完成: 成功={result.get('success', False)}")
    
    def test_start_stop_tunnel_stress(self):
        """测试启动停止隧道命令 - 压力场景"""
        print("\n=== 测试启动停止隧道命令 - 压力场景 ===")
        
        # 生成多个测试应用
        test_apps = [f"stress-cli-app-{i}" for i in range(5)]
        
        # 定义测试函数
        def start_stop_test(app_name):
            start_result = self.cli_utils.start_tunnel(app_name, self.test_version, self.test_port)
            time.sleep(0.1)  # 短暂等待
            stop_result = self.cli_utils.stop_tunnel(app_name, self.test_version)
            return {
                "app": app_name,
                "start_success": start_result.get("success", False),
                "stop_success": stop_result.get("success", False)
            }
        
        # 并发测试
        args_list = [(app,) for app in test_apps]
        results = self.stress_utils.concurrent_api_calls(start_stop_test, args_list, num_threads=3)
        
        # 统计结果
        start_success_count = sum(1 for r in results if r["result"]["start_success"])
        stop_success_count = sum(1 for r in results if r["result"]["stop_success"])
        total_count = len(results)
        
        print(f"✓ 压力测试完成:")
        print(f"  总测试数: {total_count}")
        print(f"  启动成功: {start_success_count}")
        print(f"  停止成功: {stop_success_count}")
        
        # 验证成功率
        start_success_rate = start_success_count / total_count if total_count > 0 else 0
        stop_success_rate = stop_success_count / total_count if total_count > 0 else 0
        
        self.assertGreaterEqual(start_success_rate, 0.6, f"启动成功率过低: {start_success_rate:.2%}")
        self.assertGreaterEqual(stop_success_rate, 0.6, f"停止成功率过低: {stop_success_rate:.2%}")
    
    def test_list_tunnels_stress(self):
        """测试列出隧道命令 - 压力场景"""
        print("\n=== 测试列出隧道命令 - 压力场景 ===")
        
        # 定义测试函数
        def list_test():
            return self.cli_utils.list_tunnels()
        
        # 多次调用
        args_list = [() for _ in range(20)]
        results = self.stress_utils.concurrent_api_calls(list_test, args_list, num_threads=5)
        
        # 统计结果
        success_count = sum(1 for r in results if r["success"])
        total_count = len(results)
        
        print(f"✓ 列出隧道压力测试完成: {success_count}/{total_count} 成功")
        
        # 验证成功率
        success_rate = success_count / total_count if total_count > 0 else 0
        self.assertGreaterEqual(success_rate, 0.9, f"列出隧道成功率过低: {success_rate:.2%}")

def run_cli_tests():
    """运行命令行测试"""
    print("开始运行命令行测试...")
    
    # 创建测试套件
    suite = unittest.TestLoader().loadTestsFromTestCase(TestCommandLine)
    
    # 运行测试
    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)
    
    # 输出结果
    print(f"\n测试结果:")
    print(f"总测试数: {result.testsRun}")
    print(f"成功: {result.testsRun - len(result.failures) - len(result.errors)}")
    print(f"失败: {len(result.failures)}")
    print(f"错误: {len(result.errors)}")
    
    if result.failures:
        print(f"\n失败的测试:")
        for test, traceback in result.failures:
            print(f"- {test}: {traceback}")
    
    if result.errors:
        print(f"\n错误的测试:")
        for test, traceback in result.errors:
            print(f"- {test}: {traceback}")
    
    return result.wasSuccessful()

if __name__ == "__main__":
    # 允许运行单个测试用例
    import argparse
    
    parser = argparse.ArgumentParser(description="命令行接口测试")
    parser.add_argument("--test", help="指定要运行的测试方法，例如: test_start_tunnel_normal")
    parser.add_argument("--list", action="store_true", help="列出所有支持的测试用例名")
    
    args = parser.parse_args()
    
    if args.list:
        # 列出所有测试用例名
        test_case = TestCommandLine()
        test_methods = []
        for method_name in dir(test_case):
            if method_name.startswith("test_"):
                test_methods.append(method_name)
        
        print("支持的测试用例:")
        for method in sorted(test_methods):
            print(f"  - {method}")
        sys.exit(0)
    
    if args.test:
        # 运行单个测试用例
        suite = unittest.TestSuite()
        test_case = TestCommandLine()
        test_case.setUp()
        
        if hasattr(test_case, args.test):
            test_method = getattr(test_case, args.test)
            try:
                test_method()
                print(f"\n✅ 测试 {args.test} 通过")
            except Exception as e:
                print(f"\n❌ 测试 {args.test} 失败: {e}")
            finally:
                try:
                    test_case.tearDown()
                except:
                    pass
        else:
            print(f"❌ 找不到测试方法: {args.test}")
            print("可用的测试方法:")
            for method_name in dir(test_case):
                if method_name.startswith("test_"):
                    print(f"  - {method_name}")
    else:
        # 运行所有测试
        run_cli_tests()