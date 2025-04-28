package app

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"telego/util"
	clusterconf "telego/util/cluster_conf"
	"telego/util/yamlext"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	"k8s.io/client-go/util/homedir"
)

type SshMode = int

const (
	SshModeGenOrGetKey = iota
	SshModeSetupCluster // https://qcnoe3hd7k5c.feishu.cn/wiki/V6eHwZm1aiofeykaSd5cmgPcnSe#share-Hc1hdGT26oI4I0xPaplcEhMundd
	SshModeSetupThisNode
)

type SshJob struct {
	Mode SshMode
}

func (s SshJob) ModeString() string {
	switch s.Mode {
	case SshModeGenOrGetKey:
		return "1.gen_or_get_key"
	case SshModeSetupCluster:
		return "2.setup_cluster"
	case SshModeSetupThisNode:
		return "3.setup_this_node"
	default:
		return "unknown"
	}
}

type ModJobSshStruct struct{}

var ModJobSsh ModJobSshStruct

func (m ModJobSshStruct) JobCmdName() string {
	return "ssh"
}

func (m ModJobSshStruct) ParseJob(applyCmd *cobra.Command) *cobra.Command {
	// 绑定命令行标志到结构体字段
	mode := ""
	applyCmd.Flags().StringVar(&mode, "mode", "", "Sub operation of ssh")

	applyCmd.Run = func(_ *cobra.Command, _ []string) {
		TaskId := 0
		switch mode {
		case SshJob{Mode: SshModeGenOrGetKey}.ModeString():
			TaskId = SshModeGenOrGetKey
		case SshJob{Mode: SshModeSetupCluster}.ModeString():
			TaskId = SshModeSetupCluster
		case SshJob{Mode: SshModeSetupThisNode}.ModeString():
			TaskId = SshModeSetupThisNode
		default:
			fmt.Println(color.RedString("unsupported ssh ope mode: '%s'", mode))
			os.Exit(1)
		}

		ModJobSsh.sshLocal(SshJob{
			Mode: TaskId,
		})
	}

	return applyCmd
}

func (m ModJobSshStruct) sshLocal(job SshJob) {
	switch job.Mode {
	case SshModeGenOrGetKey:
		m.genOrGetKey()
	case SshModeSetupCluster:
		m.setupCluster()
	case SshModeSetupThisNode:
		pubkey, err := util.MainNodeConfReader{}.ReadSecretConf(util.SecretConfTypeSshPublic{})
		if err != nil {
			fmt.Println(color.RedString("read pubkey from main node failed: %v", err))
			os.Exit(1)
		}
		m.setupThisNode(pubkey)
	default:
		fmt.Println(color.RedString("unsupported ssh ope mode: '%s'", job.Mode))
		os.Exit(1)
	}
}

// 用正则表达式验证公钥格式
func (m ModJobSshStruct) setupThisNode(pubkey string) {
	pubkey = strings.TrimSpace(pubkey)
	innerSetPubkeyOnThisNode := func(pubkey string) error {
		// 校验公钥格式
		isValidPubkeyFormat := func(pubkey string) bool {
			// 匹配 ssh-rsa, ssh-ed25519 等 SSH 公钥类型
			re := regexp.MustCompile(`^ssh-(rsa|ed25519|dss|ecdsa) [A-Za-z0-9+/=]+ ?.*$`)
			return re.MatchString(pubkey)
		}
		if !isValidPubkeyFormat(pubkey) {
			return fmt.Errorf("无效的公钥格式 %s", pubkey)
		}

		// 确定目标路径，假设该路径是节点的 authorized_keys 文件路径
		authorizedKeysPath := filepath.Join(homedir.HomeDir(), ".ssh/authorized_keys")
		authorizedKeysDirPath := filepath.Dir(authorizedKeysPath)
		_, err := os.Stat(authorizedKeysDirPath)
		if err != nil {
			os.MkdirAll(authorizedKeysDirPath, 0755)
		}

		// 读取现有的 authorized_keys 文件内容
		existingKeys, err := ioutil.ReadFile(authorizedKeysPath)
		if err != nil {
			return fmt.Errorf("无法读取 authorized_keys 文件: %v", err)
		}

		// 辅助函数：检查公钥是否已存在
		contains := func(keys []byte, pubkey string) bool {
			return string(keys) == pubkey || (string(keys) != "" && string(keys) == pubkey+"\n")
		}

		// 检查是否已经存在该公钥
		if contains(existingKeys, pubkey) {
			return errors.New("该公钥已经存在")
		}

		// 将新的公钥添加到文件内容
		updatedKeys := append(existingKeys, []byte("\n"+pubkey)...)

		// 将更新后的内容写回文件
		err = ioutil.WriteFile(authorizedKeysPath, updatedKeys, 0644)
		if err != nil {
			return fmt.Errorf("无法更新 authorized_keys 文件: %v", err)
		}

		return nil
	}
	err := innerSetPubkeyOnThisNode(pubkey)
	if err != nil {
		fmt.Println(color.RedString("set pubkey failed: %v", err))
	} else {
		fmt.Println(color.GreenString("set pubkey success"))
	}
}

