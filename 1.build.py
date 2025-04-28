import os
import hashlib
import subprocess
import yaml
import random
import sys

# 预声明全局变量
HOST_PROJECT_DIR = None
PROJECT_ROOT = None

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

# 切换到当前脚本所在目录
curfdir = os.path.dirname(os.path.abspath(__file__))
os.chdir(curfdir)

# debug current path and filelist
print(f"当前路径：{curfdir}")
print(f"当前目录下的文件：")
for f in os.listdir(curfdir):
    print(f"- {f}")

with open("compile_conf.yml",encoding="utf-8") as f:
    conf=yaml.safe_load(f)
main_node_ip=conf["main_node_ip"]
main_node_user=conf["main_node_user"]
# 读取代理配置，如果不存在则为None
proxy_config = conf.get("proxy", None)

def cmd_color(string,color):
    color_dict={"red":31,"green":32,"yellow":33,"blue":34,"magenta":35,"cyan":36,"white":37}
    return f"\033[{color_dict[color]}m{string}\033[0m"

def run_command(command, check=True):
    BEFORE_RUN_TITLE=cmd_color("执行命令：","blue")
    RUN_FAIL_TITLE=cmd_color(">","blue")+"\n"+cmd_color("命令执行失败：","red")
    RUN_SUCCESS_TITLE=cmd_color(">","blue")+"\n"+cmd_color("命令执行成功：","green")
    
    print(f"{BEFORE_RUN_TITLE}{command}")
    try:
        result = subprocess.run(command, shell=True, check=check, text=True, capture_output=True)
        print(result.stdout)
        if result.stderr:
            print(result.stderr)
        if check and result.returncode != 0:
            print(f"{RUN_FAIL_TITLE}{command}")
            print(f"错误输出：\n{result.stderr}")
            sys.exit(1)
        print(f"{RUN_SUCCESS_TITLE}{command}\n")
        return result.returncode == 0
    except subprocess.CalledProcessError as e:
        print(f"{RUN_FAIL_TITLE}{command}")
        print(f"错误输出：\n{e.stderr}")
        if check:
            sys.exit(1)
        return False

def os_system_sure(command):
    return run_command(command, check=True)

def os_system(command):
    return run_command(command, check=False)

# 启动时检查是否为docker ubuntu18.04
def check_ubuntu_version():
    os_release_file = "/etc/os-release"
    
    if not os.path.exists(os_release_file):
        print("无法找到 /etc/os-release 文件，可能不是基于 Linux 的系统。")
        return False
    try:
        with open(os_release_file, 'r') as f:
            lines = f.readlines()

        os_info = {}
        for line in lines:
            key, _, value = line.strip().partition("=")
            os_info[key] = value.strip('"')

        if os_info.get("NAME") == "Ubuntu" and os_info.get("VERSION_ID") == "18.04":
            print("当前系统是 Ubuntu 18.04。")
            return True
        else:
            print(f"当前系统不是 Ubuntu 18.04，而是 {os_info.get('NAME')} {os_info.get('VERSION_ID', '未知版本')}。")
            return False
    except:
        print("无法打开 /etc/os-release 文件。")
        return False

