#!/usr/bin/env python3

import os
import sys
import subprocess
import platform
import yaml
import shutil
import argparse

# 预声明全局变量
HOST_PROJECT_DIR = None
PROJECT_ROOT = None
PROXY = None
# 测试分类
TESTS = {
    "direct": [
       
    ],
    "in_docker": [
        "test/test1_build/build_test.go",  # 构建测试
        "test/test2_build_and_run_shortcut/shortcut_test.go",  # 快捷方式测试
        "test/test3_main_node_config/config_test.go",  # 主节点配置测试
    ]
}

# Docker 相关配置
DOCKER_IMAGE = "telego_test"
DOCKER_BASE_IMAGE = "docker:latest"

def find_project_root():
    """向上查找项目根目录（包含 compile_conf.tmp.yml 的目录）"""
    current_dir = os.path.dirname(os.path.abspath(__file__))
    while True:
        if os.path.exists(os.path.join(current_dir, "compile_conf.tmp.yml")):
            return current_dir
        parent_dir = os.path.dirname(current_dir)
        if parent_dir == current_dir:
            raise RuntimeError("无法找到项目根目录（未找到 compile_conf.tmp.yml）")
        current_dir = parent_dir

# 检查是否在容器内
def is_in_container():
    # 检查 /.dockerenv 文件
    if os.path.exists('/.dockerenv'):
        return True
    
    # 检查 cgroup
    try:
        with open('/proc/1/cgroup', 'r') as f:
            return 'docker' in f.read()
    except:
        return False

# 获取项目根目录
PROJECT_ROOT = find_project_root()
print(f"项目根目录：{PROJECT_ROOT}")

# 处理 build_context.yml
build_context_path = os.path.join(PROJECT_ROOT, "build_context.yml")
if not is_in_container():
    # 在主机上，更新 build_context.yml
    HOST_PROJECT_DIR = PROJECT_ROOT
    with open(build_context_path, "w") as f:
        yaml.dump({
            "HOST_PROJECT_DIR": HOST_PROJECT_DIR
        }, f, default_flow_style=False)
    print(f"已更新 build_context.yml：{build_context_path}")
else:
    # 在容器内，读取 HOST_PROJECT_DIR
    with open(build_context_path, "r") as f:
        build_context = yaml.safe_load(f)
        HOST_PROJECT_DIR = build_context["HOST_PROJECT_DIR"]
        print(f"从 build_context.yml 读取到主机项目目录：{HOST_PROJECT_DIR}")

# 处理 compile_conf.yml
def setup_compile_conf():
    conf_path = "compile_conf.yml"
    if os.path.exists(conf_path):
        # 备份现有文件
        bak_path = "compile_conf.yml.bak"
        shutil.move(conf_path, bak_path)
        print(f"已备份现有配置文件到: {bak_path}")
    
    # 创建新的配置文件
    with open(conf_path, "w") as f:
        f.write("""# check the config discription
# https://qcnoe3hd7k5c.feishu.cn/wiki/MoDOw2fxnidARCkE2hKc60jHn8b?fromScene=spaceOverview
main_node_ip: localhost
main_node_user: abc
main_node_sshport: 2222
image_repo_with_prefix: http://localhost:5000
""")
    print("已创建新的配置文件")

# 加载代理配置
def load_proxy_config():
    try:
        with open("compile_conf.yml", "r") as f:
            config = yaml.safe_load(f)
            return config.get("proxy", "")
    except Exception as e:
        print(f"读取代理设置失败: {e}")
        return ""

# 全局代理配置
PROXY = load_proxy_config()


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
    
    # 设置代理
    if PROXY:
        os.environ["http_proxy"] = PROXY
        os.environ["https_proxy"] = PROXY
        print(f"已设置代理: {PROXY}")

def run_direct_tests():
    """运行直接在主机上执行的测试"""
    for test in TESTS["direct"]:
        print(f"\n运行测试: {test}")
        cmd = ["go", "test", "-v", test]
        subprocess.run(cmd, check=True)

