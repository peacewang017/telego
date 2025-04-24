# Telego
All in one multi-cluster manager, support quick-config, offline install, devops, cloud-deployment, access control.

Telego 是一个用于管理远程服务器的工具。

## 测试

### 一键运行所有测试
```bash
# 确保已安装 Python3 和 Docker
python3 test_all.py
```
此脚本会自动检测环境，如果不是在 Docker 中运行，会自动创建 Docker 环境并运行所有测试。

### 单独运行测试

项目包含三个测试用例，用于验证基本功能：

#### 测试1：构建测试
```bash
# 方式1：进入测试目录运行
cd test/test1_build
go run main.go

# 方式2：从任意目录运行
go run test/test1_build/main.go
```
此测试验证项目是否可以成功构建。

#### 测试2：构建和快捷方式测试
```bash
# 方式1：进入测试目录运行
cd test/test2_build_and_run_shortcut
go run main.go

# 方式2：从任意目录运行
go run test/test2_build_and_run_shortcut/main.go
```
此测试验证：
1. 项目构建
2. 创建快捷方式
3. 通过快捷方式运行命令

#### 测试3：主节点配置测试
```bash
# 方式1：进入测试目录运行
cd test/test3_main_node_config
go run main.go

# 方式2：从任意目录运行
go run test/test3_main_node_config/main.go
```
此测试验证：
1. 项目构建
2. 主节点配置的读写功能
   - SSH 公钥读写
   - SSH 私钥读写

## 注意事项

1. 运行测试前请确保：
   - Go 环境已正确配置
   - 项目依赖已安装
   - 主节点配置已正确设置
   - 如果使用一键测试脚本，需要安装 Python3 和 Docker

2. 测试3需要主节点配置正确，否则会失败

3. 测试2在 Windows 系统上可能无法创建符号链接，建议在 Linux 或 macOS 上运行

4. 一键测试脚本会自动处理环境问题，建议使用此方式运行测试
