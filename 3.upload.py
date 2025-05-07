import os,sys    
# chdir to this file's dir
os.chdir(os.path.dirname(os.path.abspath(__file__)))

print("before run this, config main node as rclone 'remote'")

if len(sys.argv)==2:
    if sys.argv[1]=="onlywin":
        print("upload onlywin")
        os.system("rclone copy -P dist remote:/teledeploy/bin_telego/ --exclude telego_linux_amd64 --exclude telego_linux_arm64")    
        exit(0)
    if sys.argv[1]=="onlylinuxamd":
        print("upload onlylinuxamd")
        os.system("rclone copy -P dist remote:/teledeploy/bin_telego/ --exclude telego_windows_amd64.exe --exclude telego_linux_arm64")
        exit(0)

os.system("rclone copy -P dist remote:/teledeploy/bin_telego/")