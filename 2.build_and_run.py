import os

def os_system(command):
    print("="*50)
    print(f"执行命令：{command}")
    os.system(command)

# chdir to the directory of the script
print(f"chdir to {os.path.dirname(os.path.abspath(__file__))}")
os.chdir(os.path.dirname(os.path.abspath(__file__)))

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
