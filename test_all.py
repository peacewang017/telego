#!/usr/bin/env python3

import os
import sys
import subprocess
import platform
import yaml

# 测试分类
TESTS = {
    "direct": [
        "test/test1_build/build_test.go",  # 构建测试
    ],
    "in_docker": [
        "test/test2_build_and_run_shortcut/shortcut_test.go",  # 快捷方式测试
        "test/test3_main_node_config/config_test.go",  # 主节点配置测试
    ]
}

# Docker 相关配置
DOCKER_IMAGE = "telego_test"
DOCKER_BASE_IMAGE = "docker:latest"

def get_arch():
    """获取当前系统架构"""
    machine = platform.machine().lower()
    if machine in ['x86_64', 'amd64']:
        return 'amd64'
    elif machine in ['aarch64', 'arm64']:
        return 'arm64'
    else:
        raise RuntimeError(f"不支持的架构: {machine}")

def get_binary_path():
    """获取当前架构的二进制文件路径"""
    arch = get_arch()
    binary_name = f"telego_linux_{arch}"
    binary_path = os.path.join("dist", binary_name)
    if not os.path.exists(binary_path):
        raise RuntimeError(f"找不到二进制文件: {binary_path}")
    return binary_path

def is_docker():
    """检查是否在 Docker 容器中运行"""
    try:
        with open('/proc/1/cgroup', 'r') as f:
            return 'docker' in f.read()
    except:
        return False

def setup_go_environment():
    """设置 Go 环境变量"""
    # 设置 Go 模块支持
    os.environ["GO111MODULE"] = "on"
    os.environ["GOPATH"] = os.path.join(os.getcwd(), "go")
    os.environ["PATH"] = f"{os.environ['PATH']}:{os.path.join(os.environ['GOPATH'], 'bin')}"
    
    # 创建并设置 Go 缓存目录
    cache_base = ".cache"
    os.makedirs(cache_base, exist_ok=True)
    os.makedirs(os.path.join(cache_base, "gomodcache"), exist_ok=True)
    os.makedirs(os.path.join(cache_base, "gocache"), exist_ok=True)
    
    os.environ["GOMODCACHE"] = os.path.join(os.getcwd(), cache_base, "gomodcache")
    os.environ["GOCACHE"] = os.path.join(os.getcwd(), cache_base, "gocache")
    
    # 读取 compile_conf.yml 中的代理设置
    try:
        with open("compile_conf.yml", "r") as f:
            config = yaml.safe_load(f)
            if "proxy" in config:
                proxy = config["proxy"]
                os.environ["http_proxy"] = proxy
                os.environ["https_proxy"] = proxy
                print(f"已设置代理: {proxy}")
    except Exception as e:
        print(f"读取代理设置失败: {e}")

def run_direct_tests():
    """运行直接在主机上执行的测试"""
    for test in TESTS["direct"]:
        print(f"\n运行测试: {test}")
        cmd = ["go", "test", "-v", test]
        subprocess.run(cmd, check=True)

def run_in_docker():
    """在 Docker 中运行测试"""
    # 先拉取基础镜像
    print(f"\n拉取基础镜像 {DOCKER_BASE_IMAGE}...")
    subprocess.run(["docker", "pull", DOCKER_BASE_IMAGE], check=True)
    
    # 构建支持 Docker-in-Docker 的测试镜像
    dockerfile = f"""
FROM {DOCKER_BASE_IMAGE}

# 安装 Go 编译环境依赖
RUN apk add --no-cache gcc
RUN apk add --no-cache musl-dev
RUN apk add --no-cache go

# 安装 Python 环境
RUN apk add --no-cache python3
RUN apk add --no-cache py3-pip

# 安装 Python 依赖
RUN pip3 install pyyaml
"""
    with open("Dockerfile.test", "w") as f:
        f.write(dockerfile)
    
    try:
        print(f"\n构建测试镜像 {DOCKER_IMAGE}...")
        subprocess.run(["docker", "build", "-t", DOCKER_IMAGE, "-f", "Dockerfile.test", "."], check=True)
        
        # 获取当前目录的绝对路径
        current_dir = os.path.abspath(".")
        
        # 在 Docker 中运行测试
        for test in TESTS["in_docker"]:
            print(f"\n在 Docker 中运行测试: {test}")
            cmd = [
                "docker", "run", "--rm",
                "-v", f"{current_dir}:/telego",
                "-w", "/telego",
                "-v", "/var/run/docker.sock:/var/run/docker.sock",  # 挂载 Docker socket
                DOCKER_IMAGE,
                "go", "test", "-v", test
            ]
            subprocess.run(cmd, check=True)
    finally:
        # 清理临时文件
        if os.path.exists("Dockerfile.test"):
            os.remove("Dockerfile.test")

def run_tests():
    """运行所有测试"""
    print("开始运行测试...")
    
    # 运行直接测试
    print("\n=== 运行直接测试 ===")
    run_direct_tests()
    
    # 运行 Docker 测试
    print("\n=== 运行 Docker 测试 ===")
    run_in_docker()
    
    print("\n所有测试完成！")

if __name__ == "__main__":
    try:
        run_tests()
        print("\n所有测试完成！")
    except subprocess.CalledProcessError as e:
        print(f"\n测试失败: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"\n发生错误: {e}")
        sys.exit(1) 