func (m ModJobSshStruct) genOrGetKey() {
	fail := false
	failInfo := ""
	defer func() {
		if fail {
			fmt.Println(color.RedString("fail to gen or get key: %s", failInfo))
		} else {
			fmt.Println(color.GreenString("success to gen or get key"))
		}
	}()

	util.ConfigMainNodeRcloneIfNeed()

	// check local ed25519 exist
	localEd25519Exists := false
	// use homedir lib to get homedir
	homeDir := homedir.HomeDir()
	// ed25519Filepat: concat ed25519 filepath
	ed25519FilePath := filepath.Join(homeDir, ".ssh", "id_ed25519")
	ed25519PubFilePath := filepath.Join(homeDir, ".ssh", "id_ed25519.pub")
	{
		_, err := os.Stat(ed25519FilePath)
		if err == nil {
			localEd25519Exists = true
		}
	}

	// check remote exist public and private key with rclone at remote:/teledeploy_secret/ssh_config/
	type RemoteSshKey struct {
		Pubkey  string
		Privkey string
	}
	var remoteSshKey *RemoteSshKey
	{
		// cmds := util.ModRunCmd.SplitCmdline("rclone ls remote:/teledeploy_secret/ssh_config/")
		// output, err := util.ModRunCmd.NewBuilder(cmds[0], cmds[1:]...).BlockRun()
		pri, err1 := util.MainNodeConfReader{}.ReadSecretConf(util.SecretConfTypeSshPrivate{})
		pub, err2 := util.MainNodeConfReader{}.ReadSecretConf(util.SecretConfTypeSshPublic{})

		if err1 != nil || err2 != nil {
			fmt.Println(color.YellowString("read ssh key failed: %v, %v", err1, err2))
		} else {
			remoteSshKey = &RemoteSshKey{
				Pubkey:  pub,
				Privkey: pri,
			}
		}
	}

	if remoteSshKey != nil {
		// fetch by rclone
		// one is pub key
		util.PrintStep("ssh genOrGetKey", color.BlueString("fetching keys from remote node"))
		pubfile, err := os.Create(ed25519PubFilePath)
		if err != nil {
			fail = true
			failInfo = fmt.Sprintf("failed to create pubfile: %v", err)
			return
		}
		pubfile.WriteString(remoteSshKey.Pubkey)
		pubfile.Close()

		prifile, err := os.Create(ed25519FilePath)
		if err != nil {
			fail = true
			failInfo = fmt.Sprintf("failed to create prifile: %v", err)
			return
		}
		prifile.WriteString(remoteSshKey.Privkey)
		prifile.Close()

		cmdPerm := util.ModRunCmd.SplitCmdline("chmod 600 " + filepath.Join(homeDir, ".ssh", "id_ed25519"))
		_, err = util.ModRunCmd.NewBuilder(cmdPerm[0], cmdPerm[1:]...).WithRoot().BlockRun()
		if err != nil {
			fail = true
			failInfo = fmt.Sprintf("failed to set permission: %v", err)
			return
		}
	} else {
		if localEd25519Exists {
			fmt.Println(color.YellowString("Local ssh key exists, skip generate or fetch"))
			// return
		} else {
			// gen by ssh-keygen
			fmt.Println(color.BlueString("generating new ssh key pair..."))
			homeDir := homedir.HomeDir()
			// sshDir := filepath.Join(homeDir, ".ssh")
			sshFile := filepath.Join(homeDir, ".ssh", "id_ed25519")
			// cmds := []string{"bash", "-c", fmt.Sprintf("ssh-keygen -t ed25519 -f %s -N '' -q", sshFile)}
			cmds := util.ModRunCmd.SplitCmdline(fmt.Sprintf("ssh-keygen -t ed25519 -f %s -N ''", sshFile))
			if _, err := util.ModRunCmd.ShowProgress(cmds[0], cmds[1:]...).BlockRun(); err != nil {
				fail = true
				failInfo = fmt.Sprintf("failed to generate ed25519 keys: %v", err)
				return
			}
			cmdPerm := util.ModRunCmd.SplitCmdline("chmod 600 " + filepath.Join(homeDir, ".ssh", "id_ed25519"))
			util.ModRunCmd.RequireRootRunCmd(cmdPerm[0], cmdPerm[1:]...)
		}

		// upload to remote
		fmt.Println(color.BlueString("uploading ssh keys to remote..."))
		// cmds = strings.Split("rclone copy ~/.ssh/id_ed25519* remote:/teledeploy_secret/ssh_config/", " ")
		// util.UploadToMainNode(ed25519FilePath, "/teledeploy_secret/ssh_config/")
		// util.UploadToMainNode(ed25519PubFilePath, "/teledeploy_secret/ssh_config/")
		localpri, err := os.ReadFile(ed25519FilePath)
		if err != nil {
			fail = true
			failInfo = fmt.Sprintf("failed to read local pri key: %v", err)
			return
		}
		util.MainNodeConfWriter{}.WriteSecretConf(util.SecretConfTypeSshPrivate{}, string(localpri))

		_, err = os.ReadFile(ed25519PubFilePath)
		if err != nil {
			fail = true
			failInfo = fmt.Sprintf("failed to read local pub key: %v", err)
			return
		}

		util.MainNodeConfWriter{}.WriteSecretConf(util.SecretConfTypeSshPrivate{}, string(localpri))
	}
}


