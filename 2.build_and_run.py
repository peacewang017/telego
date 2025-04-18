import os
import sys
import subprocess

def os_system(command):
    """
    执行系统命令，支持字符串命令或命令数组
    - 当提供字符串时：使用os.system执行
    - 当提供数组时：使用subprocess.run执行，避免命令注入
    """
    print("="*50)
    
    # 检查command是字符串还是数组
    if isinstance(command, str):
        # 对于字符串命令，使用原来的os.system
        print(f"执行命令：{command}")
        code = os.system(command)
        return code
    elif isinstance(command, (list, tuple)):
        # 对于数组命令，使用subprocess.run
        # cmd_str = " ".join(str(arg) for arg in command)
        print(f"执行命令：{command} (使用subprocess)")
        try:
            result = subprocess.run(command, check=False, text=True, capture_output=True)
            if result.stdout:
                print(result.stdout)
            if result.stderr:
                print("错误输出:", result.stderr)
            return result.returncode
        except Exception as e:
            print(f"执行命令时出错: {e}")
            return 1
    else:
        print(f"不支持的命令类型: {type(command)}")
        return 1

# chdir to the directory of the script
real_path = os.path.dirname(os.path.realpath(__file__))
print(f"chdir to {real_path} (resolving symlinks)")
os.chdir(real_path)

# run the script
os_system("python3 1.build.py")

# get cursys
cursys = ""
if os.name == "nt":
    cursys = "windows"
else:
    cursys = "linux"

# get curarch
curarch = ""
if os.uname().machine == "x86_64":
    curarch = "amd64"
else:
    curarch = "arm64"

# 获取所有命令行参数
cmd_args = sys.argv[1:] 

# 构建命令数组
cmd=[]
if cursys == "windows":
    cmd = [f"./dist/telego_{cursys}_{curarch}.exe"] + cmd_args
else:
    cmd = [f"./dist/telego_{cursys}_{curarch}"] + cmd_args

# 使用改进的os_system函数执行命令
os_system(cmd)
