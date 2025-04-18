import os

def os_system(command):
    print("="*50)
    print(f"执行命令：{command}")
    os.system(command)

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

if cursys == "windows":
    os_system(f"./dist/telego_{cursys}_{curarch}.exe")
else:
    os_system(f"./dist/telego_{cursys}_{curarch}")
