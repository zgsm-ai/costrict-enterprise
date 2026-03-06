import unittest
import json
import time
import sys
from test_utils import TunnelTestUtils, StressTestUtils

# 设置UTF-8编码，确保在Windows下正确显示中文
if sys.platform.startswith('win'):
    sys.stdout.reconfigure(encoding='utf-8')
    sys.stderr.reconfigure(encoding='utf-8')

class TestRestAPI(unittest.TestCase):
    """REST API测试类"""
    
    def setUp(self):
        """测试前设置"""
        self.api_utils = TunnelTestUtils()
        self.stress_utils = StressTestUtils(self.api_utils, None)
        self.test_app = "test-api-app"
        self.test_version = "v1.0"
        self.test_port = 8080
    
    def tearDown(self):
        """测试后清理"""
        # 清理测试数据
        try:
            self.api_utils.delete_tunnel(self.test_app, self.test_version)
        except:
            pass
    
    def test_health_check_normal(self):
        """测试健康检查接口 - 正常场景"""
        print("\n=== 测试健康检查接口 - 正常场景 ===")
        
        result = self.api_utils.health_check()
        
        self.assertEqual(result, "OK", f"健康检查失败，期望: OK, 实际: {result}")
        print(f"✓ 健康检查成功: {result}")
    
    def test_health_check_server_down(self):
        """测试健康检查接口 - 异常场景（服务器不可用）"""
        print("\n=== 测试健康检查接口 - 异常场景（服务器不可用） ===")
        
        # 使用不存在的服务器地址
        temp_utils = TunnelTestUtils("http://localhost:9999")
        result = temp_utils.health_check()
        
        self.assertIn("error", result.lower(), f"应该返回错误，实际: {result}")
        print(f"✓ 服务器不可用时健康检查正确返回错误: {result}")
    
    def test_create_tunnel_normal(self):
        """测试创建隧道接口 - 正常场景"""
        print("\n=== 测试创建隧道接口 - 正常场景 ===")
        
        result = self.api_utils.create_tunnel(self.test_app, self.test_version, self.test_port)
        
        self.assertNotIn("error", str(result), f"创建隧道失败: {result}")
        self.assertEqual(result.get("name"), self.test_app, f"应用名称不匹配")
        self.assertEqual(result.get("version"), self.test_version, f"版本不匹配")
        self.assertEqual(result.get("localPort"), self.test_port, f"端口不匹配")
        
        print(f"✓ 创建隧道成功: {json.dumps(result, indent=2, ensure_ascii=False)}")
    
    def test_create_tunnel_missing_params(self):
        """测试创建隧道接口 - 异常场景（缺少参数）"""
        print("\n=== 测试创建隧道接口 - 异常场景（缺少参数） ===")
        
        # 缺少app参数
        result = self.api_utils.create_tunnel("", self.test_version, self.test_port)
        
        self.assertIn("error", str(result).lower(), f"应该返回错误，实际: {result}")
        print(f"✓ 缺少app参数时正确返回错误: {result}")
        
        # 缺少port参数（使用0作为无效端口）
        result = self.api_utils.create_tunnel(self.test_app, self.test_version, 0)
        
        self.assertIn("error", str(result).lower(), f"应该返回错误，实际: {result}")
        print(f"✓ 缺少port参数时正确返回错误: {result}")
    
    def test_create_tunnel_invalid_port(self):
        """测试创建隧道接口 - 异常场景（无效端口）"""
        print("\n=== 测试创建隧道接口 - 异常场景（无效端口） ===")
        
        # 使用无效端口
        invalid_ports = [-1, 0, 65536, 99999]
        for port in invalid_ports:
            result = self.api_utils.create_tunnel(f"{self.test_app}-{port}", self.test_version, port)
            
            self.assertIn("error", str(result).lower(), f"端口{port}应该返回错误，实际: {result}")
            print(f"✓ 无效端口{port}时正确返回错误: {result.get('error', 'Unknown error')}")
    
    def test_create_tunnel_duplicate(self):
        """测试创建隧道接口 - 异常场景（重复创建）"""
        print("\n=== 测试创建隧道接口 - 异常场景（重复创建） ===")
        
        # 第一次创建
        result1 = self.api_utils.create_tunnel(self.test_app, self.test_version, self.test_port)
        self.assertNotIn("error", str(result1), f"第一次创建失败: {result1}")
        
        # 第二次创建（相同参数）
        result2 = self.api_utils.create_tunnel(self.test_app, self.test_version, self.test_port)
        
        # 可能成功或失败，取决于业务逻辑
        print(f"✓ 重复创建测试完成，第一次: {result1.get('name', 'N/A')}, 第二次: {result2}")
    
    def test_delete_tunnel_normal(self):
        """测试删除隧道接口 - 正常场景"""
        print("\n=== 测试删除隧道接口 - 正常场景 ===")
        
        # 先创建隧道
        create_result = self.api_utils.create_tunnel(self.test_app, self.test_version, self.test_port)
        self.assertNotIn("error", str(create_result), f"创建隧道失败: {create_result}")
        
        # 删除隧道
        result = self.api_utils.delete_tunnel(self.test_app, self.test_version)
        
        self.assertNotIn("error", str(result), f"删除隧道失败: {result}")
        self.assertEqual(result.get("appName"), self.test_app, f"应用名称不匹配")
        self.assertEqual(result.get("status"), "success", f"状态不匹配")
        
        print(f"✓ 删除隧道成功: {json.dumps(result, indent=2, ensure_ascii=False)}")
    
    def test_delete_tunnel_not_exist(self):
        """测试删除隧道接口 - 异常场景（隧道不存在）"""
        print("\n=== 测试删除隧道接口 - 异常场景（隧道不存在） ===")
        
        # 删除不存在的隧道
        result = self.api_utils.delete_tunnel("non-existent-app", "v1.0")
        
        # 可能返回错误或成功，取决于业务逻辑
        print(f"✓ 删除不存在隧道测试完成: {result}")
    
    def test_list_tunnels_normal(self):
        """测试列出隧道接口 - 正常场景"""
        print("\n=== 测试列出隧道接口 - 正常场景 ===")
        
        # 创建几个测试隧道
        test_apps = [f"{self.test_app}-{i}" for i in range(3)]
        created_tunnels = []
        
        for app in test_apps:
            result = self.api_utils.create_tunnel(app, self.test_version, self.test_port + int(app.split('-')[-1]))
            if "error" not in str(result):
                created_tunnels.append(result)
        
        # 列出所有隧道
        result = self.api_utils.list_tunnels()
        
        self.assertIsInstance(result, list, f"返回结果应该是列表，实际: {type(result)}")
        print(f"✓ 列出隧道成功，共{len(result)}个隧道")
        
        # 验证创建的隧道在列表中
        for tunnel in created_tunnels:
            found = any(t.get("name") == tunnel.get("name") for t in result)
            self.assertTrue(found, f"创建的隧道{tunnel.get('name')}不在列表中")
        
        # 清理测试数据
        for app in test_apps:
            try:
                self.api_utils.delete_tunnel(app, self.test_version)
            except:
                pass
    
    def test_list_tunnels_empty(self):
        """测试列出隧道接口 - 正常场景（空列表）"""
        print("\n=== 测试列出隧道接口 - 正常场景（空列表） ===")
        
        # 确保没有测试隧道
        result = self.api_utils.list_tunnels()
        
        self.assertIsInstance(result, list, f"返回结果应该是列表，实际: {type(result)}")
        print(f"✓ 列出隧道成功，共{len(result)}个隧道")
    
    def test_get_tunnel_info_normal(self):
        """测试获取隧道信息接口 - 正常场景"""
        print("\n=== 测试获取隧道信息接口 - 正常场景 ===")
        
        # 先创建隧道
        create_result = self.api_utils.create_tunnel(self.test_app, self.test_version, self.test_port)
        self.assertNotIn("error", str(create_result), f"创建隧道失败: {create_result}")
        
        # 获取隧道信息
        result = self.api_utils.get_tunnel_info(self.test_app)
        
        self.assertNotIn("error", str(result), f"获取隧道信息失败: {result}")
        self.assertEqual(result.get("name"), self.test_app, f"应用名称不匹配")
        self.assertEqual(result.get("version"), self.test_version, f"版本不匹配")
        
        print(f"✓ 获取隧道信息成功: {json.dumps(result, indent=2, ensure_ascii=False)}")
    
    def test_get_tunnel_info_not_exist(self):
        """测试获取隧道信息接口 - 异常场景（隧道不存在）"""
        print("\n=== 测试获取隧道信息接口 - 异常场景（隧道不存在） ===")
        
        # 获取不存在隧道的信息
        result = self.api_utils.get_tunnel_info("non-existent-app")
        
        self.assertIn("error", str(result).lower(), f"应该返回错误，实际: {result}")
        print(f"✓ 获取不存在隧道信息时正确返回错误: {result}")
    
    def test_create_tunnel_stress(self):
        """测试创建隧道接口 - 压力场景"""
        print("\n=== 测试创建隧道接口 - 压力场景 ===")
        
        # 生成多个测试应用
        test_apps = [f"stress-app-{i}" for i in range(10)]
        
        # 并发创建隧道
        args_list = [(app, self.test_version, self.test_port + i) for i, app in enumerate(test_apps)]
        results = self.stress_utils.concurrent_api_calls(self.api_utils.create_tunnel, args_list, num_threads=5)
        
        # 统计结果
        success_count = sum(1 for r in results if r["success"])
        total_count = len(results)
        
        print(f"✓ 压力测试完成: {success_count}/{total_count} 成功")
        
        # 清理测试数据
        for app in test_apps:
            try:
                self.api_utils.delete_tunnel(app, self.test_version)
            except:
                pass
        
        # 验证成功率
        success_rate = success_count / total_count if total_count > 0 else 0
        self.assertGreaterEqual(success_rate, 0.8, f"成功率过低: {success_rate:.2%}")
    
    def test_api_response_time(self):
        """测试API响应时间 - 压力场景"""
        print("\n=== 测试API响应时间 - 压力场景 ===")
        
        # 测试健康检查响应时间
        health_stats = self.stress_utils.measure_response_time(self.api_utils.health_check, num_runs=50)
        print(f"健康检查响应时间: 平均={health_stats['avg_time']:.2f}ms, 成功率={health_stats['success_rate']:.2%}")
        
        # 测试列出隧道响应时间
        list_stats = self.stress_utils.measure_response_time(self.api_utils.list_tunnels, num_runs=50)
        print(f"列出隧道响应时间: 平均={list_stats['avg_time']:.2f}ms, 成功率={list_stats['success_rate']:.2%}")
        
        # 验证响应时间
        self.assertLess(health_stats['avg_time'], 1000, f"健康检查响应时间过长: {health_stats['avg_time']:.2f}ms")
        self.assertLess(list_stats['avg_time'], 1000, f"列出隧道响应时间过长: {list_stats['avg_time']:.2f}ms")
        
        # 验证成功率
        self.assertGreaterEqual(health_stats['success_rate'], 0.95, f"健康检查成功率过低: {health_stats['success_rate']:.2%}")
        self.assertGreaterEqual(list_stats['success_rate'], 0.95, f"列出隧道成功率过低: {list_stats['success_rate']:.2%}")

def run_rest_api_tests():
    """运行REST API测试"""
    print("开始运行REST API测试...")
    
    # 创建测试套件
    suite = unittest.TestLoader().loadTestsFromTestCase(TestRestAPI)
    
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
    
    parser = argparse.ArgumentParser(description="REST API测试")
    parser.add_argument("--test", help="指定要运行的测试方法，例如: test_health_check_normal")
    parser.add_argument("--list", action="store_true", help="列出所有支持的测试用例名")
    
    args = parser.parse_args()
    
    if args.list:
        # 列出所有测试用例名
        test_case = TestRestAPI()
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
        test_case = TestRestAPI()
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
        run_rest_api_tests()