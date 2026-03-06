import requests
import json
import subprocess
import time
import threading
from typing import Dict, List, Optional, Union
import logging
import concurrent.futures
import random

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class TunnelTestUtils:
    """隧道测试工具类"""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url
        self.session = requests.Session()
        self.session.headers.update({
            "Content-Type": "application/json",
            "Accept": "application/json"
        })
    
    def create_tunnel(self, app_name: str, version: str = "v1.0", port: int = 8080) -> Dict:
        """创建隧道"""
        url = f"{self.base_url}/tunnel-starter/api/v1/tunnels"
        data = {
            "app": app_name,
            "version": version,
            "port": port
        }
        
        try:
            response = self.session.post(url, json=data)
            # 仅将500-599状态码视为错误
            if 500 <= response.status_code < 600:
                logger.error(f"创建隧道失败: 服务器错误 {response.status_code}")
                return {"error": f"Server error: {response.status_code}"}
            return response.json()
        except requests.exceptions.RequestException as e:
            logger.error(f"创建隧道失败: {e}")
            return {"error": str(e)}
    
    def delete_tunnel(self, app_name: str, version: str = "v1.0") -> Dict:
        """删除隧道"""
        url = f"{self.base_url}/tunnel-starter/api/v1/tunnels/{app_name}/{version}"
        
        try:
            response = self.session.delete(url)
            # 仅将500-599状态码视为错误
            if 500 <= response.status_code < 600:
                logger.error(f"删除隧道失败: 服务器错误 {response.status_code}")
                return {"error": f"Server error: {response.status_code}"}
            return response.json()
        except requests.exceptions.RequestException as e:
            logger.error(f"删除隧道失败: {e}")
            return {"error": str(e)}
    
    def list_tunnels(self) -> List[Dict]:
        """列出所有隧道"""
        url = f"{self.base_url}/tunnel-starter/api/v1/tunnels"
        
        try:
            response = self.session.get(url)
            # 仅将500-599状态码视为错误
            if 500 <= response.status_code < 600:
                logger.error(f"列出隧道失败: 服务器错误 {response.status_code}")
                return []
            return response.json()
        except requests.exceptions.RequestException as e:
            logger.error(f"列出隧道失败: {e}")
            return []
    
    def get_tunnel_info(self, app_name: str) -> Dict:
        """获取隧道信息"""
        url = f"{self.base_url}/tunnel-starter/api/v1/tunnels/{app_name}"
        
        try:
            response = self.session.get(url)
            # 仅将500-599状态码视为错误
            if 500 <= response.status_code < 600:
                logger.error(f"获取隧道信息失败: 服务器错误 {response.status_code}")
                return {"error": f"Server error: {response.status_code}"}
            return response.json()
        except requests.exceptions.RequestException as e:
            logger.error(f"获取隧道信息失败: {e}")
            return {"error": str(e)}
    
    def health_check(self) -> str:
        """健康检查"""
        url = f"{self.base_url}/health"
        
        try:
            response = self.session.get(url)
            # 仅将500-599状态码视为错误
            if 500 <= response.status_code < 600:
                logger.error(f"健康检查失败: 服务器错误 {response.status_code}")
                return f"error: Server error: {response.status_code}"
            return response.text
        except requests.exceptions.RequestException as e:
            logger.error(f"健康检查失败: {e}")
            return f"error: {str(e)}"

class CommandLineTestUtils:
    """命令行测试工具类"""
    
    def __init__(self, executable_path: str = "./tunnel-starter"):
        self.executable_path = executable_path
    
    def run_command(self, args: List[str], timeout: int = 30) -> Dict:
        """运行命令行命令"""
        try:
            cmd = [self.executable_path] + args
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=timeout,
                encoding='utf-8',
                errors='replace'
            )
            
            return {
                "returncode": result.returncode,
                "stdout": result.stdout,
                "stderr": result.stderr,
                "success": result.returncode == 0
            }
        except subprocess.TimeoutExpired:
            return {
                "returncode": -1,
                "stdout": "",
                "stderr": "Command timed out",
                "success": False
            }
        except Exception as e:
            return {
                "returncode": -1,
                "stdout": "",
                "stderr": str(e),
                "success": False
            }
    
    def start_tunnel(self, app_name: str, version: str = "v1.0", port: int = 8080) -> Dict:
        """启动隧道"""
        args = ["start", "--app", app_name]
        if version:
            args.extend(["--version", version])
        if port:
            args.extend(["--port", str(port)])
        
        return self.run_command(args)
    
    def stop_tunnel(self, app_name: str, version: str = "v1.0") -> Dict:
        """停止隧道"""
        args = ["stop", "--app", app_name]
        if version:
            args.extend(["--version", version])
        
        return self.run_command(args)
    
    def list_tunnels(self, app_name: str = "", version: str = "") -> Dict:
        """列出隧道"""
        args = ["list"]
        if app_name:
            args.extend(["--app", app_name])
        if version:
            args.extend(["--version", version])
        
        return self.run_command(args)

class StressTestUtils:
    """压力测试工具类"""
    
    def __init__(self, api_utils: TunnelTestUtils, cli_utils: CommandLineTestUtils):
        self.api_utils = api_utils
        self.cli_utils = cli_utils
    
    def concurrent_api_calls(self, func, args_list: List[tuple], num_threads: int = 10) -> List[Dict]:
        """并发API调用测试"""
        results = []
        
        with concurrent.futures.ThreadPoolExecutor(max_workers=num_threads) as executor:
            future_to_args = {
                executor.submit(func, *args): args 
                for args in args_list
            }
            
            for future in concurrent.futures.as_completed(future_to_args):
                args = future_to_args[future]
                try:
                    result = future.result()
                    results.append({
                        "args": args,
                        "result": result,
                        "success": "error" not in str(result)
                    })
                except Exception as e:
                    results.append({
                        "args": args,
                        "result": str(e),
                        "success": False
                    })
        
        return results
    
    def generate_random_apps(self, count: int) -> List[str]:
        """生成随机应用名称"""
        return [f"test-app-{random.randint(1000, 9999)}" for _ in range(count)]
    
    def measure_response_time(self, func, *args, num_runs: int = 100) -> Dict:
        """测量响应时间"""
        times = []
        success_count = 0
        
        for _ in range(num_runs):
            start_time = time.time()
            try:
                result = func(*args)
                end_time = time.time()
                response_time = (end_time - start_time) * 1000  # 转换为毫秒
                times.append(response_time)
                
                if "error" not in str(result):
                    success_count += 1
            except Exception:
                times.append(float('inf'))
        
        if times:
            return {
                "min_time": min(times),
                "max_time": max(times),
                "avg_time": sum(times) / len(times),
                "success_rate": success_count / num_runs,
                "total_runs": num_runs
            }
        return {
            "min_time": 0,
            "max_time": 0,
            "avg_time": 0,
            "success_rate": 0,
            "total_runs": 0
        }