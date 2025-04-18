#!/usr/bin/env python3
import os
import sys
import subprocess

# 可配置变量
script_path = "2.build_and_run.py"
command_name = "telego"

def create_build_shortcut():
    """创建脚本快捷方式，使用准确的Python解释器路径"""
    if not os.path.exists(script_path):
        print(f"Error: Script path '{script_path}' does not exist.")
        return False, None
    
    # 动态生成快捷方式文件名
    basename = os.path.splitext(script_path)[0]
    build_shortcut_path = f"{basename}_shortcut.py"
    
    try:
        # 获取当前Python解释器的路径
        python_interpreter_path = sys.executable
        
        # 读取源脚本的内容
        with open(script_path, 'r') as source_file:
            source_content = source_file.read()
        
        # 创建新脚本文件，使用准确的Python解释器路径
        with open(build_shortcut_path, 'w') as target_file:
            target_file.write(f"#!{python_interpreter_path}\n")
            target_file.write(source_content)
        
        # 设置可执行权限
        os.chmod(build_shortcut_path, 0o755)
        print(f"Created {build_shortcut_path} with interpreter: {python_interpreter_path}")
        return True, build_shortcut_path
    
    except Exception as e:
        print(f"Error creating build shortcut: {e}")
        return False, None

def create_shortcut(script_path, command_name):
    # 检查脚本路径是否存在
    if not os.path.exists(script_path):
        print(f"Error: Script path '{script_path}' does not exist.")
        return

    # 检查是否具有可执行权限
    if not os.access(script_path, os.X_OK):
        print(f"Error: Script '{script_path}' is not executable. Run 'chmod +x {script_path}' to make it executable.")
        return

    # 创建符号链接
    bin_path = f"/usr/local/bin/{command_name}"
    try:
        # 获取脚本的绝对路径，确保符号链接正确
        abs_script_path = os.path.abspath(script_path)
        subprocess.run(['sudo', 'ln', '-s', abs_script_path, bin_path], check=True)
        print(f"Shortcut '{command_name}' created successfully, pointing to {abs_script_path}")
    except subprocess.CalledProcessError as e:
        print(f"Error: Failed to create shortcut. {e}")

if __name__ == "__main__":
    # 首先创建build_shortcut.py
    success, shortcut_path = create_build_shortcut()
    if success and shortcut_path:
        # 然后创建命令行快捷方式，指向新创建的shortcut脚本
        create_shortcut(shortcut_path, command_name)