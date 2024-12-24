package app

import (
	"fmt"
	"os"
	"path"

	"telego/util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

type InstallJob struct {
	BinPack string
	Bin     string // left empty to install all
	// BinMeta DeploymentBinDetails
}

type ModJobInstallStruct struct{}

var ModJobInstall ModJobInstallStruct

func (m ModJobInstallStruct) JobCmdName() string {
	return "install"
}

func (_ ModJobInstallStruct) ParseJob(installCmd *cobra.Command) *cobra.Command {
	job := &InstallJob{}
	// 绑定命令行标志到结构体字段
	installCmd.Flags().StringVar(&job.BinPack, "bin-pack", "", "Path to install")
	installCmd.Flags().StringVar(&job.Bin, "bin", "", "Path to binary")
	// bool job.BinMeta.NoDefaultInstaller
	// installCmd.Flags().BoolVar(&job.BinMeta.NoDefaultInstaller, "no-default-installer", false, "No default installer")
	// installCmd.Flags().StringVar(&job.BinMeta.WinInstaller, "win-installer", "", "Windows installer")
	// installCmd.Flags().StringVar(&job.BinMeta.Appimage, "appimage", "", "Appimage")

	installCmd.Run = func(_ *cobra.Command, _ []string) {
		fmt.Println(color.BlueString("Install job running %s %s", job.BinPack, job.Bin))
		// if job.Bin == "" {
		// 	fmt.Println(color.RedString("No bin provided"))
		// 	os.Exit(1)
		// }
		if job.BinPack == "" {
			fmt.Println(color.RedString("No bin provided"))
			os.Exit(1)
		}
		ModJobInstall.InstallLocalByJob(*job)
	}

	// err := installCmd.Execute()
	// if err != nil {
	// 	return
	// }

	return installCmd

	// return JobParse{
	// 	Cmd: installCmd,
	// 	Cb: func() {

	// 		os.Exit(0)
	// 	},
	// }
}

func (_ ModJobInstallStruct) getBinDeploymentFromMainNode(job InstallJob) (*Deployment, error) {
	url := fmt.Sprintf("http://%s:8003/%s/deployment.yml", util.MainNodeIp, job.BinPack)
	ymlData, err := util.HttpGetUrlContent(url)
	if err != nil {
		util.Logger.Warnf("getBinDeploymentFromMainNode Failed to fetch file from %s: %s", url, err)
		return nil, err
	}
	return LoadDeploymentYmlByContent("", ymlData)
}

func (_ ModJobInstallStruct) InstallLocalByJob(job InstallJob) {
	// fetch meta
	fmt.Println(color.BlueString("Fetching %s meta", job.BinPack))
	dplymnt, err := ModJobInstall.getBinDeploymentFromMainNode(job)
	if err != nil {
		fmt.Println(color.RedString("Failed to fetch meta: %s", err.Error()))
		os.Exit(1)
	}

	bins := map[string]DeploymentBinDetails{}
	if job.Bin == "" {
		for binname, bin := range dplymnt.Bin {
			// bins = append(bins, bin)
			bins[binname] = bin
		}
	} else {
		for binname, bin := range dplymnt.Bin {
			if binname == job.Bin {
				// bininfo = &bin
				// bins = append(bins, bin)
				bins[binname] = bin
				break
			}
		}
	}

	if len(bins) == 0 {
		fmt.Println(color.RedString("Failed to find bin %s in meta", job.Bin))
		return
	}

	// install
	fmt.Println(color.BlueString("Installing %s / %s", job.BinPack, bins))
	os.Chdir(util.WorkspaceDir())

	for binname, bininfo := range bins {
		installTempDir := path.Join("install_"+job.BinPack, binname)
		os.MkdirAll(installTempDir, 0755)
		defer os.RemoveAll(installTempDir)

		if !bininfo.NoDefaultInstaller {
			if util.IsWindows() {
				err := util.InstallWindowsBin(
					fmt.Sprintf("http://%s:8003/%s/%s.exe", util.MainNodeIp, job.BinPack, binname),
					installTempDir,
					binname,
				)
				if err != nil {
					fmt.Println(color.RedString("Failed to install %s: %s", binname, err.Error()))
					return
				}
			} else {
				arch := util.GetCurrentArch()
				err := util.InstallLinuxBin(
					fmt.Sprintf("http://%s:8003/%s/%s_%s", util.MainNodeIp, job.BinPack, binname, arch),
					installTempDir,
					binname,
				)
				if err != nil {
					fmt.Println(color.RedString("Failed to install %s: %s", binname, err.Error()))
					return
				}
			}
		} else if util.IsWindows() {
			if bininfo.WinInstaller != "" {
				// download win_installer
				util.DownloadFile(
					fmt.Sprintf("http://%s:8003/%s/%s", util.MainNodeIp, job.BinPack, bininfo.WinInstaller),
					path.Join(installTempDir, bininfo.WinInstaller),
				)
				// start installer
				_, err := util.ModRunCmd.ShowProgress(path.Join(installTempDir, bininfo.WinInstaller)).BlockRun()
				if err != nil {
					fmt.Println(color.RedString("Failed to install %s: %s", binname, err.Error()))
					return
				}
			} else {
				fmt.Println(color.RedString("At least one of 'non no_default_installer' or 'win_installer' should be provided for windows"))
				return
			}
		} else if bininfo.Appimage != "" {

		} else {
			fmt.Println(color.RedString("At least one of 'non no_default_installer' or 'appimage' should be provided for linux"))
			return
		}

		fmt.Println(color.GreenString("Installed %s / %s", job.BinPack, binname))
	}

}

func NewInstallCmd(binPack, bin string) []string {
	cmds := []string{"telego", "install",
		"--bin", bin,
		"--bin-pack", binPack}
	// "--no-default-installer %v "+
	// "--win-installer", meta.WinInstaller,
	// "--appimage", meta.Appimage}

	// if meta.NoDefaultInstaller {
	// 	// cmd += "--no-default-installer "
	// 	cmds = append(cmds, "--no-default-installer")
	// }
	return cmds
}

// in app entry
func (_ ModJobInstallStruct) InstallLocal(binpack string) error {
	fmt.Println(color.BlueString("install %s local", binpack))
	cmds := NewInstallCmd(binpack, "")
	_, err := util.ModRunCmd.ShowProgress(cmds[0], cmds[1:]...).BlockRun()
	return err
}

// in app entry
func (_ ModJobInstallStruct) InstallToNodes(binpack string, cluster string, nodes []string) {
	fmt.Println(color.BlueString("install %s to remote", binpack))

	user, err := util.KubeSecretSshUser(cluster)
	if err != nil {
		util.Logger.Warn("failed to get ssh user from secret: " + err.Error())
		fmt.Println(color.RedString("failed to get ssh user from secret: " + err.Error()))
		os.Exit(1)
	}
	name2Ip, err := util.KubeNodeName2Ip(cluster)
	if err != nil {
		util.Logger.Warn("failed to get node ip: " + err.Error())
		fmt.Println(color.RedString("failed to get node ip: " + err.Error()))
		os.Exit(1)
	}
	hosts := funk.Map(nodes, func(node string) string {
		return user + "@" + name2Ip[node]
	}).([]string)

	cmd := fmt.Sprintf("python3 -c \"import urllib.request, os; script = urllib.request.urlopen('http://%s:8003/bin_telego/install.py').read(); exec(script.decode());\" ", util.MainNodeIp)
	cmd += "&& " + CmdsToCmd(NewInstallCmd(binpack, ""))

	util.Logger.Debugf("install cmd: %s", cmd)
	util.StartRemoteCmds(hosts, cmd, "")
}
