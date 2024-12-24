package app

import (
	"fmt"
	"os"
	"telego/util"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ModJobDistributeDeployStruct struct {
}

var ModJobDistributeDeploy ModJobDistributeDeployStruct

func (m ModJobDistributeDeployStruct) JobCmdName() string {
	return "distribute-deploy"
}

type DistributeDeployJob struct {
	Deployer DistributeDeployer
	Mode     int
}

const (
	DistDeployModeDeployerAll = iota
	DistDeployModeThisNodeWorker
	DistDeployModeThisNodeMaster
)

func (s DistributeDeployJob) ModeString() string {
	switch s.Mode {
	case DistDeployModeDeployerAll:
		return "distribute_deploy_all"
	case DistDeployModeThisNodeWorker:
		return "distribute_deploy_this_worker"
	case DistDeployModeThisNodeMaster:
		return "distribute_deploy_this_master"
	default:
		return "unknown"
	}
}

func (j DistributeDeployJob) LoadDeployer(deployerName string) (DistributeDeployer, error) {
	switch deployerName {
	case "k3s":
		return DistributeDeployerK3s{}, nil
	default:
		return nil, fmt.Errorf("unsupported deployer name: %s", deployerName)
	}
}

func (m ModJobDistributeDeployStruct) ParseJob(applyCmd *cobra.Command) *cobra.Command {

	// 绑定命令行标志到结构体字段
	mode := ""
	deployerName := ""
	installCtxBase64 := ""

	applyCmd.Flags().StringVar(&mode, "mode", "", "Sub operation distribute deployer")
	applyCmd.Flags().StringVar(&deployerName, "deployer", "", "Deployer name")
	applyCmd.Flags().StringVar(&installCtxBase64, "install-this-ctx", "", "Install worker context encoded in base64")

	applyCmd.Run = func(_ *cobra.Command, _ []string) {
		deployer := NewDistributeDeployer(deployerName)
		if deployer == nil {
			fmt.Println(color.RedString("unsupported deployer: '%s'", deployerName))
			return
		}

		switch mode {
		case DistributeDeployJob{Mode: DistDeployModeDeployerAll}.ModeString():
			ModDistributeDeploy.SetupAll(deployer)
		case DistributeDeployJob{Mode: DistDeployModeThisNodeMaster}.ModeString():
			fmt.Println(color.BlueString("installing general part for %s", deployer.Name()))
			err := deployer.ThisNodeGeneralInstall()
			if err != nil {
				fmt.Println(color.RedString("general part install failed: %s", err))
				os.Exit(1)
			}
			fmt.Println(color.BlueString("\ninstalling master part for %s", deployer.Name()))
			err = deployer.ThisNodeMasterInstall(installCtxBase64)
			if err != nil {
				fmt.Println(color.RedString("master part install failed: %s", err))
				os.Exit(1)
			}
		case DistributeDeployJob{Mode: DistDeployModeThisNodeWorker}.ModeString():
			fmt.Println(color.BlueString("installing general part for %s", deployer.Name()))
			err := deployer.ThisNodeGeneralInstall()
			if err != nil {
				fmt.Println(color.RedString("general part install failed: %s", err))
				os.Exit(1)
			}
			fmt.Println(color.BlueString("\ninstalling worker part for %s", deployer.Name()))
			err = deployer.ThisNodeWorkerInstall(installCtxBase64)
			if err != nil {
				fmt.Println(color.RedString("worker part install failed: %s", err))
			}
		default:
			fmt.Println(color.RedString("unsupported ssh ope mode: '%s'", mode))
			os.Exit(1)
		}
	}

	return applyCmd
}

func (m ModJobDistributeDeployStruct) NewCmd(job DistributeDeployJob, installThisCtxBase64 string) []string {
	ret := []string{"telego", m.JobCmdName(),
		"--deployer", job.Deployer.Name(),
		"--mode", job.ModeString(),
	}
	if installThisCtxBase64 != "" {
		ret = append(ret, "--install-this-ctx", installThisCtxBase64)
	}
	return ret
}

func (m ModJobDistributeDeployStruct) ExecDelegate(job DistributeDeployJob) DispatchExecRes {
	cmd := m.NewCmd(job, "")
	return DispatchExecRes{
		Exit: true,
		ExitWithDelegate: func() {
			util.PrintStep("job_distribute_deploy", "start deploy "+job.Deployer.Name())
			_, err := util.ModRunCmd.ShowProgress(cmd[0], cmd[1:]...).BlockRun()
			if err != nil {
				fmt.Println(color.RedString("job_distribute_deploy error: %v", err))
			}
			fmt.Println(color.GreenString("job_distribute_deploy finished"))
		},
	}
}