def run_in_docker(only_init=False):
    """在 Docker 中运行测试
    
    Args:
        only_init: 如果为True，只初始化Docker容器而不运行测试
    """
    # 先拉取基础镜像
    print(f"\n拉取基础镜像 {DOCKER_BASE_IMAGE}...")
    result = subprocess.run(["docker", "pull", DOCKER_BASE_IMAGE], capture_output=True, text=True)
    if result.returncode != 0:
        print(f"拉取镜像失败: {result.stderr}")
        raise subprocess.CalledProcessError(result.returncode, result.args, result.stdout, result.stderr)
    print(result.stdout)
    
    # 准备代理环境变量设置
    proxy_env = ""
    if PROXY:
        proxy_env = f"""
# 设置代理环境变量
ENV http_proxy={PROXY}
ENV https_proxy={PROXY}
"""
    
    # 构建支持 Docker-in-Docker 的测试镜像
    dockerfile = f"""
FROM {DOCKER_BASE_IMAGE}
{proxy_env}
# 安装 Go 编译环境依赖
RUN apk add --no-cache gcc
RUN apk add --no-cache musl-dev
RUN apk add --no-cache go

# 安装 Python 环境
RUN apk add --no-cache python3
RUN apk add --no-cache py3-pip
RUN apk add --no-cache py3-yaml
"""
    with open("Dockerfile.test", "w") as f:
        f.write(dockerfile)
    
    try:
        print(f"\n构建测试镜像 {DOCKER_IMAGE}...")
        result = subprocess.run(["docker", "build", "-t", DOCKER_IMAGE, "-f", "Dockerfile.test", "."], 
                              capture_output=True, text=True)
        if result.returncode != 0:
            print(f"构建镜像失败: {result.stderr}")
            raise subprocess.CalledProcessError(result.returncode, result.args, result.stdout, result.stderr)
        print(result.stdout)
    finally:
        # 清理临时文件
        if os.path.exists("Dockerfile.test"):
            os.remove("Dockerfile.test")
    
    # 创建测试脚本
    test_script = """#!/bin/sh
set -e

echo "Running go mod tidy..."
go mod tidy
"""
    if not only_init:
        test_script += "echo 'Running tests...'\n"
        for test in TESTS["in_docker"]:
            test_script += f'go test -v {test}\n'
        
    else:
        test_script += "echo 'Initialized environment...'\n"
        # block forever
        # test_script += "while true; do sleep 1000; done\n"
        test_script += "tail -f /dev/null\n"
    
    with open("docker_test.sh", "w") as f:
        f.write(test_script)
    os.chmod("docker_test.sh", 0o755)

    # 获取当前目录的绝对路径
    current_dir = os.path.abspath(".")
    
    # 在 Docker 中运行测试脚本
    print("\n在 Docker 中运行测试...")
    cmd = [
        "docker", "run", 
        "--network=host",  # 使用宿主机网络
    ]
    
    # 只初始化时在后台运行，并保留容器
    if only_init:
        cmd.extend(["-d", "--name", "telego-container"])
    else:
        cmd.append("--rm")  # 非初始化模式下运行完毕删除容器
    
    # 添加剩余的通用参数
    cmd.extend([
        "-v", f"{current_dir}:/telego",
        "-w", "/telego",
        "-v", "/var/run/docker.sock:/var/run/docker.sock",
        DOCKER_IMAGE,
        "/telego/docker_test.sh"
    ])

    # 执行命令
    if only_init:
        # 在后台启动容器，只返回容器ID，不等待完成
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, check=True)
            container_id = result.stdout.strip()
            print(f"Docker 容器已在后台启动，容器 ID: {container_id}")
            return
        except subprocess.CalledProcessError as e:
            print(f"启动 Docker 容器失败 - 错误代码: {e.returncode}")
            print(f"标准输出:\n{e.stdout}")
            print(f"错误输出:\n{e.stderr}")
            raise
    else:
        # 正常模式：等待容器执行完毕
        result = subprocess.run(cmd, capture_output=True, text=True)
        if result.returncode != 0:
            print(f"测试失败 - 标准输出:\n{result.stdout}")
            print(f"测试失败 - 错误输出:\n{result.stderr}")
            raise subprocess.CalledProcessError(result.returncode, result.args, result.stdout, result.stderr)
        print(result.stdout)
    
    # 清理测试脚本
    if os.path.exists("docker_test.sh"):
        os.remove("docker_test.sh")

def initialization():
    """执行初始化操作"""
    print("初始化环境...")
    
    # 设置配置文件
    setup_compile_conf()
    
    # 运行 gen_menu.py
    print("\n=== 生成菜单配置 ===")
    try:
        subprocess.run(["python3", "gen_menu.py"], check=True)
        print("菜单配置生成成功")
    except subprocess.CalledProcessError as e:
        print(f"生成菜单配置失败: {e}")
        raise Exception(f"生成菜单配置失败: {e}")
    
    print("初始化完成")

