import os
import platform
import subprocess
import requests
import tarfile
import shutil

os.chdir(os.path.abspath(os.path.dirname(__file__)))

def download_and_install_go(version):
    # 1. 获取操作系统和架构信息
    system = platform.system().lower()
    arch = platform.machine().lower()

    # 映射架构到Go支持的格式
    arch_map = {
        "x86_64": "amd64",
        "aarch64": "arm64",
        "arm64": "arm64",
    }
    arch = arch_map.get(arch, arch)

    if system not in ["linux", "darwin", "windows"]:
        raise ValueError(f"Unsupported OS: {system}")

    if arch not in ["amd64", "arm64"]:
        raise ValueError(f"Unsupported architecture: {arch}")

    # 2. 构造下载 URL
    filename = f"go{version}.{system}-{arch}.tar.gz"
    url = f"https://go.dev/dl/{filename}"

    print(f"Downloading Go {version} from {url}...")

    # 3. 下载文件
    if not os.path.exists(filename):
        # -L 重定向；-O 下载文件 -# 进度条
        os.system(f"curl -L -O -# {url}")
        os.system(f"ls")
        if not os.path.exists(filename):
            print(f"Download failed {filename}")
            exit(1)
        print(f"Downloaded {filename}")
    else:
        print(f"Already exist {filename}")

    # 4. 解压并安装
    install_dir = "/opt" if system != "windows" else os.path.expanduser("~\\go")
    go_path = os.path.join(install_dir, "go")

    if os.path.exists(go_path):
        print("Removing existing Go installation...")
        shutil.rmtree(go_path)

    os.system(f"mkdir -p {install_dir}")

    os.system(f"tar -xzvf go{version}.linux-amd64.tar.gz -C /opt")

    def add_env_vars_if_not_present():
        config_file = os.path.expanduser("~/.bashrc")  # 对于 bash 用户，或者使用 .zshrc 等根据你的 shell 类型调整
        # 需要添加的内容
        env_vars = [
            'export GOROOT=/opt/go',  # Go 安装目录
            'export GOPATH=$HOME/go',       # Go 工作目录
            'export PATH=$PATH:$GOROOT/bin:$GOPATH/bin'  # Go 可执行文件添加到 PATH
        ]
        with open(config_file, 'r') as file:
            lines = file.readlines()
        with open(config_file, 'a') as file:
            # 确保每个变量只出现一次
            for env_var in env_vars:
                if not any(env_var in line for line in lines):
                    file.write(env_var + "\n")
                    print(f"Added: {env_var}")
                else:
                    print(f"Already exists: {env_var}")
    # 执行函数
    add_env_vars_if_not_present()

    # 6. 验证安装
    try:
        result = subprocess.run(["go", "version"], capture_output=True, text=True, check=True)
        print(f"Go version: {result.stdout.strip()}")
    except Exception as e:
        print("Failed to verify Go installation:", e)

    print("Go installed successfully.")
    # 清理
    # os.remove(filename)
        # 配置文件路径

    

if __name__ == "__main__":
    try:
        go_version = input("Enter Go version to install (default 1.23.4): ").strip() or "1.23.4"
        download_and_install_go(go_version)
    except Exception as e:
        print("Error:", e)
