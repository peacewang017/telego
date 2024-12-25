package app

import (
	"fmt"
	"os"
	"telego/util"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type ModJobRcloneStruct struct{}

var ModJobRclone ModJobRcloneStruct

// Usage:
// telego rclone --mount {localPath} {remoteUser@remoteHostIP:remotePath}

type rcloneJob struct {
	jobType   string // mount / sync
	mountArgv *rcloneMountArgv
	// syncArgv rcloneSyncArgv
}

type rcloneMountArgv struct {
	remotePath string
	localPath  string
}

func (m ModJobRcloneStruct) JobCmdName() string {
	return "rclone"
}

func (m ModJobRcloneStruct) ParseJob(rcloneCmd *cobra.Command) *cobra.Command {
	job := &rcloneJob{
		mountArgv: &rcloneMountArgv{},
	}

	// 读入参数
	rcloneCmd.Flags().StringVar(&job.jobType, "mode", "", "rclone - sub operation") // --mode mount/sync
	rcloneCmd.Flags().StringVar(&job.mountArgv.remotePath, "remotepath", "", "rclone mount - remote path")
	rcloneCmd.Flags().StringVar(&job.mountArgv.localPath, "localpath", "", "rclone mount - local mountPath")

	rcloneCmd.Run = func(_ *cobra.Command, _ []string) {
		switch job.jobType {
		case "mount":
			m.doMount(job.mountArgv)
		default:
			fmt.Println(color.RedString("unsupported rclone sub operation"))
			os.Exit(1)
		}
	}

	return rcloneCmd
}

func (m ModJobRcloneStruct) doMount(argv *rcloneMountArgv) error {
	fmt.Printf("%s, %s", argv.remotePath, argv.localPath)
	if argv.localPath == "" || argv.remotePath == "" {
		return fmt.Errorf("rclone mount argument empty")
	}
	if _, err := os.Stat(argv.localPath); os.IsNotExist(err) {
		return fmt.Errorf("local path %s does not exist", argv.localPath)
	}
	_, err := util.RunCmdWithTimeoutCheck(
		fmt.Sprintf("sudo rclone --debug mount %s %s", argv.remotePath, argv.localPath),
		1*time.Second,
		func(output string) bool {
			return len(output) < 50
		},
	)
	return err
}