if not check_ubuntu_version():
    os.system("python3 gen_menu.py")
    
    print("将启动 ubuntu 18.04进行编译")
    os.system("mkdir -p build")

    def _recover_proxy():
        pass
    recover_proxy=_recover_proxy

    # PROXY_ENV=""
    if os.environ.get("http_proxy"):  
        PROXY_ADDR=os.environ['http_proxy']
        if PROXY_ADDR.find("127.0.0.1")!=-1: 
            del os.environ['http_proxy']
            del os.environ['https_proxy']
            def r():
                os.environ['http_proxy']=PROXY_ADDR
                os.environ['https_proxy']=PROXY_ADDR
            recover_proxy=r
            
    os_system_sure("docker pull ubuntu:18.04")

    # 创建dockerfile
    with open("build/Dockerfile", "w") as f:
        f.write(f"""
FROM ubuntu:18.04
                
RUN apt-get update
RUN apt-get install -y git
RUN apt-get install -y curl 
RUN apt-get install -y wget 
RUN apt-get install -y python3 
RUN apt-get install -y python3-yaml 
RUN apt-get clean

RUN wget https://go.dev/dl/go1.23.2.linux-amd64.tar.gz
RUN rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go1.23.2.linux-amd64.tar.gz \
    && rm -rf go1.23.2.linux-amd64.tar.gz

ENV PATH="/usr/local/go/bin:${{PATH}}"
ENV PYTHONIOENCODING=utf-8

CMD ["bash"]
""")
    
    os.chdir("build")
    flist=os.listdir("../")
    blacklist=[".cache",".git",".dist","build","telego_prjs","teleyard-template"]
    for f in flist:
        if f in blacklist:
            continue
        os.system(f"cp -r ../{f} .")
    def check_image_exists(image_name):
        result = subprocess.run(['docker', 'images', '-q', image_name], stdout=subprocess.PIPE)
        return len(result.stdout.decode('utf-8').strip()) > 0
    if not check_image_exists("telego_build"):
        # os_system_sure("docker pull ubuntu:18.04")
        os_system_sure("docker build -t telego_build .")
    recover_proxy()
    os.chdir("..")
    os.system("rm -rf build")
    
    # run: mount HOST_PROJECT_DIR to /telego, workdir is /telego
    print("building telego at:", HOST_PROJECT_DIR)
    print(f"files in {HOST_PROJECT_DIR}:")
    for f in os.listdir(HOST_PROJECT_DIR):
        print(f"- {f}")

    randname="telego_build_"+str(random.randint(10000000,99999999))
    os_system_sure(f"docker run --name {randname} -v {HOST_PROJECT_DIR}:/telego -w /telego telego_build bash -c 'python3 1.build.py -- privilege'")
    # remove container
    os_system_sure(f"docker rm -f {randname}")
    # end of call the docker
    exit(0)


print("正在容器 ubuntu 18.04 中编译")

# 定义源文件和输出目录
OUTPUT_DIR = "./dist"
CHECKSUM_FILE = "./dist/checksums.txt"

# 创建输出目录
if not os.path.exists(OUTPUT_DIR):
    os.makedirs(OUTPUT_DIR)

# 交叉编译目标平台
targets = [
    ("windows", "amd64"),
    # ("windows", "386"),
    ("linux", "amd64"),
    ("linux", "arm64"),
    # ("darwin", "arm64"),
    # ("darwin", "arm64")
]

# 用于存储校验码的字典
checksums = {}

# 遍历目标平台并进行编译
for goos, goarch in targets:
    output_file = os.path.join(OUTPUT_DIR, f"telego_{goos}_{goarch}")
    if goos == "windows":
        output_file += ".exe"
    
    print(f"Building for {goos}/{goarch}...")
    
    # 构建命令
    cmd = [
        "go", "build",
        "-o", output_file,
        "-buildvcs=false",
        "."
    ]
    
    # 设置环境变量
    env = os.environ.copy()
    env["GOOS"] = goos
    env["GOARCH"] = goarch
    env["GOMODCACHE"] = "/telego/.cache/gomodcache"
    env["GOCACHE"] = "/telego/.cache/gocache"
    
    # 如果配置了代理，设置Go编译环境代理
    if proxy_config:
        env["http_proxy"] = proxy_config
        env["https_proxy"] = proxy_config
        # env["GOPROXY"] = "https://goproxy.cn,direct"  # 使用中国区Go代理
    
    # 执行命令
    result = subprocess.run(cmd, env=env)
    
    if result.returncode == 0:
        print(f"Successfully built {output_file}")
        
        # 计算校验码
        with open(output_file, "rb") as f:
            sha256_hash = hashlib.sha256(f.read()).hexdigest()
        
        checksums[output_file] = sha256_hash
        print(f"SHA-256 checksum for {output_file}: {sha256_hash}")
    else:
        print(f"Failed to build {output_file}")
        print(f"Error: {result.stderr}")

# 将校验码保存到文件
with open(CHECKSUM_FILE, "w") as f:
    for output_file, checksum in checksums.items():
        output_file=os.path.basename(output_file)
        f.write(f"{output_file}  {checksum}\n")

print(f"Checksums saved to {CHECKSUM_FILE}")



