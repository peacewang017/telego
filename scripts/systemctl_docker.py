import os

sudo=""
# if not root, use sudo
if os.geteuid() != 0:
    sudo="sudo "

current_dir = os.path.dirname(os.path.abspath(__file__))
mock_systemctl_path = os.path.join(current_dir, "mock_systemctl")
target_path = "/usr/bin/systemctl"

# 备份原始 systemctl 如果存在的话
backup_command = f"{sudo}mv {target_path} {target_path}.backup 2>/dev/null || true"
os.system(backup_command)

# 创建符号链接
link_command = f"{sudo}ln -sf {mock_systemctl_path} {target_path}"
print(f"Creating symlink: {link_command}")
result = os.system(link_command)
if result == 0:
    print(f"成功: mock_systemctl 已链接到 {target_path}")
else:
    print(f"错误: 创建符号链接失败，退出代码 {result}") 


print(f"设置权限: {target_path}")
result2 = os.system(f"{sudo}chmod 700 {target_path}")
if result2 == 0:
    print(f"成功: 设置权限成功")
else:
    print(f"错误: 设置权限失败，退出代码 {result2}")

print(f"测试运行")
result3 = os.system(f"{sudo}systemctl status docker")
if result3 == 0:
    print(f"成功: 测试运行成功")
else:
    print(f"错误: 测试运行失败，退出代码 {result3}")
