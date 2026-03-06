#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
隧道启动器完整测试套件
运行所有REST API和命令行接口测试，包括正常场景、异常场景和压力场景
"""

import sys
import os
import json
import time
import argparse
from datetime import datetime
from typing import Dict, List, Any

# 设置UTF-8编码，确保在Windows下正确显示中文
if sys.platform.startswith('win'):
    sys.stdout.reconfigure(encoding='utf-8')
    sys.stderr.reconfigure(encoding='utf-8')

# 添加当前目录到Python路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from test_rest_api import run_rest_api_tests
from test_cli import run_cli_tests

class TestRunner:
    """测试运行器"""
    
    def __init__(self):
        self.results = {
            "timestamp": datetime.now().isoformat(),
            "total_tests": 0,
            "passed_tests": 0,
            "failed_tests": 0,
            "error_tests": 0,
            "test_suites": {},
            "summary": ""
        }
    
    def run_tests(self, test_type: str = "all") -> Dict[str, Any]:
        """运行测试"""
        print("=" * 60)
        print("隧道启动器测试套件")
        print("=" * 60)
        print(f"测试开始时间: {self.results['timestamp']}")
        print(f"测试类型: {test_type}")
        print("-" * 60)
        
        # 运行REST API测试
        if test_type in ["all", "api"]:
            print("\n🚀 开始运行REST API测试...")
            api_start_time = time.time()
            api_success = run_rest_api_tests()
            api_duration = time.time() - api_start_time
            
            self.results["test_suites"]["rest_api"] = {
                "success": api_success,
                "duration": api_duration,
                "timestamp": datetime.now().isoformat()
            }
            
            if api_success:
                self.results["passed_tests"] += 1
                print("✅ REST API测试通过")
            else:
                self.results["failed_tests"] += 1
                print("❌ REST API测试失败")
        
        # 运行命令行测试
        if test_type in ["all", "cli"]:
            print("\n🚀 开始运行命令行接口测试...")
            cli_start_time = time.time()
            cli_success = run_cli_tests()
            cli_duration = time.time() - cli_start_time
            
            self.results["test_suites"]["command_line"] = {
                "success": cli_success,
                "duration": cli_duration,
                "timestamp": datetime.now().isoformat()
            }
            
            if cli_success:
                self.results["passed_tests"] += 1
                print("✅ 命令行接口测试通过")
            else:
                self.results["failed_tests"] += 1
                print("❌ 命令行接口测试失败")
        
        # 计算总测试数
        self.results["total_tests"] = len(self.results["test_suites"])
        
        # 生成总结
        self._generate_summary()
        
        return self.results
    
    def _generate_summary(self):
        """生成测试总结"""
        total_duration = sum(suite.get("duration", 0) for suite in self.results["test_suites"].values())
        
        summary_lines = [
            f"测试完成时间: {datetime.now().isoformat()}",
            f"总测试套件数: {self.results['total_tests']}",
            f"通过测试套件: {self.results['passed_tests']}",
            f"失败测试套件: {self.results['failed_tests']}",
            f"错误测试套件: {self.results['error_tests']}",
            f"总测试时长: {total_duration:.2f}秒",
            f"成功率: {self.results['passed_tests'] / max(self.results['total_tests'], 1) * 100:.1f}%"
        ]
        
        self.results["summary"] = "\n".join(summary_lines)
    
    def save_results(self, output_file: str = "test_results.json"):
        """保存测试结果到文件"""
        try:
            with open(output_file, 'w', encoding='utf-8') as f:
                json.dump(self.results, f, indent=2, ensure_ascii=False)
            print(f"\n📄 测试结果已保存到: {output_file}")
        except Exception as e:
            print(f"\n❌ 保存测试结果失败: {e}")
    
    def print_summary(self):
        """打印测试总结"""
        print("\n" + "=" * 60)
        print("测试总结")
        print("=" * 60)
        print(self.results["summary"])
        
        # 打印各测试套件详情
        for suite_name, suite_result in self.results["test_suites"].items():
            status = "✅ 通过" if suite_result["success"] else "❌ 失败"
            print(f"\n{suite_name.upper()}: {status}")
            print(f"  耗时: {suite_result['duration']:.2f}秒")
            print(f"  完成时间: {suite_result['timestamp']}")
        
        print("\n" + "=" * 60)

def main():
    """主函数"""
    parser = argparse.ArgumentParser(description="隧道启动器测试套件")
    parser.add_argument(
        "--type", 
        choices=["all", "api", "cli"],
        default="all",
        help="测试类型: all(全部), api(REST API), cli(命令行)"
    )
    parser.add_argument(
        "--output",
        default="test_results.json",
        help="测试结果输出文件路径"
    )
    parser.add_argument(
        "--no-save",
        action="store_true",
        help="不保存测试结果到文件"
    )
    
    args = parser.parse_args()
    
    # 运行测试
    runner = TestRunner()
    results = runner.run_tests(args.type)
    
    # 打印总结
    runner.print_summary()
    
    # 保存结果
    if not args.no_save:
        runner.save_results(args.output)
    
    # 返回适当的退出码
    if results["failed_tests"] == 0 and results["error_tests"] == 0:
        print("\n🎉 所有测试通过!")
        return 0
    else:
        print(f"\n⚠️  有 {results['failed_tests'] + results['error_tests']} 个测试套件失败")
        return 1

if __name__ == "__main__":
    sys.exit(main())