with open("dist/install.py", 'w') as f:
    f.write(f'FILESERVER="http://{main_node_ip}:8003/bin_telego"'+"""
            
import os, tempfile, urllib.request
# if is windows, download to tempdir, move to C:\System32\telego
if os.name == 'nt':
    tempdir=os.path.expanduser("~")+"\\\\telego_install"
    try:
        os.makedirs(tempdir)
    except:
        pass
    os.system(f"curl -L {FILESERVER}/telego_windows_amd64.exe -o {tempdir}\\\\telego.exe")
    os.system(f"move {tempdir}\\\\telego.exe C:\\\\Windows\\\\System32\\\\telego.exe")
else:
    import pwd
    tempdir=tempfile.mkdtemp()
    # check root
    prefix=""
    if os.geteuid() != 0:
        prefix="sudo "
            
    ARCH=""
    if os.uname().machine == "aarch64":
        ARCH="arm64"
    if os.uname().machine == "x86_64":
        ARCH="amd64"
    if os.uname().machine == "arm64":
        ARCH="arm64"
    os.chdir(tempdir)
    urllib.request.urlretrieve(f"{FILESERVER}/telego_linux_{ARCH}", "telego")
    os.system("chmod +x telego")
    curuser=pwd.getpwuid(os.getuid()).pw_name
    os.system(f"{prefix}mv telego /usr/bin/telego")
    os.system(f"{prefix}chown {curuser} /usr/bin/telego")
    
""")
    
with open("dist/install.sh", "w") as f:
    f.write(f"""
#!/bin/bash
FILESERVER="http://{main_node_ip}:8003/bin_telego"
"""+"""
# Check OS type
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    ARCH=""
    MACHINE_TYPE=$(uname -m)

    if [ "$MACHINE_TYPE" == "aarch64" ]; then
        ARCH="arm64"
    elif [ "$MACHINE_TYPE" == "x86_64" ]; then
        ARCH="amd64"
    fi

    TEMP_DIR=$(mktemp -d)

    # Check if root
    if [ $(id -u) -ne 0 ]; then
        PREFIX="sudo"
    else
        PREFIX=""
    fi

    # Download and install telego
    cd $TEMP_DIR
    curl -L "$FILESERVER/telego_linux_${ARCH}" -o telego
    chmod +x telego

    CUR_USER=$(whoami)
    $PREFIX mv telego /usr/bin/telego
    $PREFIX chown $CUR_USER /usr/bin/telego
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    TEMP_DIR="$HOME\\telego_install"

    # Create temporary directory
    mkdir -p $TEMP_DIR

    # Download telego for Windows
    curl -L "$FILESERVER/telego_windows_amd64.exe" -o "$TEMP_DIR\\telego.exe"

    # Move to System32
    mv "$TEMP_DIR\\telego.exe" "C:\\Windows\\System32\\telego.exe"
else
    echo "Unsupported OS"
    exit 1
fi

""")
    
with open("dist/install.sh", "w") as f:
    f.write(f"""
#!/bin/bash
FILESERVER="http://{main_node_ip}:8003/bin_telego"
"""+"""
# Check OS type
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    ARCH=""
    MACHINE_TYPE=$(uname -m)

    if [ "$MACHINE_TYPE" == "aarch64" ]; then
        ARCH="arm64"
    elif [ "$MACHINE_TYPE" == "x86_64" ]; then
        ARCH="amd64"
    fi

    TEMP_DIR=$(mktemp -d)

    # Check if root
    if [ $(id -u) -ne 0 ]; then
        PREFIX="sudo"
    else
        PREFIX=""
    fi

    # Download and install telego
    cd $TEMP_DIR
    curl -L "$FILESERVER/telego_linux_${ARCH}" -o telego
    chmod +x telego

    CUR_USER=$(whoami)
    $PREFIX mv telego /usr/bin/telego
    $PREFIX chown $CUR_USER /usr/bin/telego
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    TEMP_DIR="$HOME\\telego_install"

    # Create temporary directory
    mkdir -p $TEMP_DIR

    # Download telego for Windows
    curl -L "$FILESERVER/telego_windows_amd64.exe" -o "$TEMP_DIR\\telego.exe"

    # Move to System32
    mv "$TEMP_DIR\\telego.exe" "C:\\Windows\\System32\\telego.exe"
else
    echo "Unsupported OS"
    exit 1
fi

""")

def load_compile_conf():
    try:
        with open('compile_conf.yml', 'r') as f:
            return yaml.safe_load(f)
    except FileNotFoundError:
        return {}

def set_proxy_from_conf():
    if not os.environ.get("http_proxy"):
        conf = load_compile_conf()
        if conf and 'proxy' in conf:
            proxy = conf['proxy']
            os.environ['http_proxy'] = proxy
            os.environ['https_proxy'] = proxy
            return True
    return False
    # 尝试从配置文件设置代理
proxy_set = set_proxy_from_conf()
