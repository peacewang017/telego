import os

CUR_F_DIR=os.path.dirname(os.path.abspath(__file__))
# telego_src_dir=os.path.join(CUR_F_DIR,"src")
os.chdir(CUR_F_DIR)

# whereis bash
bash_path=os.popen("whereis bash").read().split()[1]

with open("../1.run.sh") as f:
    content=f.read()
    prjdir=os.path.dirname(CUR_F_DIR)
    content=f"cd {prjdir}\n"+content
    content=f"#!{bash_path}\n"+content
with open("telego","w") as f:
    f.write(content)
os.system("chmod +x telego")
os.system("sudo mv telego /usr/local/bin/")

print("Updated telego dev, just type telego in command")