def run_tests():
    """运行所有测试"""
    print("开始运行测试...")
    
    # 初始化环境
    initialization()
    
    # 运行直接测试
    print("\n=== 运行直接测试 ===")
    try:
        run_direct_tests()
    except subprocess.CalledProcessError as e:
        print(f"直接测试失败: {e}")
        raise Exception(f"直接测试失败: {e}")
    
    # 运行 Docker 测试
    print("\n=== 运行 Docker 测试 ===")
    try:
        run_in_docker()
    except subprocess.CalledProcessError as e:
        print(f"Docker 测试失败: {e}")
        raise Exception(f"Docker 测试失败: {e}")
    
    print("\n所有测试完成！")

def update_ci_workflow():
    """生成 GitHub Actions 工作流文件"""
    print("正在生成 CI 工作流文件...")
    
    # 读取模板
    template_path = os.path.join(PROJECT_ROOT, '.github/workflows/test.yml.tmp')
    with open(template_path, 'r') as f:
        template_content = f.read()
    
    # 构建测试步骤
    test_steps = [
        {
            "name": "\"Step 1: 初始化环境\"",
            "run": "python test_all.py --only-init 2>&1 | tee test_output_1_init.log"
        }
    ]

    step_index = 2
    
    # 添加直接测试步骤 (根据TESTS["direct"]生成)
    if TESTS["direct"]:
        for i, test in enumerate(TESTS["direct"]):
            test_name = os.path.basename(test).replace("_test.go", "")
            test_steps.append({
                "name": f"\"Step {step_index}: {test_name}\"",
                "run": f"go test -v {test} 2>&1 | tee test_output_{step_index}_{test_name}.log",
                # "continue-on-error": True
            })
            step_index += 1
    
    # 添加Docker测试步骤 (根据TESTS["in_docker"]生成)
    if TESTS["in_docker"]:
        for i, test in enumerate(TESTS["in_docker"]):
            test_name = os.path.basename(test).replace("_test.go", "")
            test_steps.append({
                "name": f"\"Step {step_index}: {test_name}\"",
                "run": f"docker exec telego-container go test ./{test} -v 2>&1 | tee test_output_{step_index}_{test_name}.log",
                # "continue-on-error": True
            })
            step_index += 1
    
    # 生成 YAML 格式的步骤
    steps_yaml = ""
    for step in test_steps:
        steps_yaml += f"      - name: {step['name']}\n"
        steps_yaml += f"        run: |\n"
        steps_yaml += f"          {step['run']}\n"
        if 'continue-on-error' in step and step['continue-on-error']:
            steps_yaml += f"        continue-on-error: true\n"
    
    # 替换模板中的注释部分
    workflow_content = template_content.replace("      # - name: Run tests\n      #   run: |\n      #     set -x  # 启用命令回显\n      #     python test_all.py 2>&1 | tee test_output.log\n      #   continue-on-error: true", steps_yaml)
    
    # 保存工作流文件
    output_path = os.path.join(PROJECT_ROOT, '.github/workflows/test.yml')
    with open(output_path, 'w') as f:
        f.write(workflow_content)
    
    print(f"CI 工作流文件已生成: {output_path}")

def parse_args():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(description='运行测试并更新 CI 配置')
    parser.add_argument('--update-ci', action='store_true', help='更新 CI 工作流文件')
    parser.add_argument('--only-init', action='store_true', help='仅执行初始化步骤')
    return parser.parse_args()

if __name__ == "__main__":
    args = parse_args()
    
    try:
        if args.update_ci:
            update_ci_workflow()
            print("CI 工作流已更新，退出...")
            sys.exit(0)
        
        if args.only_init:
            initialization()
            run_in_docker(only_init=True)
            print("初始化完成，退出...")
            sys.exit(0)
        
        # 默认行为：运行所有测试
        run_tests()
        print("\n所有测试完成！")
    except subprocess.CalledProcessError as e:
        print(f"\n测试失败: {e}")
        raise Exception(f"测试失败: {e}")
    except Exception as e:
        print(f"\n发生错误: {e}")
        raise Exception(f"发生错误: {e}") 