// https://qcnoe3hd7k5c.feishu.cn/wiki/V6eHwZm1aiofeykaSd5cmgPcnSe#share-Hc1hdGT26oI4I0xPaplcEhMundd
func (m ModJobSshStruct) setupCluster() {
	ok, yamlFilePath := util.StartTemporaryInputUI(color.GreenString(
		"初始集群配置需要初始集群配置文件 cluster_config.yml"),
		"此处键入 yaml 配置路径",
		"回车确认，ctrl+c取消，参照https://github.com/340Lab/serverless_benchmark_plus/blob/main/middlewares/cluster_config.yml")
	if !ok {
		fmt.Println("User canceled config cluster")
		os.Exit(1)
	}
	// load yaml
	// 读取 YAML 文件
	data, err := ioutil.ReadFile(yamlFilePath)
	if err != nil {
		fmt.Println(color.RedString("读取配置文件失败: %v", err))
		os.Exit(1)
	}

	// 解析 YAML
	var clusterConf clusterconf.ClusterConfYmlModel
	err = yamlext.UnmarshalAndValidate(data, &clusterConf)
	if err != nil {
		fmt.Println(color.RedString("解析 YAML 文件失败: %v", err))
		os.Exit(1)
	}

	// 打印解析后的内容
	fmt.Printf("解析后的集群配置: %+v\n", clusterConf)

	hosts := funk.Map(clusterConf.Nodes, func(nodename string, node clusterconf.ClusterConfYmlModelNode) string {
		return fmt.Sprintf("%s@%s", clusterConf.Global.SshUser, node.Ip)
	}).([]string)

	// read pubkey
	pubkeyFile := filepath.Join(homedir.HomeDir(), ".ssh", "id_ed25519.pub")
	pubkeybytes, err := os.ReadFile(pubkeyFile)
	if err != nil {
		fmt.Println(color.RedString("read pubkey failed: %v", err))
		os.Exit(1)
	}
	_ = base64.StdEncoding.EncodeToString(pubkeybytes)

	util.StartRemoteCmds(
		hosts,
		// install telego,
		util.ModRunCmd.CmdModels().InstallTelegoWithPy()+" && "+
			// update authorized_keys
			strings.Join(m.NewSshCmd(SshJob{Mode: SshModeSetupThisNode}.ModeString()), " "),
		clusterConf.Global.SshPasswd,
	)
}

func (m ModJobSshStruct) NewSshCmd(
	sshModeStr string,
) []string {
	switch sshModeStr {
	case SshJob{Mode: SshModeGenOrGetKey}.ModeString(),
		SshJob{Mode: SshModeSetupCluster}.ModeString(),
		SshJob{Mode: SshModeSetupThisNode}.ModeString():
		return []string{"telego", "ssh",
			"--mode", sshModeStr}
	default:
		fmt.Println(color.RedString("unsupported ssh ope mode: '%s'", sshModeStr))
		os.Exit(1)
	}
	return []string{}
}

func (m ModJobSshStruct) ExecDelegate(sshModeStr string) DispatchExecRes {
	cmd := ModJobSsh.NewSshCmd(sshModeStr)
	return DispatchExecRes{
		Exit: true,
		ExitWithDelegate: func() {
			fmt.Println(color.BlueString("Starting job ssh... cmds[%v]", cmd))
			_, err := util.ModRunCmd.ShowProgress(cmd[0], cmd[1:]...).BlockRun()
			if err != nil {
				fmt.Println(color.RedString("job ssh error: %v", err))
			}
			fmt.Println(color.GreenString("job ssh finished"))
		},
	}
}

// // in app entry
// func (_ ModJobSshStruct) SshLocal(k8sprj string, k8sdp *Deployment, clusterContextName string) {
// 	// cmds := []string{}
// 	cmds := ModJobSsh.NewSshCmd(k8sprj, k8sdp.K8s, k8sdp.Helms, clusterContextName)
// 	util.ModRunCmd.RunCommandShowProgress(cmds[0], cmds[1:]...)
// 	// cmd += NewSshCmd(binpack, bin, binBin[bin])

// 	// util.Logger.Debugf("apply cmds split: %s", cmds)

// }
